package multi

import (
	"sync"

	"github.com/nerago/mopgear-go/items"
)

const (
	c_defaultTimeoutSeconds          = 600
	c_prepThreadCount                = 8
	c_simThreadCount                 = 4
	c_highsThreadCount               = 10
	c_permuteThreadCount             = 8
	c_decimateThreadCount            = 12
	c_mainProposal_threadCount       = 4
	c_additionalProposal_threadCount = 4
)

type seenMap struct {
	content map[items.ItemId]uint32
	mutex   sync.Mutex
}

func (seen *seenMap) Add(equipMap *items.FullEquipMap) {
	seen.mutex.Lock()
	defer seen.mutex.Unlock()

	for item := range equipMap.AllItemSeq() {
		seen.content[item.ItemId()]++
	}
}

func (seen *seenMap) AddScaled(equipMap *items.FullEquipMap, scale uint32) {
	seen.mutex.Lock()
	defer seen.mutex.Unlock()

	for item := range equipMap.AllItemSeq() {
		seen.content[item.ItemId()] += scale
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
