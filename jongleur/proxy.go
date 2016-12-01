package jongleur

import (
    "errors"
    "io"
    "net"
    "time"
)

type WriteCloseableConn interface {
    net.Conn
    CloseWrite() error
}

var _ WriteCloseableConn = (*net.TCPConn)(nil)
var _ WriteCloseableConn = (*net.UnixConn)(nil)

const serviceUnavailable = "HTTP/1.1 503 Service unavailable\n"

func runProxy(listener net.Listener, data *runtimeData) {
    var n int64 = 0
    for {
        connection, err := listener.Accept()
        if err != nil {
            data.logger.Printf("[Server] %s\n", err.Error())
            return
        }
        n++
        go handleConnection(connection, data, n)
    }
}

func handleConnection(clientConnection net.Conn, data *runtimeData, n int64) {
    defer clientConnection.Close()

    if conn, ok := clientConnection.(WriteCloseableConn); ok {
        doHandleConnection(conn, data, n)
        return
    }

    data.logger.Printf("Unsupported connection: %+v\n", clientConnection)
}

func doHandleConnection(clientConnection WriteCloseableConn, data *runtimeData, n int64) {
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

        serviceConnection, err := dialTCP(host, 2 * time.Second)
        if err != nil {
            data.logger.Printf("[%d] Connection to endpoint \"%s\" failed: %s\n", n, host, err.Error())
            continue
        }

        if data.verbose {
            data.logger.Printf("[%d] Connected successfully, transferring data...\n", n)
        }

        link(clientConnection, serviceConnection, data)

        if data.verbose {
            data.logger.Printf("[%d] Data is successfully transferred\n", n)
        }

        return
    }

    if data.verbose {
        data.logger.Printf("[%d] All connection attempts failed\n", n)
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

func dialTCP(host string, timeout time.Duration) (*net.TCPConn, error) {
    conn, err := net.DialTimeout("tcp", host, timeout)
    if err != nil {
        return nil, err
    }

    tcpConn, ok := conn.(*net.TCPConn)
    if !ok {
        defer conn.Close()
        return nil, errors.New("Not a TCP connection")
    }

    return tcpConn, nil
}

func link(clientConnection WriteCloseableConn, serviceConnection WriteCloseableConn, data *runtimeData) {
    defer serviceConnection.Close()

    done := make(chan bool, 2)

    go copyStream(clientConnection, serviceConnection, data.requestPatcher, done)
    go copyStream(serviceConnection, clientConnection, data.responsePatcher, done)

    <-done
    <-done
}

func copyStream(from io.Reader, to WriteCloseableConn, patcher Patcher, done chan<- bool) {
    defer func() {
        done <- true
    }()

    defer to.CloseWrite()

    io.Copy(patcher(to), from)
}
