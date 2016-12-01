package regular

import (
    "errors"
    etcd "github.com/coreos/etcd/client"
    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/utils"
    "strconv"
    "strings"
)

type Config struct {
    Verbose    bool
    Items      string
    Listen     string
    RemotePort int
    Period     int
    Etcd       string
}

func (config *Config) ToJongleurConfig() (*jongleur.Config, error) {
    if err := utils.Check(config); err != nil {
        return nil, err
    }

    if strings.Contains(config.Items, "/") {
        return nil, errors.New("Invalid symbol in items: '/'")
    }

    etcdKey := utils.EtcdItemsKey(config.Items)
    remotePortStr := config.getRemotePortStr()

    return &jongleur.Config{
        Verbose: config.Verbose,
        Listen: config.Listen,
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
                        item := simpleKey(node.Key)

                        if remotePortStr != "" {
                            item = strings.Replace(item, "*", remotePortStr, -1)
                        }

                        if !strings.Contains(item, "*") {
                            newItems = append(newItems, item)
                        }
                    }
                }
            }

            return newItems, nil
        },
        RequestPatcher: jongleur.IDENTICAL_PATCHER,
        ResponsePatcher: jongleur.IDENTICAL_PATCHER,
    }, nil
}

func (config *Config) getRemotePortStr() string {
    if config.RemotePort == -1 {
        return ""
    } else {
        return strconv.Itoa(config.RemotePort)
    }
}

func simpleKey(key string) string {
    pos := strings.LastIndex(key, "/")
    if pos == -1 {
        return key
    } else {
        return key[pos + 1:]
    }
}
