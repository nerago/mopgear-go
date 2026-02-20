package indexed

import (
	"paladin_gearing_go/model"
	"paladin_gearing_go/types/common"
	"paladin_gearing_go/types/items"
	"paladin_gearing_go/types/stats"
	"sync"
	"testing"
)

const (
	targetCount = 10
)

// test all individual item options are covered in a reasonable percent
// no or few duplicate sets considered

type peekRecorder struct {
	seen  []items.SolvableItemSet
	mutex sync.Mutex
}

func (record *peekRecorder) Len() int {
	return len(record.seen)
}
func (record *peekRecorder) Add(item *items.SolvableItemSet) {
	record.mutex.Lock()
	record.seen = append(record.seen, *item)
	record.mutex.Unlock()
}

func makeOptions() (*items.SolvableOptionsMap, *model.Model) {
	options := items.SolvableOptionsMap{}
	options[common.Equip_Head] = []items.SolvableItem{testItem(100, 11)}
	options[common.Equip_Neck] = []items.SolvableItem{testItem(200, 22), testItem(201, 23)}
	options[common.Equip_Shoulder] = []items.SolvableItem{testItem(301, 31), testItem(302, 32), testItem(303, 33), testItem(304, 34), testItem(305, 35)}

	model := model.Model_Testing()
	return &options, &model
}

func testItem(id, statValue uint32) items.SolvableItem {
	block := stats.StatBlock_of(stats.Stat_Crit, statValue)
	return items.SolvableItem{ItemId: id, TotalCap: block, TotalRated: block}
}

func TestMultiInt(t *testing.T) {
	peekRecord := peekRecorder{}
	options, model := makeOptions()
	max, skip := initSkipValues(options, targetCount)
	_ = mainLoop_multiThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	verifyRecord(t, &peekRecord)
}

func TestMultiBig(t *testing.T) {
	peekRecord := peekRecorder{}
	options, model := makeOptions()
	max, skip := initSkipValues(options, targetCount)
	mainLoop_multiThread_big(options, max, skip, model, peekRecord.Add)
	verifyRecord(t, &peekRecord)
}

func TestSingleInt(t *testing.T) {
	peekRecord := peekRecorder{}
	options, model := makeOptions()
	max, skip := initSkipValues(options, targetCount)
	mainLoop_singleThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	verifyRecord(t, &peekRecord)
}

func TestSingleBig(t *testing.T) {
	peekRecord := peekRecorder{}
	options, model := makeOptions()
	max, skip := initSkipValues(options, targetCount)
	mainLoop_singleThread_big(options, max, skip, model, peekRecord.Add)
	verifyRecord(t, &peekRecord)
}

func verifyRecord(t *testing.T, peekRecord *peekRecorder) {
	if peekRecord.Len() != targetCount {
		t.Fatalf("wrong count %d %d", peekRecord.Len(), targetCount)
	}
}
