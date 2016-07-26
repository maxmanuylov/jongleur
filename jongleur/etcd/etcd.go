package etcd

import (
    _etcd "github.com/coreos/etcd/client"
    "github.com/coreos/etcd/discovery"
    "github.com/coreos/etcd/etcdserver"
    "github.com/coreos/etcd/pkg/types"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/util"
    "golang.org/x/net/context"
    "net/http"
    "net/url"
)

type Config struct {
    Local     bool
    Port      int
    Period    int
    Discovery string
}

func (config *Config) ToJongleurConfig() (*jongleur.Config, error) {
    if err := util.Check(config); err != nil {
        return nil, err
    }

    etcdPeers, err := discovery.GetCluster(config.Discovery, "")
    if err != nil {
        return nil, err
    }

    etcdPeerUrlsMap, err := types.NewURLsMap(etcdPeers)
    if err != nil {
        return nil, err
    }

    etcdCluster, err := etcdserver.GetClusterFromRemotePeers(etcdPeerUrlsMap.URLs(), http.DefaultTransport /*todo*/)
    if err != nil {
        return nil, err
    }

    return &jongleur.Config{
        Local: config.Local,
        Port: config.Port,
        Period: config.Period,
        Etcd: etcdCluster.ClientURLs(),
        ItemsLoader: func (etcdClient _etcd.Client) ([]string, error) {
            if err := etcdClient.Sync(context.Background()); err != nil {
                return nil, err
            }

            newItems := make([]string, 0)

            for _, endpoint := range etcdClient.Endpoints() {
                url, err := url.Parse(endpoint)
                if err == nil {
                    newItems = append(newItems, url.Host)
                }
            }

            return newItems, nil
        },
    }, nil
}
