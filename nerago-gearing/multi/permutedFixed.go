package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/util/channel_op"
)

type fixedPermuteEntry struct {
	paramIndex int
	slot       items.SlotEquip
	itemId     items.ItemId
}

func (job *MultiSetJob) prepareFixedPermutations() <-chan []fixedPermuteEntry {
	optionEntriesList := make([][]fixedPermuteEntry, 0)
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for slot, itemIdList := range semiFixed {
			entriesList := make([]fixedPermuteEntry, 0)
			for _, itemId := range itemIdList {
				entry := fixedPermuteEntry{paramIndex, slot, itemId}
				entriesList = append(entriesList, entry)
			}
			optionEntriesList = append(optionEntriesList, entriesList)
		}
	}

	return channel_op.PermuteAsChannel(optionEntriesList)
}

// func applyAllowRate(specificAllowRates map[items.ItemId]float32, comboChannel chan commonCombo) <-chan commonCombo {
//     if len(specificAllowRates) == 1 {
//         itemId, rate := util.MapFirstEntry(specificAllowRates)
//         return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
//             func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
//                 rng := rand.New(rand.NewSource(int64(threadNum)))
//                 for combo := range inChan {
//                     applyAllowEntry(itemId, rate, &combo, rng)
//                     outChan <- combo
//                 }
//             })
//     } else {
//         return channel_op.TransformAll_ChannelToChannel(generateThreadCount, comboChannel,
//             func(threadNum int, inChan <-chan commonCombo, outChan chan<- commonCombo) {
//                 rng := rand.New(rand.NewSource(int64(threadNum)))
//                 for combo := range inChan {
//                     for itemId, rate := range specificAllowRates {
//                         applyAllowEntry(itemId, rate, &combo, rng)
//                     }
//                     outChan <- combo
//                 }
//             })
//     }
// }

// func applyAllowEntry(itemId items.ItemId, rate float32, combo *commonCombo, rng *rand.Rand) {
//     allow := rng.Float32() < rate
//     combo.setAllow(itemId, allow)
// }
// func printChosenCombo(combo commonCombo, printer *util.PrintRecorder) {
//     for itemId, entry := range combo.entryMap {
//         if entry.Forbidden {
//             printer.Printf("COMMON %d forbidden\n", itemId)
//         } else {
//             printer.Printf("common[%d] = stats.ReforgeRecipe{From: stats.%s, To: stats.%s}\n", item.ItemId(), item.Reforge.From.EnumName(), item.Reforge.To.EnumName())
//             printer.Printf("COMMON %s\n", entry.Item.CreateString())
//         }
//     }
//     for _, entry := range combo.entryMap {
//         if !entry.Forbidden {
//             item := entry.Item
//             if item.Reforge.IsEmpty() {
//                 printer.Printf("common[%d] = stats.ReforgeRecipe_empty\n", item.ItemId())
//             } else {
//                 printer.Printf("common[%d] = stats.ReforgeRecipe{From: stats.%s, To: stats.%s}\n", item.ItemId(), item.Reforge.From.EnumName(), item.Reforge.To.EnumName())
//             }
//         }
//     }
// }

// type splitHighCollector struct {
//     allowIds       []items.ItemId
//     highCollectors []util_rank.HighestCollectorN[multiProposedOutput]
// }

// func splitHighCollector_make(specificAllowRates map[items.ItemId]float32, topCount uint64) splitHighCollector {
//     collector := splitHighCollector{}

//     splitCount := []uint64{topCount}
//     for itemId, percent := range specificAllowRates {
//         nextSplitCount := make([]uint64, 0, len(splitCount)*2)
//         for _, count := range splitCount {
//             trueCount := uint64(float32(count) * percent)
//             falseCount := count - trueCount
//             nextSplitCount = append(nextSplitCount, falseCount, trueCount)
//         }
//         splitCount = nextSplitCount

//         collector.allowIds = append(collector.allowIds, itemId)
//     }

//     for _, count := range splitCount {
//         collector.highCollectors = append(collector.highCollectors, util_rank.HighestCollector_ForN(count, (*multiProposedOutput).Equals))
//     }

//     return collector
// }

// func splitHighCollector_OfChannel(channel <-chan splitHighCollector, expectNum int) []multiProposedOutput {
//     var collector *splitHighCollector = nil
//     for range expectNum {
//         threadResult := <-channel
//         if collector == nil {
//             collector = &threadResult
//         } else {
//             collector.Merge_Mutating(&threadResult)
//         }
//     }
//     return collector.ResultsFlat()
// }

// func (collector *splitHighCollector) Offer(output *multiProposedOutput) {
//     choices := output.combo.allowChoices

//     index := 0
//     blockSize := len(collector.highCollectors) / 2

//     for _, itemId := range collector.allowIds {
//         itemChoice := choices[itemId]
//         if itemChoice == true {
//             index += blockSize
//         }
//         blockSize /= 2
//     }

//     collector.highCollectors[index].Offer(output, output.totalRatingSum)
// }

// func (collector *splitHighCollector) Merge_Mutating(other *splitHighCollector) {
//     for i := range collector.highCollectors {
//         collector.highCollectors[i].Merge_Mutating(&other.highCollectors[i])
//     }
// }

// func (collector *splitHighCollector) ResultsFlat() []multiProposedOutput {
//     result := []multiProposedOutput{}
//     for i := range collector.highCollectors {
//         subList := collector.highCollectors[i].ResultsFlat()
//         result = append(result, subList...)
//     }
//     return result
// }
