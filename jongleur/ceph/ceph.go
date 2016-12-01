package ceph

import (
    "fmt"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/jongleur/regular"
    "github.com/maxmanuylov/jongleur/utils"
    "io"
    "net"
    "strings"
)

type Config struct {
    regular.Config
}

func (config *Config) ToJongleurConfig() (*jongleur.Config, error) {
    jongleurConfig, err := config.Config.ToJongleurConfig()
    if err != nil {
        return nil, err
    }

    network, addr := jongleurConfig.SplitNetAddr()
    if !strings.HasPrefix(network, "tcp") {
        return nil, fmt.Errorf("TCP address is required for Ceph mode: %s", jongleurConfig.Listen)
    }

    ipStr, portStr, err := net.SplitHostPort(addr)
    if err != nil {
        return nil, fmt.Errorf("Failed to parse TCP address \"%s\": %v", addr, err)
    }

    ip := net.ParseIP(ipStr)
    if ip == nil {
        return nil, fmt.Errorf("Failed to parse monitor IP: %s", ipStr)
    }

    port, err := utils.ParsePort(portStr)
    if err != nil {
        return nil, fmt.Errorf("Failed to parse TCP port \"%s\": %v", portStr, err)
    }

    jongleurConfig.ResponsePatcher = func(originalWriter io.Writer) io.Writer {
        return &bytesPatcher{
            originalWriter: originalWriter,
            newBytes: append([]byte{byte(port / 256), byte(port % 256)}, ip.To4()...),
            skip: 19,
        }
    }

    return jongleurConfig, nil
}

type bytesPatcher struct {
    originalWriter io.Writer
    newBytes       []byte
    skip           int
}

func (bp *bytesPatcher) Write(originalBytes []byte) (int, error) {
    if len(originalBytes) == 0 {
        return 0, nil
    }

    if len(bp.newBytes) == 0 {
        return bp.originalWriter.Write(originalBytes)
    }

    var (
        n, nTotal int
        err error
    )

    if bp.skip > 0 {
        var originalBytesToWrite []byte
        if len(originalBytes) < bp.skip {
            originalBytesToWrite = originalBytes
        } else {
            originalBytesToWrite = originalBytes[:bp.skip]
        }

        nTotal, err = bp.originalWriter.Write(originalBytesToWrite)
        bp.skip -= nTotal
        if err != nil || nTotal == len(originalBytes) || bp.skip > 0 {
            return nTotal, err
        }
    }

    newBytesToWrite := bp.newBytes[:min(len(bp.newBytes), len(originalBytes) - nTotal)]

    n, err = bp.originalWriter.Write(newBytesToWrite)
    bp.newBytes = bp.newBytes[n:]
    nTotal += n
    if err != nil || nTotal == len(originalBytes) || len(bp.newBytes) > 0 {
        return nTotal, err
    }

    n, err = bp.originalWriter.Write(originalBytes[nTotal:])
    return n + nTotal, err
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
