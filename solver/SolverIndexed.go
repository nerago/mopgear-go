package solver

import (
	"fmt"
	"iter"
	"math/big"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/util"
	"time"
)

var int_one = big.NewInt(1)

func SolverIndexed_RunFull(itemOptions *SolvableOptionsMap, model *Model) SolvableItemSet {
	slotSizes := slotSizes(itemOptions)
	max := itemOptions.TotalCombinationCount()
	index := big.NewInt(0)
	best := BestCollector1[SolvableItemSet]{}

	go trackProgress(index, max)

	for index.Cmp(max) < 0 {
		set := makeSet(itemOptions, &slotSizes, index)
		if model.CheckSet(set) {
			rating := model.CalcRating(set)
			best.Add(set, rating)
		}
		index.Add(index, int_one)
	}

	return best.GetBest()
}

func trackProgress(index *big.Int, max *big.Int) {
	startTime := time.Now()
	for true {
		time.Sleep(time.Second * 5)

		var ratio big.Rat
		ratio.SetFrac(index, max)
		percent, _ := ratio.Float64()

		if percent > 0 {
			timeTaken := time.Since(startTime)
			totalEstimate := time.Duration(float64(timeTaken) / percent)
			estimateRemain := totalEstimate - timeTaken
			fmt.Printf("%d $.1f%% %s\n", index, percent*100, estimateRemain.String())
		}
	}
}

func slotSizes(itemOptions *SolvableOptionsMap) [16]*big.Int {
	slotSizes := [16]*big.Int{}
	for i, array := range itemOptions {
		slotSizes[i] = big.NewInt(int64(len(array)))
	}
}

func makeSet(itemOptions *SolvableOptionsMap, slotSizes *[16]*big.Int, mainIndex *big.Int) *SolvableItemSet {
	equip := SolvableEquipMap{}

	var div, mod big.Int
	for slot, array := range itemOptions {
		size := slotSizes[slot]
		div.DivMod(mainIndex, size, &mod)

		slotIndex := mod.Int64()
		choice := &array[slotIndex]
		equip[slot] = choice

		mainIndex = &div
	}

	return SolvableItemSet_Of(equip)
}

func seq_SolverIndexed_RunFull(itemOptions *SolvableOptionsMap, model *Model) {
	// combos := itemOptions.TotalCombinationCount()
	// indexSeq = makeIndexSeq(combos, int_one)
}

func makeIndexSeq(max *big.Int, skip *big.Int) iter.Seq[*big.Int] {
	return func(yield func(*big.Int) bool) {
		value := big.NewInt(0)
		for value.Cmp(max) < 0 {
			if !yield(value) {
				return
			}
			value.Add(value, int_one)
		}
	}
}
