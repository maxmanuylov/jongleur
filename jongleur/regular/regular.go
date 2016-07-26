package regular

import (
    "errors"
    etcd "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/util"
    "golang.org/x/net/context"
    "strconv"
    "strings"
)

type Config struct {
    Items  string
    Local  bool
    Port   int
    Period int
    Etcd   string
}

func (config *Config) ToJongleurConfig() (*jongleur.Config, error) {
    if err := util.Check(config); err != nil {
        return nil, err
    }

    if strings.Contains(config.Items, "/") {
        return nil, errors.New("Invalid symbol in items: '/'")
    }

    etcdKey := util.EtcdItemsKey(config.Items)
    portStr := strconv.Itoa(config.Port)

    return &jongleur.Config{
        Local: config.Local,
        Port: config.Port,
        Period: config.Period,
        Etcd: []string{config.Etcd},
        ItemsLoader: func (etcdClient etcd.Client) ([]string, error) {
            keys := etcd.NewKeysAPI(etcdClient)

            response, err := keys.Get(context.Background(), etcdKey, nil)
            if err != nil {
                return nil, err
            }

            if response.Node == nil {
                return nil, nil
            }

            newItems := make([]string, 0)

            if response.Node.Nodes != nil {
                for _, node := range response.Node.Nodes {
                    if !node.Dir {
                        newItems = append(newItems, strings.Replace(simpleKey(node.Key), "*", portStr, -1))
                    }
                }
            }

            return newItems, nil
        },
    }, nil
}

func simpleKey(key string) string {
    pos := strings.LastIndex(key, "/")
    if pos == -1 {
        return key
    } else {
        return key[pos + 1:]
    }
}