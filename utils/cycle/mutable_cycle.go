package cycle

import (
    "log"
    "sync"
)

type MutableCycle struct {
    c     chan<- string

    cycle *Cycle
    index map[string]bool
    logger *log.Logger

    lock  *sync.RWMutex
}

func NewMutableCycle(c chan<- string, logger *log.Logger) *MutableCycle {
    return &MutableCycle{c: c, cycle: nil, index: nil, logger: logger, lock: &sync.RWMutex{}}
}

func (mcycle *MutableCycle) SyncItems(newItems []string) {
    if mcycle.itemsDiffer(newItems) {
        mcycle.setNewItemsIfNeeded(newItems)
    }
}

func (mcycle *MutableCycle) Stop() {
    mcycle.lock.Lock()
    defer mcycle.lock.Unlock()

    mcycle.doStop()
}

func (mcycle *MutableCycle) itemsDiffer(newItems []string) bool {
    mcycle.lock.RLock()
    defer mcycle.lock.RUnlock()

    return mcycle.doItemsDiffer(newItems)
}

func (mcycle *MutableCycle) setNewItemsIfNeeded(newItems []string) {
    mcycle.lock.Lock()
    defer mcycle.lock.Unlock()

    if !mcycle.doItemsDiffer(newItems) {
        return
    }

    logger := mcycle.logger
    if logger != nil {
        mcycle.logger.Printf("Updating endpoints: %v\n", newItems)
    }

    mcycle.doStop()

    if len(newItems) != 0 {
        mcycle.index = make(map[string]bool)
        for _, item := range newItems {
            mcycle.index[item] = true
        }

        mcycle.cycle = NewCycle(newItems)
        mcycle.cycle.Start(mcycle.c)
    }

    if logger != nil {
        mcycle.logger.Println("Endpoints are updated")
    }
}

func (mcycle *MutableCycle) doItemsDiffer(newItems []string) bool {
    if mcycle.index == nil {
        return len(newItems) != 0
    }

    if len(newItems) != len(mcycle.index) {
        return true
    }

    for _, item := range newItems {
        if _, ok := mcycle.index[item]; !ok {
            return true
        }
    }

    return false
}

func (mcycle *MutableCycle) doStop() {
    if mcycle.cycle != nil {
        mcycle.cycle.Stop()
        mcycle.cycle = nil
    }
    mcycle.index = nil
}
