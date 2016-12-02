package jongleur

import (
    "errors"
    etcd "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/utils"
    "github.com/maxmanuylov/jongleur/utils/cycle"
    "github.com/maxmanuylov/utils/application"
    "io"
    "log"
    "net"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type ItemsLoader func (etcdClient etcd.Client) ([]string, error)

type Patcher func(io.Writer) io.Writer

var IDENTICAL_PATCHER Patcher = func(originalWriter io.Writer) io.Writer {
    return originalWriter
}

type Config struct {
    Verbose         bool
    Listen          string // "[<network>@]<addr>"
    Period          int
    Etcd            []string
    ItemsLoader     ItemsLoader
    RequestPatcher  Patcher
    ResponsePatcher Patcher
}

func Run(config *Config, logger *log.Logger) error {
    if err := utils.Check(config); err != nil {
        return err
    }

    data, err := config.createRuntimeData(logger)
    if err != nil {
        return err
    }

    defer data.mcycle.Stop()

    etcdTicker := time.NewTicker(data.period)
    defer etcdTicker.Stop()

    go func() {
        for range etcdTicker.C {
            syncItems(data)
        }
    }()

    listener, err := config.listen()
    if err != nil {
        return err
    }

    defer listener.Close()

    go runProxy(listener, data)
    data.logger.Printf("Listening for TCP connections on %+v\n", listener.Addr())

    application.WaitForTermination()

    return nil
}

type runtimeData struct {
    period          time.Duration
    logger          *log.Logger
    etcdClient      etcd.Client
    loadItems       ItemsLoader
    mcycle          *cycle.MutableCycle
    hosts           <-chan string
    requestPatcher  Patcher
    responsePatcher Patcher
    verbose         bool
}

func (config *Config) createRuntimeData(logger *log.Logger) (*runtimeData, error) {
    if config.Period <= 0 {
        return nil, errors.New("Period must be positive")
    }

    periodDuration := time.Duration(config.Period) * time.Second
    semiPeriodDuration := periodDuration / 2

    etcdClient, err := etcd.New(etcd.Config{
        Endpoints:               config.Etcd,
        Transport:               etcd.DefaultTransport,
        HeaderTimeoutPerRequest: semiPeriodDuration,
    })
    if err != nil {
        return nil, err
    }

    hosts := make(chan string)

    return &runtimeData{
        period: periodDuration,
        logger: logger,
        etcdClient: etcdClient,
        loadItems: config.ItemsLoader,
        mcycle: cycle.NewMutableCycle(hosts, logger),
        hosts: hosts,
        requestPatcher: config.RequestPatcher,
        responsePatcher: config.ResponsePatcher,
        verbose: config.Verbose,
    }, nil
}

func syncItems(data *runtimeData) {
    newItems, err := data.loadItems(data.etcdClient)
    if err != nil {
        data.logger.Printf("Failed to get items from etcd: %v\n", err)
        return
    }

    if newItems != nil {
        data.mcycle.SyncItems(newItems)
    }
}

func (config *Config) SplitNetAddr() (string, string) {
    atPos := strings.Index(config.Listen, "@")
    if atPos == -1 {
        return "tcp", config.Listen
    } else {
        return config.Listen[:atPos], config.Listen[atPos + 1:]
    }
}

func (config *Config) listen() (net.Listener, error) {
    network, addr := config.SplitNetAddr()

    if strings.HasPrefix(network, "unix") {
        if err := os.MkdirAll(filepath.Dir(addr), 0755); err != nil {
            return nil, err
        }
    }

    return net.Listen(network, addr)
}
