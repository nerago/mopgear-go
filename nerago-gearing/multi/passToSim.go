package multi

import (
	"cmp"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

func checkNoConflicts(outputSet []multi_types.SingleProposedOutput, printer *util.PrintRecorder) bool {
	itemById := make(map[items.ItemId]*items.FullItem)
	for outputIndex := range outputSet {
		for item := range outputSet[outputIndex].FullSet.Items().AllItemSeq() {
			existing, found := itemById[item.ItemId()]
			if !found {
				itemById[item.ItemId()] = item
			} else if !existing.Equals(item) {
				printer.Printf("!! CONFLICT %s\n!!          %s\n", item.CreateString(), existing.CreateString())
				return false
			}
		}
	}
	return true
}

func (job *MultiSetJob) existingGearAsProposal() multi_types.MultiProposedOutput {
	proposal := multi_types.MultiProposedOutput{Id: "00000000-0000-4000-8000-000000000000"}
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		single := multi_types.SingleProposed_FromEquip(param.exactEquippedGear, &param.MultiSetParam)
		proposal.Parts = append(proposal.Parts, single)
		proposal.TotalRatingSum += single.ResultRating
	}
	proposal.Combo = multi_types.CommonCombo_FromProposed(proposal.Parts)
	return proposal
}

type simulateJob struct {
	spec        stats.SpecType
	goal        stats.OptimiseGoal
	fight       stats.WowSim_Fight
	equip       items.FullEquipMap
	professions model.ProfessionInfo
}

type simulateJobResult struct {
	job    simulateJob
	result simulate.SimData
}

func (simJob *simulateJob) Equals(other *simulateJob) bool {
	return simJob.spec == other.spec && simJob.fight == other.fight && simJob.equip.Equals(&other.equip) && simJob.professions == other.professions
}

type simulateMultiResult struct {
	proposed multi_types.MultiProposedOutput
	result   []simulate.SimData
}

func (job *MultiSetJob) prepareSimList(proposalList <-chan multi_types.MultiProposedOutput) <-chan simulateJob {
	jobChannel := channel_op.Map_ChannelToChannel(2, proposalList, func(proposal multi_types.MultiProposedOutput, next chan<- simulateJob) {
		for _, output := range proposal.Parts {
			job := simulateJob{output.Spec, output.Model.Goal, output.Model.SimulateAs, *output.FullSet.Items(), output.Model.Professions}
			next <- job
		}
	})

	return util.RemoveDuplicatesFunc_Channels(jobChannel, (*simulateJob).Equals)
}

func (job *MultiSetJob) runSims(jobChan <-chan simulateJob, trackProgress *util.TrackProgress) []simulateJobResult {
	// job.printer.Printf("@@@@@@@@@@ RUN SIM JOBS %d @@@@@@@@@@\n", len(jobList))
	// trackProgress.RunOuterTracking(len(jobList))
	defer trackProgress.Stop()

	return channel_op.Map_ChannelToSlice(simThreadCount, jobChan, func(sim simulateJob, resultChan chan<- simulateJobResult) {
		result := simulate.WowSim_Execute_SpecifyAll(job.simRunSize, sim.spec, sim.goal, sim.fight, sim.professions, &sim.equip, nil, trackProgress.MakeNested())
		job.printer.Printf("sim %22s fight=%d %s\n", sim.spec.Name(), sim.fight, result.CompactStringGeneral())
		resultChan <- simulateJobResult{sim, result}
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
	multiResult := simulateMultiResult{proposal, make([]simulate.SimData, len(proposal.Parts))}
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
	result.proposed.Combo.Print(job.printer)

	for specIndex, specResult := range result.result {
		param := &job.params[specIndex]
		job.printer.Printf("\n---------------- %s ----------------\n", param.Label)

		output := result.proposed.Parts[specIndex]
		output.Report(job.printer)
		specResult.Print(job.printer)

		for slot, itemId := range param.ReportVariant {
			variantEquip := *output.FullSet.Items()
			variantItem := job.findVariantItem(result, itemId, param)
			variantEquip[slot] = variantItem
			job.printer.Printf("\n---------------- %s %s ----------------\n", param.Label, variantItem.BaseName())
			tools.WowSimJson_Write(&variantEquip, &param.Model, job.printer)
			job.printer.Println0()
		}
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

	outputTypes := []simulate.SimType{simulate.Sim_DPS, simulate.Sim_DTPS, simulate.Sim_TMI, simulate.Sim_DEATH}

	csv := util.CSVOutputByColumn{}
	csv.InitRows(len(job.params)*len(outputTypes) + 1)
	csv.AddString("id")
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		for _, resultType := range outputTypes {
			csv.AddToBuilder(func(b *util.StringBuild2) {
				b.WriteString(resultType.Name())
				b.WriteString(" (")
				b.WriteString(param.Label)
				b.WriteRune(')')
			})
		}
	}
	csv.FinishColumn()

	for _, simResult := range simResultList {
		csv.AddString(simResult.proposed.Id)

		for _, resultStat := range simResult.result {
			for _, resultType := range outputTypes {
				value := resultStat.Get(resultType)
				csv.AddFloat64(value, -1)
			}
		}

		csv.FinishColumn()
	}

	csv.Write(job.printer)
}

func (job *MultiSetJob) suggestResultFromRankings(results []simulateMultiResult) {
	// type rankable struct {
	// 	result        *simulateMultiResult
	// 	simRankDetail map[simulate.SimType]int
	// 	simScore      float64
	// }
	// entries := util.MapSliceAsNew(results, func(result *simulateMultiResult) rankable {
	// 	return rankable{
	// 		result:        result,
	// 		simRankDetail: make(map[simulate.SimType]int),
	// 	}
	// })

	// // score each sim
	// for _, simType := range stathighs.G_RequiredSims {
	// 	for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), entries, func(x *rankable) float64 { return x.input.SimResult.Get(simType) }) {
	// 		entry.simRankDetail[simType] = simDetailRank
	// 		entry.combinedSimRankScore += float64(simDetailRank) * simRatios.Get(simType)
	// 	}
	// }

	simResultTypeList := []simulate.SimType{simulate.Sim_DPS, simulate.Sim_DTPS, simulate.Sim_TMI, simulate.Sim_DEATH}
	rankInputArrays := util.MapMapSlice[int, simulate.SimType, float64]{}
	for _, result := range results {
		for paramIndex, simStats := range result.result {
			for _, simType := range simResultTypeList {
				value := simStats.Get(simType)
				rankInputArrays.Add(paramIndex, simType, value)
			}
		}
	}

	for paramIndex := range job.params {
		for _, simType := range simResultTypeList {
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
		sumOfRanks := 0
		for paramIndex, simStats := range result.result {
			for _, simType := range simResultTypeList {
				value := simStats.Get(simType)
				rankArray, _ := rankInputArrays.ValuesForKeyAsSlice(paramIndex, simType)
				valueRank := slices.Index(rankArray, value)
				sumOfRanks += valueRank
			}
		}
		best.Offer(&result, float64(sumOfRanks))
	}

	job.printer.Println("Best ranked result")
	job.reportSimResults_One(best.GetBestOrPanic())
}
