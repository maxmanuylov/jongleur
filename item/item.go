package item

import (
    "errors"
    "fmt"
    etcd "github.com/coreos/etcd/client"
    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/maxmanuylov/jongleur/utils"
    "github.com/maxmanuylov/utils/application"
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type StringHolder struct {
    Value string
}

type Config struct {
    Type      string
    Host      string
    Health    *StringHolder // Health check can be disabled
    Period    int
    Tolerance int
    Etcd      string
}

func Run(config *Config, logger *log.Logger) error {
    if err := utils.Check(config); err != nil {
        return err
    }

    data, err := config.createRuntimeData(logger)
    if err != nil {
        return err
    }

    ticker := time.NewTicker(data.period)
    defer ticker.Stop()

    go func() {
        for range ticker.C {
            checkAndRefreshItem(data)
        }
    }()

    application.WaitForTermination()

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

    if port != "*" {
        portNum, err := strconv.Atoi(port)
        if err != nil {
            return nil, err
        }

        if portNum < 0 || portNum > 0xFFFF {
            return nil, errors.New("Invalid port value")
        }
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
        period: periodDuration,
        logger: logger,
        httpClient: &http.Client{
            Timeout: semiPeriodDuration,
        },
        healthUrl: config.Health.Value,
        etcdClient: etcdClient,
        etcdKey: fmt.Sprintf("%s/%s", utils.EtcdItemsKey(config.Type), config.Host),
        ttl: periodDuration * time.Duration(config.Tolerance) + semiPeriodDuration,
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
    if data.healthUrl == "" {
        return true, nil
    }

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