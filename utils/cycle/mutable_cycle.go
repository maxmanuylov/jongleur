package cycle

import (
    "sync"
)

type MutableCycle struct {
    c     chan<- string

    cycle *Cycle
    index map[string]bool

    lock  *sync.RWMutex
}

func NewMutableCycle(c chan<- string) *MutableCycle {
    return &MutableCycle{c, nil, nil, &sync.RWMutex{}}
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

    mcycle.doStop()

    if len(newItems) != 0 {
        mcycle.index = make(map[string]bool)
        for _, item := range newItems {
            mcycle.index[item] = true
        }

        mcycle.cycle = NewCycle(newItems)
        mcycle.cycle.Start(mcycle.c)
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
