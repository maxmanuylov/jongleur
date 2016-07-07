package item

import (
    "errors"
    "fmt"
    etcd "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/util"
    "golang.org/x/net/context"
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type Config struct {
    Type      string
    Host      string
    Health    string
    Period    int
    Tolerance int
    Etcd      string
}

func Run(config *Config, logger *log.Logger) error {
    if err := util.Check(config); err != nil {
        return err
    }

    data, err := config.createRuntimeData(logger)
    if err != nil {
        return err
    }

    ticker := time.NewTicker(data.period)

    go func() {
        for range ticker.C {
            checkAndRefreshItem(data)
        }
    }()

    util.WaitForTermination()

    ticker.Stop()

    return nil
}

type runtimeData struct {
    period     time.Duration
    logger     *log.Logger
    httpClient *http.Client
    healthUrl  string
    etcdClient etcd.Client
    etcdKey    string
    ttl        time.Duration
}

func (config *Config) createRuntimeData(logger *log.Logger) (*runtimeData, error) {
    if strings.Contains(config.Type, "/") {
        return nil, errors.New("Invalid symbol in type: '/'")
    }

    host, port, err := net.SplitHostPort(config.Host)
    if err != nil {
        return nil, err
    }

    if strings.Contains(host, "/") {
        return nil, errors.New("Invalid symbol in host: '/'")
    }

    if _, err = strconv.Atoi(port); err != nil {
        return nil, err
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

    return &runtimeData{
        periodDuration,
        logger,
        &http.Client{
            Timeout: semiPeriodDuration,
        },
        fmt.Sprintf("http://%s/%s", config.Host, strings.TrimPrefix(config.Health, "/")),
        etcdClient,
        fmt.Sprintf("%s/%s", util.EtcdItemsKey(config.Type), config.Host),
        periodDuration * time.Duration(config.Tolerance) + semiPeriodDuration,
    }, nil
}

func checkAndRefreshItem(data *runtimeData) {
    isAlive, err := isItemAlive(data)
    if err != nil {
        data.logger.Printf("Failed to perform health check: %s\n", err.Error())
        return
    }

    if isAlive {
        if err = refreshItem(data); err != nil {
            data.logger.Printf("Failed to refresh the item in etcd: %s\n", err.Error())
        }
    }
}

func isItemAlive(data *runtimeData) (bool, error) {
    request, err := http.NewRequest("HEAD", data.healthUrl, nil)
    if err != nil {
        return false, err
    }

    request.Header.Add("User-Agent", "curl/7.43.0")

    response, err := data.httpClient.Do(request)
    if err != nil {
        return false, err
    }

    data.logger.Printf("Health status: %s\n", response.Status)

    return response.StatusCode / 100 == 2, nil
}

func refreshItem(data *runtimeData) error {
    keys := etcd.NewKeysAPI(data.etcdClient)

    context := context.Background()

    _, err := keys.Set(context, data.etcdKey, "42", &etcd.SetOptions{
        PrevExist: etcd.PrevNoExist,
        TTL: data.ttl,
        Refresh: false,
    })

    if err != nil && !isAlreadyExistsError(err) {
        return err
    }

    _, err = keys.Set(context, data.etcdKey, "", &etcd.SetOptions{
        PrevExist: etcd.PrevExist,
        TTL: data.ttl,
        Refresh: true,
    })

    return err
}

func isAlreadyExistsError(err error) bool {
    etcdErr, ok := err.(etcd.Error)
    return ok && etcdErr.Code == etcd.ErrorCodeNodeExist
}