package item

import (
    "github.com/maxmanuylov/jongleur/util"
)

type Config struct {
    Type      string
    Host      string
    Health    string
    Period    int
    Tolerance int
    Etcd      string
}

func Run(config *Config) error {
    if err := util.Check(config); err != nil {
        return err
    }

    // todo

    return nil
}
