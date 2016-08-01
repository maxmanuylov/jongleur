package jongleur

import (
    "errors"
    etcd "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/utils"
    "github.com/maxmanuylov/jongleur/utils/cycle"
    "github.com/maxmanuylov/utils/application"
    "log"
    "net"
    "time"
)

type ItemsLoader func (etcdClient etcd.Client) ([]string, error)

type Config struct {
    Local       bool
    Port        int
    Period      int
    Etcd        []string
    ItemsLoader ItemsLoader
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

    tcpListener, err := net.ListenTCP("tcp", config.getTCPAddr())
    if err != nil {
        return err
    }

    defer tcpListener.Close()

    go runProxy(tcpListener, data)
    data.logger.Printf("Listening for TCP connections on port %v\n", config.Port)

    application.WaitForTermination()

    return nil
}

type runtimeData struct {
    period     time.Duration
    logger     *log.Logger
    etcdClient etcd.Client
    loadItems  ItemsLoader
    mcycle     *cycle.MutableCycle
    hosts      <-chan string
}

func (config *Config) createRuntimeData(logger *log.Logger) (*runtimeData, error) {
    if config.Port < 0 || config.Port > 0xFFFF {
        return nil, errors.New("Invalid port value")
    }

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
        mcycle: cycle.NewMutableCycle(hosts),
        hosts: hosts,
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

func (config *Config) getTCPAddr() (*net.TCPAddr) {
    if config.Local {
        return &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: config.Port}
    } else {
        return &net.TCPAddr{Port: config.Port}
    }
}
