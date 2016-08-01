package boot

import (
    "flag"
    "fmt"
    "github.com/maxmanuylov/jongleur/item"
    "github.com/maxmanuylov/jongleur/jongleur"
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
    jongleurItemName = jongleurName + " " + itemName
    jongleurEtcdName = jongleurName + " " + etcdName
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

    flagSet.BoolVar(&config.Local, "local", false, "flag to restrict listen interface to \"127.0.0.1\"; default is \"0.0.0.0\"")
    flagSet.IntVar(&config.Port, "port", 2379, "local port to listen")
    flagSet.IntVar(&config.Period, "period", 10, "etcd members list synchronization period in seconds")
    flagSet.StringVar(&config.Discovery, "discovery", "", "etcd discovery URL (required)")

    return flagSet
}

func jongleurFlagSet(config *regular.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurName, flagSet)

    flagSet.StringVar(&config.Items, "items", "", "type of the service to proxy (required)")
    flagSet.BoolVar(&config.Local, "local", false, "flag to restrict listen interface to \"127.0.0.1\"; default is \"0.0.0.0\"")
    flagSet.IntVar(&config.Port, "port", 0, "local port to listen; interface to listen is always \"0.0.0.0\" (required)")
    flagSet.IntVar(&config.Period, "period", 10, "service instances list synchronization period in seconds")
    flagSet.StringVar(&config.Etcd, "etcd", "http://127.0.0.1:2379", "etcd URL")

    return flagSet
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
