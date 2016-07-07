package jongleur

import (
    "errors"
    etcd "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/util"
    "github.com/maxmanuylov/jongleur/util/cycle"
    "golang.org/x/net/context"
    "log"
    "net"
    "strings"
    "time"
)

type Config struct {
    Items  string
    Port   int
    Period int
    Etcd   string
}

func Run(config *Config, logger *log.Logger) error {
    if err := util.Check(config); err != nil {
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

    tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{Port:config.Port})
    if err != nil {
        return err
    }

    defer tcpListener.Close()

    go runProxy(tcpListener, data)
    data.logger.Printf("Listening for TCP connections on port %v\n", config.Port)

    util.WaitForTermination()

    return nil
}

type runtimeData struct {
    period     time.Duration
    logger     *log.Logger
    etcdClient etcd.Client
    etcdKey    string
    mcycle     *cycle.MutableCycle
    hosts      <-chan string
}

func (config *Config) createRuntimeData(logger *log.Logger) (*runtimeData, error) {
    if strings.Contains(config.Items, "/") {
        return nil, errors.New("Invalid symbol in items: '/'")
    }

    if config.Port < 0 || config.Port > 0xFFFF {
        return nil, errors.New("Invalid port value")
    }

    if config.Period <= 0 {
        return nil, errors.New("Period must be positive")
    }

    periodDuration := time.Duration(config.Period) * time.Second
    semiPeriodDuration := periodDuration / 2

    etcdClient, err := etcd.New(etcd.Config{
        Endpoints:               []string{config.Etcd},
        Transport:               etcd.DefaultTransport,
        HeaderTimeoutPerRequest: semiPeriodDuration,
    })
    if err != nil {
        return nil, err
    }

    hosts := make(chan string)

    return &runtimeData{
        periodDuration,
        logger,
        etcdClient,
        util.EtcdItemsKey(config.Items),
        cycle.NewMutableCycle(hosts),
        hosts,
    }, nil
}

func syncItems(data *runtimeData) {
    keys := etcd.NewKeysAPI(data.etcdClient)

    response, err := keys.Get(context.Background(), data.etcdKey, nil)
    if err != nil {
        data.logger.Printf("Failed to get items from etcd: %s\n", err.Error())
        return
    }

    if response.Node == nil {
        return
    }

    newItems := make([]string, 0)

    if response.Node.Nodes != nil {
        for _, node := range response.Node.Nodes {
            if !node.Dir {
                newItems = append(newItems, simpleKey(node.Key))
            }
        }
    }

    data.mcycle.SyncItems(newItems)
}

func simpleKey(key string) string {
    pos := strings.LastIndex(key, "/")
    if pos == -1 {
        return key
    } else {
        return key[pos + 1:]
    }
}
