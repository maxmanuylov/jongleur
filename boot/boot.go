package boot

import (
    "flag"
    "fmt"
    "github.com/maxmanuylov/jongleur/item"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/jongleur/ceph"
    "github.com/maxmanuylov/jongleur/jongleur/etcd"
    "github.com/maxmanuylov/jongleur/jongleur/regular"
    "github.com/maxmanuylov/jongleur/utils"
    "log"
    "os"
)

const (
    jongleurName = "jongleur"
    itemName = "item"
    etcdName = "etcd"
    cephName = "ceph"
    jongleurItemName = jongleurName + " " + itemName
    jongleurEtcdName = jongleurName + " " + etcdName
    jongleurCephName = jongleurName + " " + cephName
)

func Run() {
    if len(os.Args) < 2 {
        printCommonUsageAndExit()
    }

    switch os.Args[1] {
    case itemName:
        runItem(os.Args[2:])
    case etcdName:
        runEtcdProxy(os.Args[2:])
    case cephName:
        runCephMonProxy(os.Args[2:])
    default:
        runJongleur(os.Args[1:])
    }
}

func runItem(args []string) {
    config := &item.Config{Health:&item.StringHolder{}}

    flagSet := itemFlagSet(config)
    flagSet.Parse(args)

    if err := item.Run(config, newLogger()); err != nil {
        printErrorAndExit(err, jongleurItemName, flagSet)
    }
}

func runEtcdProxy(args []string) {
    config := &etcd.Config{}

    flagSet := etcdFlagSet(config)
    flagSet.Parse(args)

    jongleurConfig, err := config.ToJongleurConfig()
    if err != nil {
        printErrorAndExit(err, jongleurEtcdName, flagSet)
    }

    if err := jongleur.Run(jongleurConfig, newLogger()); err != nil {
        printErrorAndExit(err, jongleurEtcdName, flagSet)
    }
}

func runCephMonProxy(args []string) {
    config := &ceph.Config{}

    flagSet := cephFlagSet(config)
    flagSet.Parse(args)

    jongleurConfig, err := config.ToJongleurConfig()
    if err != nil {
        printErrorAndExit(err, jongleurCephName, flagSet)
    }

    if err := jongleur.Run(jongleurConfig, newLogger()); err != nil {
        printErrorAndExit(err, jongleurCephName, flagSet)
    }
}

func runJongleur(args []string) {
    config := &regular.Config{}

    flagSet := jongleurFlagSet(config)
    flagSet.Parse(args)

    jongleurConfig, err := config.ToJongleurConfig()
    if err != nil {
        printErrorAndExit(err, jongleurName, flagSet)
    }

    if err := jongleur.Run(jongleurConfig, newLogger()); err != nil {
        printErrorAndExit(err, jongleurName, flagSet)
    }
}

func itemFlagSet(config *item.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurItemName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurItemName, flagSet)

    flagSet.StringVar(&config.Type, "type", "", "service type; must be the same for all instances of the same service (required)")
    flagSet.StringVar(&config.Host, "host", "", "advertised host; use \"<ip>:<port>\" format to advertise the specified port and \"<ip>:*\" format to advertise all the ports (request destination port is used in this case); service must be available from the network by this host (required)")
    flagSet.StringVar(&config.Health.Value, "health", "", "service health checking HTTP URL; response code 2xx is expected to treat service healthy; if not specified heath check is disabled")
    flagSet.IntVar(&config.Period, "period", 5, "health check period in seconds")
    flagSet.IntVar(&config.Tolerance, "tolerance", 3, "number of allowed sequential health check failures to not treat the service as dead")
    flagSet.StringVar(&config.Etcd, "etcd", "http://127.0.0.1:2379", "etcd URL")

    return flagSet
}

func etcdFlagSet(config *etcd.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurEtcdName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurEtcdName, flagSet)

    flagSet.BoolVar(&config.Verbose, "verbose", false, "flag to enable verbose output")
    flagSet.StringVar(&config.Listen, "listen", "", "listen address in form \"[<network>@]address\"; network can be \"tcp\" or \"unix\"; default network is \"tcp\" (required)")
    flagSet.IntVar(&config.Period, "period", 10, "service instances list synchronization period in seconds")
    flagSet.StringVar(&config.Discovery, "discovery", "", "etcd discovery URL (required)")

    return flagSet
}

func cephFlagSet(config *ceph.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurCephName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurCephName, flagSet)

    appendJongleurFlags(&config.Config, flagSet)

    return flagSet
}

func jongleurFlagSet(config *regular.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurName, flagSet)

    appendJongleurFlags(config, flagSet)

    return flagSet
}

func appendJongleurFlags(config *regular.Config, flagSet *flag.FlagSet) {
    flagSet.BoolVar(&config.Verbose, "verbose", false, "flag to enable verbose output")
    flagSet.StringVar(&config.Items, "items", "", "type of the service to proxy (required)")
    flagSet.StringVar(&config.Listen, "listen", "", "listen address in form \"[<network>@]address\"; network can be \"tcp\" or \"unix\"; default network is \"tcp\" (required)")
    flagSet.IntVar(&config.RemotePort, "remote-port", -1, "remote port to transfer requests to in case of using \"*\" for item ports; by default items with \"*\" ports are ignored")
    flagSet.IntVar(&config.Period, "period", 10, "service instances list synchronization period in seconds")
    flagSet.StringVar(&config.Etcd, "etcd", "http://127.0.0.1:2379", "etcd URL")
}

func printCommonUsageAndExit() {
    fmt.Fprintln(os.Stderr, "Usage:")
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintf(os.Stderr, "  * %s <options>", jongleurItemName)
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintln(os.Stderr, "\truns service instance keeper")
    fmt.Fprintf(os.Stderr, "  * %s <options>", jongleurName)
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintln(os.Stderr, "\truns load balancing proxy")
    fmt.Fprintf(os.Stderr, "  * %s [%s] --help", jongleurName, itemName)
    fmt.Fprintln(os.Stderr, "")
    fmt.Fprintln(os.Stderr, "\tshows detailed options")
    fmt.Fprintln(os.Stderr, "")

    os.Exit(2)
}

func printErrorAndExit(err error, name string, flagSet *flag.FlagSet) {
    fmt.Fprintln(os.Stderr, err.Error())

    if _, isUsageError := err.(utils.UsageError); isUsageError {
        fmt.Fprintln(os.Stderr, "")
        usageFunc(name, flagSet)()
        os.Exit(2)
    }

    os.Exit(255)
}

func usageFunc(name string, flagSet *flag.FlagSet) func() {
    return func() {
        fmt.Fprintf(os.Stderr, "Available \"%s\" options:", name)
        fmt.Fprintln(os.Stderr, "")
        fmt.Fprintln(os.Stderr, "")
        flagSet.PrintDefaults()
        fmt.Fprintln(os.Stderr, "")
    }
}

func newLogger() *log.Logger {
    return log.New(os.Stderr, "", log.LstdFlags | log.Lshortfile)
}
