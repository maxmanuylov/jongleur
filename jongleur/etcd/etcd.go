package etcd

import (
    _etcd "github.com/coreos/etcd/client"
    "github.com/coreos/etcd/discovery"
    "github.com/coreos/etcd/etcdserver"
    "github.com/coreos/etcd/pkg/types"
    "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/utils"
    "github.com/maxmanuylov/jongleur/utils/etcd"
    "net/http"
    "net/url"
)

type Config struct {
    Verbose   bool
    Listen    string
    Period    int
    Discovery string
}

func (config *Config) ToJongleurConfig() (*jongleur.Config, error) {
    if err := utils.Check(config); err != nil {
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

    itemsLoader, err := etcd_utils.NewEtcdItemsLoader(config.Period, etcdCluster.ClientURLs(), func (etcdClient _etcd.Client) ([]string, error) {
        if err := etcdClient.Sync(context.Background()); err != nil {
            return nil, err
        }

        newItems := make([]string, 0)

        for _, endpoint := range etcdClient.Endpoints() {
            endpointUrl, err := url.Parse(endpoint)
            if err == nil {
                newItems = append(newItems, endpointUrl.Host)
            }
        }

        return newItems, nil
    })

    return &jongleur.Config{
        Verbose: config.Verbose,
        Listen: config.Listen,
        Period: config.Period,
        ItemsLoader: itemsLoader,
        RequestPatcher: jongleur.IDENTICAL_PATCHER,
        ResponsePatcher: jongleur.IDENTICAL_PATCHER,
    }, err
}
