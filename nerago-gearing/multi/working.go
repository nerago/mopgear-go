package multi

import (
	"maps"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind/weight_types"
)

type workingGroup struct {
	job        *MultiSetJob
	workers    map[string]*specWorker
	weightType weight_types.WeightType
}

type specWorker struct {
	// copied from specItemPrep
	label                        string
	model                        gear_model.SpecModel
	expectAllBonusItemsAvailable bool
	seenInSolutions              *seenMap

	// actual working vars
	itemOptionsWork items.FullOptionsMap
	weightType      weight_types.WeightType
	baselineResult  solver.SolveOutput
	ratingMultiply  float64
}

func (work *specWorker) Label() string {
	return work.label
}
func (work *specWorker) Model() *gear_model.SpecModel {
	return &work.model
}
func (work *specWorker) ItemOptions() *items.FullOptionsMap {
	return &work.itemOptionsWork
}
func (work *specWorker) AddSeen(equipMap *items.FullEquipMap) {
	work.seenInSolutions.Add(equipMap)
}
func (work *specWorker) AddSeenScaled(equipMap *items.FullEquipMap, scale uint32) {
	work.seenInSolutions.AddScaled(equipMap, scale)
}

func (job *MultiSetJob) prepareWorkingGroups() <-chan *workingGroup {
	job.workGroups = make(map[weight_types.WeightType]*workingGroup)
	for _, weightType := range job.input.WeightTypeList {
		group := &workingGroup{
			workers:    make(map[string]*specWorker),
			weightType: weightType,
		}

		for label, prep := range job.itemPrep {
			group.workers[label] = &specWorker{
				label:                        prep.label,
				model:                        prep.model,
				seenInSolutions:              prep.seenInSolutions,
				itemOptionsWork:              prep.itemOptions.Clone(),
				expectAllBonusItemsAvailable: prep.inputs.ExpectAllBonusItemsAvailable,
				weightType:                   weightType,
			}
		}
		job.workGroups[weightType] = group
	}

	groupChannel := util_async.SeqToChannel(maps.Values(job.workGroups))
	preparedChannel := util_async.Map_ChannelToChannel(c_prepThreadCount, groupChannel, func(group *workingGroup) *workingGroup {
		group.prepareWorkers(&job.input, job.printer)
		return group
	})

	if job.input.RunDecimate {
		return job.runDecimate(preparedChannel)
	} else {
		return preparedChannel
	}
}

func (group *workingGroup) prepareWorkers(input *multi_types.JobInputs, printer *util.PrintRecorder) {
	workChannel := util_async.SeqToChannel(maps.Values(group.workers))
	util_async.ForEach_Channel(c_prepThreadCount, workChannel, func(work *specWorker) {
		work.runBaseline(printer, input.TimeLimitEachSolve)
	})
	group.prepareRatingMultipliersGroup(input, printer)
}

func (work *specWorker) runBaseline(printer *util.PrintRecorder, timeout int) {
	printer.Printf("BASELINE for %s %d\n", work.Label(), work.weightType)
	work.baselineResult = solver.Solver(
		work.ItemOptions(),
		work.Model(),
		printer,
		work.weightType,
		timeout,
	)

	if !work.baselineResult.Success {
		panic("failed to find baseline for " + work.Label())
	}
	work.baselineResult.Report(printer)
	work.AddSeen(work.baselineResult.FullSet.Items())
}

func (group *workingGroup) prepareRatingMultipliersGroup(input *multi_types.JobInputs, printer *util.PrintRecorder) {
	var totalPercent float64
	for _, param := range input.Param {
		totalPercent += param.ItemInputs.RequestRatingPercent
	}
	if util.FloatEqualsZero(totalPercent) {
		panic("percent total is zero")
	}

	for label, work := range group.workers {
		param := input.GetSetParam(label)
		if param == nil {
			panic("param " + label + " missing")
		}

		actualRequest := param.ItemInputs.RequestRatingPercent / totalPercent
		work.prepareRatingMultiplier(actualRequest, printer)
	}
}

func (work *specWorker) prepareRatingMultiplier(requestRatingPercent float64, printer *util.PrintRecorder) {
	const targetCombined float64 = 10.0
	baselineRating := work.baselineResult.ResultRating

	targetForThis := targetCombined * requestRatingPercent
	multiplyRatingsBy := targetForThis / baselineRating
	work.ratingMultiply = multiplyRatingsBy

	printer.Printf("MULTIPLIERS %s base=%e mult=%e value=%e percent=%.2f\n",
		work.Label(), work.baselineResult.ResultRating, work.ratingMultiply,
		baselineRating*work.ratingMultiply,
		baselineRating*work.ratingMultiply/targetCombined*100,
	)
}
