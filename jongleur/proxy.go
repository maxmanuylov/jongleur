package jongleur

import (
    "errors"
    "io"
    "net"
    "time"
)

const serviceUnavailable = "HTTP/1.1 503 Service unavailable\n"

func runProxy(tcpListener *net.TCPListener, data *runtimeData) {
    var n int64 = 0
    for {
        tcpConnection, err := tcpListener.AcceptTCP()
        if err != nil {
            data.logger.Printf("TCP server: %s\n", err.Error())
            return
        }
        n++
        go handleConnection(tcpConnection, data, n)
    }
}

func handleConnection(clientConnection *net.TCPConn, data *runtimeData, n int64) {
    defer clientConnection.Close()

    if data.verbose {
        data.logger.Printf("[%d] Accepted connection from %+v\n", n, clientConnection.RemoteAddr())
    }

    for i := 0; i < 10; i++ {
        host, err := nextHost(data)
        if err != nil {
            clientConnection.Write([]byte(err.Error()))
            if data.verbose {
                data.logger.Printf("[%d] Failed to get the next endpoint: %s\n", n, err.Error())
            }
            return
        }

        if data.verbose {
            data.logger.Printf("[%d] Attempt #%d. Endpoint: %s\n", n, i + 1, host)
        }

        serviceConnection, err := net.DialTimeout("tcp", host, 2 * time.Second)
        if err != nil {
            data.logger.Printf("[%d] Connection to endpoint \"%s\" failed: %s\n", n, host, err.Error())
            continue
        }

        if data.verbose {
            data.logger.Println("[%d] Connected successfully, transferring data...", n)
        }

        link(clientConnection, serviceConnection)

        if data.verbose {
            data.logger.Println("[%d] Data is successfully transferred", n)
        }

        return
    }

    if data.verbose {
        data.logger.Println("[%d] All connection attempts failed", n)
    }

    clientConnection.Write([]byte(serviceUnavailable))
}

func nextHost(data *runtimeData) (string, error) {
    select {
    case host := <-data.hosts:
        return host, nil
    case <-time.After(time.Second):
        return "", errors.New(serviceUnavailable)
    }
}

func link(clientConnection *net.TCPConn, serviceConnection net.Conn) {
    defer serviceConnection.Close()

    done := make(chan bool, 2)

    go copyStream(clientConnection, serviceConnection, done)
    go copyStream(serviceConnection, clientConnection, done)

    <-done
    <-done
}

func copyStream(from io.Reader, to io.Writer, done chan<- bool) {
    defer func() {
        done <- true
    }()

    io.Copy(to, from)
}
