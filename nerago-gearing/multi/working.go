package multi

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/weightfind/weight_types"
)

type specWorking struct {
	itemPrep        *specItemPrep // todo break link
	itemOptionsWork items.FullOptionsMap
	weightType      weight_types.WeightType
	baselineResult  solver.SolveOutput
	ratingMultiply  float64
}

func (job *MultiSetJob) prepareWorking() {
	for label, prep := range job.itemPrep {
		for _, weightType := range job.input.WeightTypeList {
			job.working.Put(label, weightType, &specWorking{
				itemPrep:   prep,
				weightType: weightType,
			})
		}
	}

	workingChannel := util_async.SeqToChannel(job.working.SeqValues())
	util_async.ForEach_Channel(c_prepThreadCount, workingChannel, func(work *specWorking) {
		work.runBaseline(job.printer)
	})

	for _, nested := range job.working.SeqKey2NestedKey1Value() {
		job.prepareRatingMultipliersGroup(nested, job.printer)
	}
}

func (work *specWorking) runBaseline(printer *util.PrintRecorder) {
	printer.Printf("BASELINE for %s %d\n", work.itemPrep.label, work.weightType)
	work.baselineResult = solver.Solver(solver.SolveInput{
		ItemOptions: &work.itemPrep.itemOptions,
		Model:       &work.itemPrep.model,
		WeightType:  work.weightType,
		Printer:     printer})

	if !work.baselineResult.Success {
		panic("failed to find baseline for " + work.itemPrep.label)
	}
	work.baselineResult.Report(printer)
	work.itemPrep.seenInSolutions.Add(work.baselineResult.FullSet.Items())
}

func (job *MultiSetJob) prepareRatingMultipliersGroup(nested iter.Seq2[string, *specWorking], printer *util.PrintRecorder) {
	var totalPercent float64
	for _, work := range nested {
		param := work.itemPrep.inputs
		totalPercent += param.RequestRatingPercent
	}

	if util.FloatEqualsZero(totalPercent) {
		panic("percent total is zero")
	}

	for _, work := range nested {
		param := work.itemPrep.inputs
		actualRequest := param.RequestRatingPercent / totalPercent
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
		work.itemPrep.label, work.baselineResult.ResultRating, work.ratingMultiply,
		baselineRating*work.ratingMultiply,
		baselineRating*work.ratingMultiply/targetCombined*100,
	)
}
