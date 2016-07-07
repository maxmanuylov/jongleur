package jongleur

import (
    "errors"
    "io"
    "net"
    "time"
)

const serviceUnavailable = "HTTP/1.1 503 Service unavailable\n"

func runProxy(tcpListener *net.TCPListener, data *runtimeData) {
    for {
        tcpConnection, err := tcpListener.AcceptTCP()
        if err != nil {
            data.logger.Printf("TCP server: %s\n", err.Error())
            return
        }
        go handleConnection(tcpConnection, data)
    }
}

func handleConnection(clientConnection *net.TCPConn, data *runtimeData) {
    defer clientConnection.Close()

    for i := 0; i < 10; i++ {
        host, err := nextHost(data)
        if err != nil {
            clientConnection.Write([]byte(err.Error()))
            return
        }

        serviceConnection, err := net.DialTimeout("tcp", host, 2 * time.Second)
        if err != nil {
            data.logger.Printf("Connection to host \"%s\" failed: %s\n", host, err.Error())
            continue
        }

        link(clientConnection, serviceConnection)

        return
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
