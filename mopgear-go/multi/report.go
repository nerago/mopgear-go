package multi

import (
	"cmp"
	"os"
	"slices"
	"time"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/simrank"
)

func (job *MainJob) reportSimResults(multiResultList []simulateMultiResult, includeConsole bool) {
	if includeConsole {
		job.printer.Println("@@@@@@@@@@@@@@@@ RESULTS @@@@@@@@@@@@@@@@")
		for result := range util_collection.ForPointer(multiResultList) {
			job.reportSimResults_One(result, job.printer)
		}
	}

	if job.setFilename != "" {
		err := util.WriteFuncToFileWithTemp(job.setFilename, func(file *os.File) {
			filePrint := util.FilePrintableMake(file)
			for result := range util_collection.ForPointer(multiResultList) {
				job.reportSimResults_One(result, filePrint)
			}
		})
		if err != nil {
			job.printer.Println(err.Error())
		}
	}
}

func (job *MainJob) reportSimResults_One(result *simulateMultiResult, printer util.Printable) {
	printer.Printf("&&&&&&&&&&&&& %s\n", result.proposed.Id)
	printer.Printf("Weight Type %d\n", result.proposed.WeightType)

	if result.proposed.PermuteLabel != "" {
		printer.Println(result.proposed.PermuteLabel)
	}
	//result.proposed.Combo.Print(job.printer)

	for _, label := range job.paramOrderSlice() {
		prep := job.itemPrep[label]
		simData := result.simMap[label]
		output := result.proposed.Parts[label]
		input := prep.inputs

		printer.Printf("\n---------------- %s ----------------\n", label)
		output.Report(prep.model, printer)
		printer.Println(simData.CompactStringGeneral())

		if len(input.ReportVariant) > 0 {
			variantEquip := *output.FullSet.Items()
			stringBuild := util.StringBuild2{}
			stringBuild.WriteString("---------------- ")
			stringBuild.WriteString(label)
			stringBuild.WriteRune(' ')
			for slot, itemId := range input.ReportVariant {
				variantItem := job.findVariantItem(result, itemId, prep)
				variantEquip[slot] = variantItem
				stringBuild.WriteString(variantItem.BaseName())
				stringBuild.WriteRune(' ')
			}
			stringBuild.WriteString(" ----------------")
			printer.PrintlnFromBuild(stringBuild)
			tools.WowSimJsonWrite(&variantEquip, prep.model, printer)
			printer.Println0()
		}
	}

	reGemItems := listReGem(result.proposed)
	if len(reGemItems) > 0 {
		stringBuild := util.StringBuild2{}
		stringBuild.WriteString("....... REGEM .......")
		for _, item := range reGemItems {
			item.AppendFullName(&stringBuild)
			stringBuild.WriteString(" : ")
			for _, gem := range item.GemChoice() {
				gem.AppendString(&stringBuild)
			}
			stringBuild.WriteRune('\n')
		}

		printer.PrintlnFromBuild(stringBuild)
	}

	upgradeItems := job.listUpgradeItems(result.proposed)
	if len(upgradeItems) > 0 {
		stringBuild := util.StringBuild2{}
		stringBuild.WriteString("....... UPGRADE .......")
		for _, item := range upgradeItems {
			item.AppendFullName(&stringBuild)
			stringBuild.WriteRune('\n')
		}

		printer.PrintlnFromBuild(stringBuild)
	}

	printer.Println0()
	printer.Println0()
}

func (job *MainJob) findVariantItem(result *simulateMultiResult, itemId items.ItemId, prep *specItemPrep) *items.FullItem {
	variantItem := result.proposed.FindItemById(itemId)
	if variantItem != nil {
		return variantItem
	}

	if item, found := prep.itemOptions.FindItemIdFirstOptional(itemId); found {
		return item
	}

	for _, otherPrep := range job.itemPrep {
		if item, found := otherPrep.itemOptions.FindItemIdFirstOptional(itemId); found {
			return item
		}
	}

	_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, prep.model, job.printer)
	return example
}
func (job *MainJob) reportAsCsv(simResultList []*simMultiRankable) {
	outputTypesByParam, rowCount, needPermuteLine := csvPrepareCollections(simResultList, job.itemPrep)
	labelOrder := job.paramOrderSlice()

	csv := csvStartHeader(labelOrder, outputTypesByParam, rowCount, needPermuteLine)
	job.csvBuildLines(simResultList, &csv, labelOrder, outputTypesByParam, needPermuteLine)

	if job.csvFilename != "" {
		job.writeToCsvFile(csv)
		job.printer.Println("@@ CSV output written to other file")
	} else {
		job.printer.Println("@@@@@@@@@@@@@@@@ SPREADSHEET COPY @@@@@@@@@@@@@@@@")
		csv.WriteTo(job.printer)
		job.printer.Println0()
		job.printer.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	}
}

func (job *MainJob) writeToCsvFile(csv util.CSVOutputByColumn) {
	err := util.WriteFuncToFileWithTemp(job.csvFilename, func(file *os.File) {
		csv.WriteTo(util.FilePrintableMake(file))
	})
	if err != nil {
		job.printer.Println(err.Error())
	}
}

func (job *MainJob) csvBuildLines(simResultList []*simMultiRankable, csv *util.CSVOutputByColumn, labelOrder []string, outputTypesByParam map[string][]stats.SimType, needPermuteLine bool) {
	for index, simResult := range simResultList {
		proposed := simResult.result.proposed
		simMap := simResult.result.simMap

		csv.AddString(proposed.Id)

		if len(simMap) != len(job.itemPrep) {
			panic("unexpected result size")
		}

		for _, label := range labelOrder {
			simData := simMap[label]
			simTypes := outputTypesByParam[label]
			for _, simType := range simTypes {
				value := simData.Get(simType)
				csv.AddFloat64(value, -1)
			}
		}

		csv.AddInt(int(proposed.WeightType))
		csv.AddInt(index + 1)
		csv.AddInt(countReGem(proposed))
		if needPermuteLine {
			csv.AddString("\"" + proposed.PermuteLabel + "\"")
		}

		csv.FinishColumn()
	}
}

func csvStartHeader(labelOrder []string, outputTypesByParam map[string][]stats.SimType, rowCount int, needPermuteLine bool) util.CSVOutputByColumn {
	csv := util.CSVOutputByColumn{}
	csv.InitRows(rowCount + 4)
	csv.AddString("id")
	for _, label := range labelOrder {
		simTypes := outputTypesByParam[label]
		for _, resultType := range simTypes {
			csv.AddToBuilder(func(b *util.StringBuild2) {
				b.WriteString(resultType.Name())
				b.WriteString(" (")
				b.WriteString(label)
				b.WriteRune(')')
			})
		}
	}
	csv.AddString("weight")
	csv.AddString("rank")
	csv.AddString("regem")
	if needPermuteLine {
		csv.AddString("permute")
	}
	csv.FinishColumn()
	return csv
}

func csvPrepareCollections(simResultList []*simMultiRankable, itemPrepMap map[string]*specItemPrep) (map[string][]stats.SimType, int, bool) {
	rowCount := 0
	outputTypesByParam := make(map[string][]stats.SimType, len(itemPrepMap))
	for label, prep := range itemPrepMap {
		simTypes := prep.model.SimPriority.SimTypes()
		outputTypesByParam[label] = simTypes
		rowCount += len(simTypes)
	}

	needPermuteLine := false
	for _, simResult := range simResultList {
		if simResult.result.proposed.PermuteLabel != "" {
			needPermuteLine = true
		}
	}
	if needPermuteLine {
		rowCount++
	}
	return outputTypesByParam, rowCount, needPermuteLine
}

func countReGem(multiProposed *multi_types.MultiProposedOutput) int {
	allItems := make([]*items.FullItem, 0)
	for _, part := range multiProposed.Parts {
		allItems = slices.AppendSeq(allItems, part.FullSet.Items().AllItemSeq())
	}
	util_collection.RemoveDuplicatesFunc_InPlace(&allItems, func(a, b **items.FullItem) bool { return (*a).Equals(*b) })

	countReGemItems := 0
	for _, item := range allItems {
		if item.HasBeenRegemmed() {
			countReGemItems++
		}
	}
	return countReGemItems
}

func listReGem(multiProposed *multi_types.MultiProposedOutput) []*items.FullItem {
	allItems := make([]*items.FullItem, 0)
	for _, part := range multiProposed.Parts {
		allItems = slices.AppendSeq(allItems, part.FullSet.Items().AllItemSeq())
	}
	util_collection.FilterSliceInPlace(&allItems, func(item **items.FullItem) bool {
		return (*item).HasBeenRegemmed()
	})
	util_collection.RemoveDuplicatesFunc_InPlace(&allItems, func(a, b **items.FullItem) bool { return (*a).Equals(*b) })
	return allItems
}

func (job *MainJob) listUpgradeItems(proposed *multi_types.MultiProposedOutput) []*items.FullItem {
	upgradeItems := make([]*items.FullItem, 0)
	for proposedItem := range proposed.SeqItem() {
		equipped := job.bagsGear.GetWithItemId(proposedItem.ItemId())
		if equipped != nil {
			bagItem := db.WowSimDB_LoadItemById(equipped.ItemId, equipped.UpgradeStepOrItemLevel)
			if proposedItem.UpgradeLevel() > bagItem.UpgradeLevel() {
				upgradeItems = append(upgradeItems, proposedItem)
			}
		} else {
			for _, prep := range job.itemPrep {
				equipItem := prep.exactEquippedGear.FindItemId(proposedItem.ItemId())
				if equipItem != nil && proposedItem.UpgradeLevel() > equipItem.UpgradeLevel() {
					upgradeItems = append(upgradeItems, proposedItem)
				}
			}
		}
	}
	util_collection.RemoveDuplicatesFunc_InPlace(&upgradeItems, func(a, b **items.FullItem) bool {
		return (*a).Equals(*b)
	})
	return upgradeItems
}

func (job *MainJob) handleBestRankedResult(best util_rank.BestCollector1[simulateMultiResult]) {
	job.printer.Println("Best ranked result")
	bestMultiResult := best.GetBestPointerOrPanic()
	job.reportSimResults_One(bestMultiResult, job.printer)

	if job.input.WriteBestToGearFiles {
		job.writeToGearFiles(bestMultiResult)
	}
}

func (job *MainJob) incrementalReporting(channel <-chan *simulateMultiResultPending, done *util_async.FutureVoid) []simulateMultiResult {
	updateLoopTime := time.Second * 30
	incrementalSlice := util_async.CollectChannelToSliceIncremental(channel)
	for done.WaitForLimitedDuration(updateLoopTime) {
		tempSlice := incrementalSlice.SliceTemp()
		completedResults := job.pendingToCompletedResults(tempSlice, false)
		_, rankedData := job.rankAllResults(completedResults)
		job.reportAsCsv(rankedData)
		job.reportSimResults(completedResults, false)
	}
	tempSlice := incrementalSlice.SliceTemp()
	return job.pendingToCompletedResults(tempSlice, true)
}

func (job *MainJob) pendingToCompletedResults(pendingSlice []*simulateMultiResultPending, isFinal bool) []simulateMultiResult {
	completed := make([]simulateMultiResult, 0)
pendingLoop:
	for _, pending := range pendingSlice {
		if pending.ready {
			simPending := pending.getPendingList()
			simMap := make(map[string]*stats.SimData, len(simPending))
			for k, v := range simPending {
				if v == nil {
					if isFinal {
						panic("ERROR: unexpected nil pending")
					} else {
						job.printer.Println("ERROR: unexpected nil pending")
						continue pendingLoop
					}
				} else if v.resultPending == nil {
					if isFinal {
						panic("ERROR: unexpected nil simData")
					} else {
						job.printer.Println("ERROR: unexpected nil simData")
						continue pendingLoop
					}
				}
				simMap[k] = v.resultPending
			}

			mr := simulateMultiResult{
				proposed: pending.proposed,
				simMap:   simMap,
			}
			completed = append(completed, mr)
		}
	}
	return completed
}

func (job *MainJob) rankAllResults(resultList []simulateMultiResult) (util_rank.BestCollector1[simulateMultiResult], []*simMultiRankable) {
	// build nicely sortable set of pointer based slices
	multiRankSlice := util_collection.MapSliceAsNew(resultList, func(result *simulateMultiResult) *simMultiRankable {
		return &simMultiRankable{
			result: result,
			singles: util_collection.MapBasicMap(result.simMap, func(sim *stats.SimData) *simSingleRankable {
				return &simSingleRankable{simData: sim}
			}),
		}
	})

	// build ranking per spec
	for label, prep := range job.itemPrep {
		entries := util_collection.MapSliceAsNew_NoPointer(multiRankSlice, func(mr *simMultiRankable) *simSingleRankable {
			return mr.singles[label]
		})

		simrank.RankSimsStatisticalFlatSingle(prep.model.SimPriority.SimTypes(), entries, &prep.model.SimPriority)
	}

	// highest combined rank across specs
	best := util_rank.BestCollector1[simulateMultiResult]{}
	for _, multiRank := range multiRankSlice {
		sumOfRanks := 0.0
		for label, singleRank := range multiRank.singles {
			prep := job.itemPrep[label]
			ratingPercent := prep.inputs.RequestRatingPercent
			sumOfRanks += float64(singleRank.rank) * ratingPercent
		}
		multiRank.sumOfRanks = sumOfRanks
		best.Offer(multiRank.result, sumOfRanks)
	}

	slices.SortFunc(multiRankSlice, func(a, b *simMultiRankable) int {
		return cmp.Compare(b.sumOfRanks, a.sumOfRanks)
	})

	return best, multiRankSlice
}

type simMultiRankable struct {
	result     *simulateMultiResult
	singles    map[string]*simSingleRankable
	sumOfRanks float64
}

type simSingleRankable struct {
	simData *stats.SimData
	score   float64
	rank    int
}

func (s *simSingleRankable) GetSimData() *stats.SimData {
	return s.simData
}

func (s *simSingleRankable) GetSimScore() float64 {
	return s.score
}

func (s *simSingleRankable) ResetSimScore() {
	s.score = 0
}

func (s *simSingleRankable) IncrementSimScore(add float64) {
	s.score += add
}

func (s *simSingleRankable) SetSimRank(targetRank int) {
	s.rank = targetRank
}

func (s *simSingleRankable) GetSimRank() int {
	return s.rank
}
