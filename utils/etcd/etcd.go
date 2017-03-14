package etcd_utils

import (
    "fmt"
    etcd_client "github.com/coreos/etcd/client"
    "github.com/maxmanuylov/jongleur/jongleur"
    "time"
)

func NewEtcdItemsLoader(period int, etcdEndpoints []string, loader func (etcd_client.Client) ([]string, error)) (jongleur.ItemsLoader, error) {
    etcdClient, err := etcd_client.New(etcd_client.Config{
        Endpoints:               etcdEndpoints,
        Transport:               etcd_client.DefaultTransport,
        HeaderTimeoutPerRequest: time.Duration(period) * time.Second / 2,
    })

    if err != nil {
        return nil, err
    }

    return func() ([]string, error) {
        return loader(etcdClient)
    }, nil
}

func EtcdItemsKey(itemType string) string {
    return fmt.Sprintf("/jongleur/items/%s", itemType)
}