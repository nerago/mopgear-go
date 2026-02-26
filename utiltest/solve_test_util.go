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
	TargetCountMinimal = 5  // worst slot has 5 options, should be able to try them all asap
	TargetCountFull = 40

	// TargetCountStandard = 8 // ends up with skip of 5, doesn't cycle shoulders properly since that matches a options size
)

func MakeTestOptions() (*items.SolvableOptionsMap, *model.Model) {
	options := items.SolvableOptionsMap{}
	options[common.Equip_Head] = []items.SolvableItem{testItem(100, 11)}
	options[common.Equip_Neck] = []items.SolvableItem{testItem(200, 22), testItem(201, 23)}
	options[common.Equip_Shoulder] = []items.SolvableItem{testItem(300, 31), testItem(301, 32), testItem(302, 33), testItem(303, 32), testItem(304, 31)}
	options[common.Equip_Back] = []items.SolvableItem{testItem(400, 44), testItem(401, 43), testItem(402, 42), testItem(403, 41)}
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
	if peekRecord.Len() < targetCount*3/4 || peekRecord.Len() > targetCount*5/4 {
		t.Fatalf("wrong count actual=%d expect=%d", peekRecord.Len(), targetCount)
	}

	verifyAllItemsTried(t, peekRecord, options)
	verifyUniqueCheckValid(t, options)
	verifySetsAllUnique(t, peekRecord.Seen)
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

func verifyUniqueCheckValid(t *testing.T, options *items.SolvableOptionsMap) {
	// test the test
	allSets := make([]items.SolvableItemSet, 0)
	allSets = recurAdd(allSets, options, 0, items.SolvableEquipMap{})
	verifySetsAllUnique(t, allSets)
}

func recurAdd(allSets []items.SolvableItemSet, options *items.SolvableOptionsMap, slot int, equip items.SolvableEquipMap) []items.SolvableItemSet {
	if slot >= len(options) {
		set := items.SolvableItemSet_Of(equip)
		return append(allSets, set)
	} else if len(options[slot]) > 0 {
		for _, item := range options[slot] {
			equip[slot] = &item
			allSets = recurAdd(allSets, options, slot + 1, equip)
		}
		return allSets
	} else {
		return recurAdd(allSets, options, slot + 1, equip)
	}
}

func verifySetsAllUnique(t *testing.T, seen []items.SolvableItemSet) {
	duplicateCount := 0
	for a := range seen {
		innerLoop: for b := a + 1; b < len(seen); b++ {
			if seen[a] == seen[b] {
				// t.Fatalf("duplicate sets %d %d", a, b)
				t.Logf("duplicate sets %d %d", a, b)
				duplicateCount++
				break innerLoop
			}
		}
	}

	if duplicateCount > len(seen) / 10 {
		t.Fatalf("duplicates %f%%", float64(duplicateCount) / float64(len(seen)) * 100.0)
	}
}
