package multi

import (
	"paladin_gearing_go/items"
	"sync"
)

const (
	c_prepThreadCount  = 8
	c_simThreadCount   = 4
	c_highsThreadCount = 10
)

type seenMap struct {
	content map[items.ItemId]uint32
	mutex   sync.Mutex
}

func (seen *seenMap) Add(itemSet *items.FullItemSet) {
	seen.mutex.Lock()
	defer seen.mutex.Unlock()

	for item := range itemSet.Items().AllItemSeq() {
		seen.content[item.ItemId()]++
	}
}

func (seen *seenMap) AddOther(other *seenMap) {
	seen.mutex.Lock()
	other.mutex.Lock()
	defer other.mutex.Unlock()
	defer seen.mutex.Unlock()

	for id, count := range other.content {
		seen.content[id] += count
	}
}
