package multi

import (
	"cmp"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func checkNoConflicts(outputSet []multi_types.SingleProposedOutput, printer *util.PrintRecorder) bool {
	itemByRef := make(map[items.ItemRef]*items.FullItem)
	for outputIndex := range outputSet {
		for item := range outputSet[outputIndex].FullSet.Items().AllItemSeq() {
			ref := items.ItemRef_Of(item)
			existing, found := itemByRef[ref]
			if !found {
				itemByRef[ref] = item
			} else if !existing.Equals(item) {
				printer.Printf("!! CONFLICT %s\n!!          %s\n", item.CreateString(), existing.CreateString())
				return false
			}
		}
	}
	return true
}

type simulateJob struct {
	spec        stats.SpecType
	goal        stats.OptimiseGoal
	fight       stats.WowSim_Fight
	simSpeedUp  int
	equip       items.FullEquipMap
	professions gear_model.ProfessionInfo
}

type simulateJobResult struct {
	job    simulateJob
	result stats.SimData
}

func (simJob *simulateJob) Equals(other *simulateJob) bool {
	return simJob.spec == other.spec && simJob.fight == other.fight && simJob.equip.Equals(&other.equip) && simJob.professions == other.professions
}

type simulateMultiResult struct {
	proposed multi_types.MultiProposedOutput
	result   []stats.SimData
}

func (job *MultiSetJob) prepareSimList(proposalList <-chan multi_types.MultiProposedOutput) <-chan simulateJob {
	jobChannel := util_async.MapMulti_ChannelToChannel(2, proposalList, func(proposal multi_types.MultiProposedOutput, nextChan chan<- simulateJob) {
		for _, output := range proposal.Parts {
			nextChan <- simulateJob{output.Spec, output.Model.Goal, output.Model.SimulateAs, output.Model.SimSpeedUp, *output.FullSet.Items(), output.Model.Professions}
		}
	})

	return util_async.Channel_RemoveDuplicatesFunc(jobChannel, (*simulateJob).Equals)
}

func (job *MultiSetJob) runSims(jobChan <-chan simulateJob, trackProgress *util.TrackProgress, expectedCount *util_async.Future[int]) *util_async.FutureCancellable[[]simulateJobResult] {
	trackProgress.RunOuterTracking(0)
	expectedCount.ForwardSuccessfulResultToCallback(func(count int) {
		trackProgress.UpdateExpectedChildCount(count)
	})

	return util_async.Map_ChannelToSlice_FutureCancellable(simThreadCount, jobChan, trackProgress.SetDone, func(sim simulateJob) simulateJobResult {
		result := simulate.WowSim_Execute_SpecifyAll(job.simRunSize, sim.simSpeedUp, sim.spec, sim.goal, sim.fight, sim.professions, &sim.equip, nil, trackProgress.NewChild())
		job.printer.Printf("sim %22s fight=%d %s\n", sim.spec.Name(), sim.fight, result.CompactStringGeneral())
		return simulateJobResult{sim, result}
	})
}

func (job *MultiSetJob) linkSimResults(proposalList []multi_types.MultiProposedOutput, jobList []simulateJobResult) []simulateMultiResult {
	resultList := make([]simulateMultiResult, 0, len(proposalList))
	for _, proposal := range proposalList {
		result := linkSimResult(proposal, jobList)
		resultList = append(resultList, result)
	}
	return resultList
}

func linkSimResult(proposal multi_types.MultiProposedOutput, resultList []simulateJobResult) simulateMultiResult {
	multiResult := simulateMultiResult{proposal, make([]stats.SimData, len(proposal.Parts))}
	for partIndex := range proposal.Parts {
		part := &proposal.Parts[partIndex]
		for resultIndex := range resultList {
			simResult := &resultList[resultIndex]
			if part.FullSet.Items().Equals(&simResult.job.equip) && part.Spec == simResult.job.spec && part.Model.SimulateAs == simResult.job.fight {
				multiResult.result[partIndex] = simResult.result
				break
			}
		}
	}
	return multiResult
}

func (job *MultiSetJob) reportSimResults(multiResultList []simulateMultiResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ RESULTS @@@@@@@@@@@@@@@@")
	for _, result := range multiResultList {
		job.reportSimResults_One(result)
	}
}

func (job *MultiSetJob) reportSimResults_One(result simulateMultiResult) {
	job.printer.Printf("&&&&&&&&&&&&& %s\n", result.proposed.Id)
	if result.proposed.PermuteLabel != "" {
		job.printer.Println(result.proposed.PermuteLabel)
	}
	result.proposed.Combo.Print(job.printer)

	for specIndex, specResult := range result.result {
		param := &job.params[specIndex]
		job.printer.Printf("\n---------------- %s ----------------\n", param.Label)

		output := result.proposed.Parts[specIndex]
		output.Report(job.printer)
		specResult.Print(job.printer)

		if len(param.ReportVariant) > 0 {
			variantEquip := *output.FullSet.Items()
			stringBuild := util.StringBuild2{}
			stringBuild.WriteString("---------------- ")
			stringBuild.WriteString(param.Label)
			stringBuild.WriteRune(' ')
			for slot, itemId := range param.ReportVariant {
				variantItem := job.findVariantItem(result, itemId, param)
				variantEquip[slot] = variantItem
				stringBuild.WriteString(variantItem.BaseName())
				stringBuild.WriteRune(' ')
			}
			stringBuild.WriteString(" ----------------")
			job.printer.PrintlnFromBuild(stringBuild)
			tools.WowSimJson_Write(&variantEquip, &param.Model, job.printer)
			job.printer.Println0()
		}
	}

	regemmedItems := listRegem(result.proposed)
	if len(regemmedItems) > 0 {
		stringBuild := util.StringBuild2{}
		stringBuild.WriteString("....... REGEM .......")
		for _, item := range regemmedItems {
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

func (job *MultiSetJob) findVariantItem(result simulateMultiResult, itemId items.ItemId, param *multiSetParamInternal) *items.FullItem {
	variantItem := result.proposed.FindItemById(itemId)
	if variantItem != nil {
		return variantItem
	}

	if item, found := param.itemOptions.FindItemIdFirstOptional(itemId); found {
		return item
	}

	for paramIndex := range job.params {
		otherParam := &job.params[paramIndex]
		if item, found := otherParam.itemOptions.FindItemIdFirstOptional(itemId); found {
			return item
		}
	}

	_, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, items.MAX_UPGRADE_LEVEL, items.NO_RANDOM_SUFFIX, &param.Model, job.printer)
	return example
}

func (job *MultiSetJob) reportAsCsv(simResultList []simulateMultiResult) {
	job.printer.Println("@@@@@@@@@@@@@@@@ SPREADSHEET COPY @@@@@@@@@@@@@@@@")

	rowCount := 0
	outputTypesByParam := make([][]stats.SimType, len(job.params))
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		simTypes := param.Model.SimPriority.SimTypes()
		outputTypesByParam[paramIndex] = simTypes
		rowCount += len(simTypes)
	}

	needPermuteLine := false
	for _, simResult := range simResultList {
		if simResult.proposed.PermuteLabel != "" {
			needPermuteLine = true
		}
	}
	if needPermuteLine {
		rowCount++
	}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(rowCount + 2)
	csv.AddString("id")
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		simTypes := outputTypesByParam[paramIndex]
		for _, resultType := range simTypes {
			csv.AddToBuilder(func(b *util.StringBuild2) {
				b.WriteString(resultType.Name())
				b.WriteString(" (")
				b.WriteString(param.Label)
				b.WriteRune(')')
			})
		}
	}
	csv.AddString("regem")
	if needPermuteLine {
		csv.AddString("permute")
	}
	csv.FinishColumn()

	for _, simResult := range simResultList {
		csv.AddString(simResult.proposed.Id)

		if len(simResult.result) != len(job.params) {
			panic("unexpected result size")
		}

		for paramIndex, resultStat := range simResult.result {
			simTypes := outputTypesByParam[paramIndex]
			for _, resultType := range simTypes {
				value := resultStat.Get(resultType)
				csv.AddFloat64(value, -1)
			}
		}

		csv.AddInt(countRegem(simResult.proposed))
		if needPermuteLine {
			csv.AddString("\"" + simResult.proposed.PermuteLabel + "\"")
		}

		csv.FinishColumn()
	}

	csv.Write(job.printer)
}

func countRegem(multiProposed multi_types.MultiProposedOutput) int {
	allItems := make([]*items.FullItem, 0)
	for part := range util_collection.ForPointer(multiProposed.Parts) {
		allItems = slices.AppendSeq(allItems, part.FullSet.Items().AllItemSeq())
	}
	util_collection.RemoveDuplicatesFunc_InPlace(&allItems, func(a, b **items.FullItem) bool { return (*a).Equals(*b) })

	countRegemmed := 0
	for _, item := range allItems {
		if item.HasBeenRegemmed() {
			countRegemmed++
		}
	}
	return countRegemmed
}

func listRegem(multiProposed multi_types.MultiProposedOutput) []*items.FullItem {
	itemSlice := make([]*items.FullItem, 0)
	for part := range util_collection.ForPointer(multiProposed.Parts) {
		itemSlice = slices.AppendSeq(itemSlice, part.FullSet.Items().AllItemSeq())
	}
	util_collection.FilterSliceInPlace(&itemSlice, func(item **items.FullItem) bool {
		return (*item).HasBeenRegemmed()
	})
	util_collection.RemoveDuplicatesFunc_InPlace(&itemSlice, func(a, b **items.FullItem) bool { return (*a).Equals(*b) })
	return itemSlice
}

// TODO refactor with Accuracy code
func (job *MultiSetJob) suggestResultFromRankings(results []simulateMultiResult) {
	rankInputArrays := util_collection.MapMapSlice[int, stats.SimType, float64]{}
	for _, result := range results {
		for paramIndex, simStats := range result.result {
			simTypeList := job.params[paramIndex].Model.SimPriority.SimTypes()
			for _, simType := range simTypeList {
				value := simStats.Get(simType)
				rankInputArrays.Add(paramIndex, simType, value)
			}
		}
	}

	for paramIndex := range job.params {
		simTypeList := job.params[paramIndex].Model.SimPriority.SimTypes()
		for _, simType := range simTypeList {
			rankInputArrays.MapInternalSlice(paramIndex, simType, func(rankValues []float64) []float64 {
				if simType.IsHighGood() {
					slices.SortFunc(rankValues, func(a, b float64) int { return cmp.Compare(a, b) })
				} else {
					slices.SortFunc(rankValues, func(a, b float64) int { return cmp.Compare(b, a) })
				}
				return rankValues
			})
		}
	}

	best := util_rank.BestCollector1[simulateMultiResult]{}

	for _, result := range results {
		sumOfRanks := 0.0
		for paramIndex, simStats := range result.result {
			ratingPercent := job.params[paramIndex].RequestRatingPercent
			simTypeList := job.params[paramIndex].Model.SimPriority.SimTypes()
			for _, simType := range simTypeList {
				simRatio := job.params[paramIndex].Model.SimPriority.GetOrPanic(simType)
				value := simStats.Get(simType)
				rankArray, _ := rankInputArrays.ValuesForKeyAsSlice(paramIndex, simType)
				valueRank := slices.Index(rankArray, value)
				sumOfRanks += float64(valueRank) * simRatio * ratingPercent
			}
		}
		best.Offer(&result, sumOfRanks)
	}

	job.printer.Println("Best ranked result")
	bestMultiResult := best.GetBestOrPanic()
	job.reportSimResults_One(bestMultiResult)

	if job.writeBestToGearFiles {
		job.writeToGearFiles(bestMultiResult)
	}
}

func (job *MultiSetJob) writeToGearFiles(result simulateMultiResult) {
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		gearFile := param.GearFile

		itemSet := result.proposed.Parts[paramIndex].FullSet
		gearJson := tools.WowSimJson_Write(itemSet.Items(), &param.Model, util.PrintRecorder_Nop())

		util.WriteStringToFile(gearFile, gearJson)
	}
}
