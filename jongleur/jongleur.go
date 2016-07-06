package jongleur

import (
    "github.com/maxmanuylov/jongleur/util"
)

type Config struct {
    Items  string
    Port   int
    Period int
    Etcd   string
}

func Run(config *Config) error {
    if err := util.Check(config); err != nil {
        return err
    }

    // todo

    return nil
}