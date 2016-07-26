#!/bin/bash

ETCD_VERSION="v2.3.2"

rm -rf vendor

govendor init

govendor fetch github.com/coreos/etcd@=$ETCD_VERSION
govendor fetch golang.org/x/net/context