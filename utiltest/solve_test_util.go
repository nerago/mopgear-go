package utiltest

import (
	"paladin_gearing_go/model"
	"paladin_gearing_go/types/common"
	"paladin_gearing_go/types/items"
	"paladin_gearing_go/types/stats"
	"sync"
	"testing"
)

const (
	TargetCountStandard = 13 // initially matches one from indexed_test
	TargetCountMinimal = 13  // worst slot has 5 options, should be able to try them all asap
	TargetCountFull = 40

	// TargetCountStandard = 8 // ends up with skip of 5, doesn't cycle shoulders properly since that matches a options size
)

func MakeTestOptions() (*items.SolvableOptionsMap, *model.Model) {
	options := items.SolvableOptionsMap{}
	options[common.Equip_Head] = []items.SolvableItem{testItem(100, 11)}
	options[common.Equip_Neck] = []items.SolvableItem{testItem(200, 22), testItem(201, 23)}
	options[common.Equip_Shoulder] = []items.SolvableItem{testItem(301, 31), testItem(302, 32), testItem(303, 33), testItem(304, 32), testItem(305, 31)}
	options[common.Equip_Back] = []items.SolvableItem{testItem(401, 44), testItem(402, 43), testItem(403, 42), testItem(404, 41)}
	model := model.Model_Testing()
	return &options, &model
}

func MakeTestExpectedBest() items.SolvableEquipMap {
	equip := items.SolvableEquipMap{}
	equip[common.Equip_Head] = testItemPointer(100, 11)
	equip[common.Equip_Neck] = testItemPointer(201, 23)
	equip[common.Equip_Shoulder] = testItemPointer(303, 33)
	equip[common.Equip_Back] = testItemPointer(401, 44)
	return equip
}

func testItem(id, statValue uint32) items.SolvableItem {
	block := stats.StatBlock_of(stats.Stat_Crit, statValue)
	return items.SolvableItem{ItemId: id, TotalCap: block, TotalRated: block}
}

func testItemPointer(id, statValue uint32) *items.SolvableItem {
	item := testItem(id, statValue)
	return &item
}

type PeekTestRecorder struct {
	Seen  []items.SolvableItemSet
	mutex sync.Mutex
}

func (record *PeekTestRecorder) Len() int {
	return len(record.Seen)
}
func (record *PeekTestRecorder) Add(item *items.SolvableItemSet) {
	record.mutex.Lock()
	record.Seen = append(record.Seen, *item)
	record.mutex.Unlock()
}

// test all individual item options are covered in a reasonable percent
// no or few duplicate sets considered
func VerifyRecord(t *testing.T, peekRecord *PeekTestRecorder, options *items.SolvableOptionsMap, targetCount int) {
	if peekRecord.Len() < targetCount-1 || peekRecord.Len() > targetCount+1 {
		t.Fatalf("wrong count actual=%d expect=%d", peekRecord.Len(), targetCount)
	}

	verifyAllItemsTried(t, peekRecord, options)
	verifySetsAllUnique(t, peekRecord)
}

func verifyAllItemsTried(t *testing.T, peekRecord *PeekTestRecorder, options *items.SolvableOptionsMap) {
	seenCounts := make(map[uint32]int)
	for _, itemSet := range peekRecord.Seen {
		for item := range itemSet.Items.AllItemSeq() {
			seenCounts[item.ItemId]++
		}
	}

	for item := range options.AllItemSeq() {
		if seenCounts[item.ItemId] == 0 {
			t.Fatalf("never tried item %d", item.ItemId)
		}
	}
}

func verifySetsAllUnique(t *testing.T, peekRecord *PeekTestRecorder) {
	for a := range peekRecord.Seen {
		for b := a + 1; b < len(peekRecord.Seen); b++ {
			if peekRecord.Seen[a] == peekRecord.Seen[b] {
				t.Fatalf("duplicate sets %d %d", a, b)
			}
		}
	}
}
