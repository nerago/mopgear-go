package multi

import (
	"iter"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind/weight_types"
)

type specWorking struct {
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

func (work *specWorking) Label() string {
	return work.label
}
func (work *specWorking) Model() *gear_model.SpecModel {
	return &work.model
}
func (work *specWorking) ItemOptions() *items.FullOptionsMap {
	return &work.itemOptionsWork
}
func (work *specWorking) AddSeen(equipMap *items.FullEquipMap) {
	work.seenInSolutions.Add(equipMap)
}
func (work *specWorking) AddSeenScaled(equipMap *items.FullEquipMap, scale uint32) {
	work.seenInSolutions.AddScaled(equipMap, scale)
}

func (job *MultiSetJob) prepareWorking() {
	for label, prep := range job.itemPrep {
		for _, weightType := range job.input.WeightTypeList {
			job.working.Put(label, weightType, &specWorking{
				label:                        prep.label,
				model:                        prep.model,
				seenInSolutions:              prep.seenInSolutions,
				itemOptionsWork:              prep.itemOptions.Clone(),
				expectAllBonusItemsAvailable: prep.inputs.ExpectAllBonusItemsAvailable,
				weightType:                   weightType,
			})
		}
	}

	workingChannel := util_async.SeqToChannel(job.working.SeqValues())
	util_async.ForEach_Channel(c_prepThreadCount, workingChannel, func(work *specWorking) {
		work.runBaseline(job.printer, job.input.TimeLimitEachSolve)
	})

	for _, nested := range job.working.SeqKey2NestedKey1Value() {
		job.prepareRatingMultipliersGroup(nested, job.printer)
	}
}

func (work *specWorking) runBaseline(printer *util.PrintRecorder, timeout int) {
	printer.Printf("BASELINE for %s %d\n", work.Label, work.weightType)
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

func (job *MultiSetJob) prepareRatingMultipliersGroup(nested iter.Seq2[string, *specWorking], printer *util.PrintRecorder) {
	var totalPercent float64
	for _, param := range job.input.Param {
		totalPercent += param.ItemInputs.RequestRatingPercent
	}

	if util.FloatEqualsZero(totalPercent) {
		panic("percent total is zero")
	}

	for label, work := range nested {
		param := job.input.GetSetParam(label)
		if param == nil {
			panic("param " + label + " missing")
		}

		actualRequest := param.ItemInputs.RequestRatingPercent / totalPercent
		work.prepareRatingMultiplier(actualRequest, printer)
	}
}

func (work *specWorking) prepareRatingMultiplier(requestRatingPercent float64, printer *util.PrintRecorder) {
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
