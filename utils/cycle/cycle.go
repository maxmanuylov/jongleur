package cycle

type Cycle struct {
    items   []string
    nextI   int
    stop    chan bool
}

func NewCycle(items []string) *Cycle {
    if len(items) == 0 {
        panic("Attempt to create an empty cycle")
    }
    return &Cycle{items, 0, make(chan bool, 1)}
}

func (cycle *Cycle) Start(c chan<- string) {
    go cycle.run(c)
}

func (cycle *Cycle) Stop() {
    cycle.stop <- true
}

func (cycle *Cycle) run(c chan<- string) {
    for {
        select {
        case <-cycle.stop:
            return
        default:
            c <- cycle.items[cycle.nextI]
            cycle.nextI++
            cycle.nextI %= len(cycle.items)
        }
    }
}
