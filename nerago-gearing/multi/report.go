package multi

import (
	"cmp"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/simrank"
	"slices"
)

func (job *MultiSetJob) reportSimResults(multiResultList []simulateMultiResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ RESULTS @@@@@@@@@@@@@@@@")
	for result := range util_collection.ForPointer(multiResultList) {
		job.reportSimResults_One(result)
	}
}

func (job *MultiSetJob) reportSimResults_One(result *simulateMultiResult) {
	job.printer.Printf("&&&&&&&&&&&&& %s\n", result.proposed.Id)
	job.printer.Printf("Weight Type %d\n", result.proposed.WeightType)

	if result.proposed.PermuteLabel != "" {
		job.printer.Println(result.proposed.PermuteLabel)
	}
	//result.proposed.Combo.Print(job.printer)

	for _, label := range job.paramOrderSlice() {
		prep := job.itemPrep[label]
		simData := result.simMap[label]
		output := result.proposed.Parts[label]
		input := prep.inputs

		job.printer.Printf("\n---------------- %s ----------------\n", label)
		output.Report(&prep.model, job.printer)
		job.printer.Println(simData.CompactStringGeneral())

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
			job.printer.PrintlnFromBuild(stringBuild)
			tools.WowSimJson_Write(&variantEquip, &prep.model, job.printer)
			job.printer.Println0()
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

		job.printer.PrintlnFromBuild(stringBuild)
	}

	job.printer.Println0()
	job.printer.Println0()
}

func (job *MultiSetJob) findVariantItem(result *simulateMultiResult, itemId items.ItemId, prep *specItemPrep) *items.FullItem {
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

	_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &prep.model, job.printer)
	return example
}
func (job *MultiSetJob) reportAsCsv(simResultList []*simMultiRankable) {
	job.printer.Println("@@@@@@@@@@@@@@@@ SPREADSHEET COPY @@@@@@@@@@@@@@@@")

	outputTypesByParam, rowCount, needPermuteLine := csvPrepareCollections(simResultList, job.itemPrep)
	labelOrder := job.paramOrderSlice()

	csv := csvStartHeader(labelOrder, outputTypesByParam, rowCount, needPermuteLine)

	for _, simResult := range simResultList {
		proposed := &simResult.result.proposed
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
		csv.AddInt(countReGem(proposed))
		if needPermuteLine {
			csv.AddString("\"" + proposed.PermuteLabel + "\"")
		}

		csv.FinishColumn()
	}

	csv.Write(job.printer)
}

func csvStartHeader(labelOrder []string, outputTypesByParam map[string][]stats.SimType, rowCount int, needPermuteLine bool) util.CSVOutputByColumn {
	csv := util.CSVOutputByColumn{}
	csv.InitRows(rowCount + 3)
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

func listReGem(multiProposed multi_types.MultiProposedOutput) []*items.FullItem {
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

func (job *MultiSetJob) handleBestRankedResult(best util_rank.BestCollector1[simulateMultiResult]) {
	job.printer.Println("Best ranked result")
	bestMultiResult := best.GetBestPointerOrPanic()
	job.reportSimResults_One(bestMultiResult)

	if job.input.WriteBestToGearFiles {
		job.writeToGearFiles(bestMultiResult)
	}
}

func (job *MultiSetJob) rankAllResults(resultList []simulateMultiResult) (util_rank.BestCollector1[simulateMultiResult], []*simMultiRankable) {
	// build nicely sortable set of pointer based slices
	multiRankSlice := util_collection.MapSliceAsNew(resultList, func(result *simulateMultiResult) *simMultiRankable {
		return &simMultiRankable{
			result: result,
			singles: util_collection.MapBasicMap(result.simMap, func(sim stats.SimData) *simSingleRankable {
				return &simSingleRankable{simData: sim}
			}),
		}
	})

	// build ranking per spec
	for label, prep := range job.itemPrep {
		entries := util_collection.MapSliceAsNew_NoPointer(multiRankSlice, func(mr *simMultiRankable) *simSingleRankable {
			return mr.singles[label]
		})

		simrank.RankSimsStatisticalFlat(prep.model.SimPriority.SimTypes(), entries, &prep.model.SimPriority)
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
	simData stats.SimData
	score   float64
	rank    int
}

func (s *simSingleRankable) GetSimData() *stats.SimData {
	return &s.simData
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
