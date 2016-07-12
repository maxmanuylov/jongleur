package main

import (
    "flag"
    "fmt"
    "github.com/maxmanuylov/jongleur/jongleur"
    "github.com/maxmanuylov/jongleur/item"
    "github.com/maxmanuylov/jongleur/util"
    "log"
    "os"
)

const (
    jongleurName = "jongleur"
    itemName = "item"
    jongleurItemName = jongleurName + " " + itemName
)

func main() {
    if len(os.Args) < 2 {
        printCommonUsageAndExit()
    }

    if os.Args[1] == itemName {
        runItem(os.Args[2:])
    } else {
        runJongleur(os.Args[1:])
    }
}

func runItem(args []string) {
    config := &item.Config{Health:item.StringHolder{}}

    flagSet := itemFlagSet(config)
    flagSet.Parse(args)

    if err := item.Run(config, newLogger()); err != nil {
        printErrorAndExit(err, jongleurItemName, flagSet)
    }
}

func runJongleur(args []string) {
    config := &jongleur.Config{}

    flagSet := jongleurFlagSet(config)
    flagSet.Parse(args)

    if err := jongleur.Run(config, newLogger()); err != nil {
        printErrorAndExit(err, jongleurName, flagSet)
    }
}

func itemFlagSet(config *item.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurItemName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurItemName, flagSet)

    flagSet.StringVar(&config.Type, "type", "", "service type; must be the same for all instances of the same service (required)")
    flagSet.StringVar(&config.Host, "host", "", "advertised host in \"<ip>:<port>\" format; service must be available from the network by this host (required)")
    flagSet.StringVar(&config.Health.Value, "health", "", "service health checking HTTP URL; response code 2xx is expected to treat service healthy; if not specified heath check is disabled")
    flagSet.IntVar(&config.Period, "period", 5, "health check period in seconds")
    flagSet.IntVar(&config.Tolerance, "tolerance", 3, "number of allowed sequential health check failures to not treat the service as dead")
    flagSet.StringVar(&config.Etcd, "etcd", "http://127.0.0.1:2379", "etcd URL")

    return flagSet
}

func jongleurFlagSet(config *jongleur.Config) *flag.FlagSet {
    flagSet := flag.NewFlagSet(jongleurName, flag.ExitOnError)

    flagSet.Usage = usageFunc(jongleurName, flagSet)

    flagSet.StringVar(&config.Items, "items", "", "type of the service to proxy (required)")
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

    if _, isUsageError := err.(util.UsageError); isUsageError {
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
