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
	__itemPrep       *specItemPrep // todo break link
	_label           string
	_model           gear_model.SpecModel
	_itemOptionsWork items.FullOptionsMap
	weightType       weight_types.WeightType
	baselineResult   solver.SolveOutput
	ratingMultiply   float64
}

func (work *specWorking) label() string {
	return work._label
}
func (work *specWorking) model() *gear_model.SpecModel {
	return &work._model
}
func (work *specWorking) itemOptions() *items.FullOptionsMap {
	return &work._itemOptionsWork
}
func (work *specWorking) requestRatingPercent() float64 {
	return work.__itemPrep.inputs.RequestRatingPercent
}
func (work *specWorking) addSeen(equipMap *items.FullEquipMap) {
	work.__itemPrep.seenInSolutions.Add(equipMap)
}
func (work *specWorking) addSeenScaled(equipMap *items.FullEquipMap, scale uint32) {
	work.__itemPrep.seenInSolutions.AddScaled(equipMap, scale)
}

func (job *MultiSetJob) prepareWorking() {
	for label, prep := range job.itemPrep {
		for _, weightType := range job.input.WeightTypeList {
			job.working.Put(label, weightType, &specWorking{
				__itemPrep:       prep,
				_label:           prep.label,
				_model:           prep.model,
				_itemOptionsWork: prep.itemOptions.Clone(),
				weightType:       weightType,
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
	printer.Printf("BASELINE for %s %d\n", work.label, work.weightType)
	work.baselineResult = solver.Solver(
		work.itemOptions(),
		work.model(),
		printer,
		work.weightType,
		timeout,
	)

	if !work.baselineResult.Success {
		panic("failed to find baseline for " + work.label())
	}
	work.baselineResult.Report(printer)
	work.addSeen(work.baselineResult.FullSet.Items())
}

func (job *MultiSetJob) prepareRatingMultipliersGroup(nested iter.Seq2[string, *specWorking], printer *util.PrintRecorder) {
	var totalPercent float64
	for _, work := range nested {
		totalPercent += work.requestRatingPercent()
	}

	if util.FloatEqualsZero(totalPercent) {
		panic("percent total is zero")
	}

	for _, work := range nested {
		actualRequest := work.requestRatingPercent() / totalPercent
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
		work.label(), work.baselineResult.ResultRating, work.ratingMultiply,
		baselineRating*work.ratingMultiply,
		baselineRating*work.ratingMultiply/targetCombined*100,
	)
}
