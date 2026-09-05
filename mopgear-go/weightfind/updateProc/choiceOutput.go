package updateProc

import (
	"slices"
	"sync"

	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type choiceOutput struct {
	label string

	_mutex   sync.Mutex
	_choices []weightChoice
	_summary util.StringBuild2

	input   *updateInputs
	printer *util.PrintRecorder

	_accuracyMutex     sync.Mutex
	_accuracyPrepBasic weightfind.EvaluateAccuracyPrepared
	_accuracyPrepStat  weightfind.EvaluateAccuracyPrepared
}

type weightChoice struct {
	choiceName    string
	weight1       weight_types.Weight1Basic
	weight2       *weight_types.Weight2
	weight3       *weight_types.Weight3
	hadExtended   bool
	accuracy1     float64
	accuracy1Stat float64
	accuracyX     float64
	accuracyXStat float64
	weightResult  *weight_types.WeightResultCommon
	weightOrig    weight_types.IWeight
}

func (wc *choiceOutput) startReport() {
	wc._mutex.Lock()
	defer wc._mutex.Unlock()

	wc._summary.WriteString("Weights Accuracy Summary ::::: ")
	wc._summary.WriteString(wc.label)
	wc._summary.WriteString(" ::::: ")
}

func (wc *choiceOutput) initAccuracy() {
	wc._accuracyMutex.Lock()
	defer wc._accuracyMutex.Unlock()
	wc._accuracyPrepBasic.Init(wc.input.dataAll, &wc.input.targetRatio, false, false)
	wc._accuracyPrepStat.Init(wc.input.dataAll, &wc.input.targetRatio, true, true)
}

func (wc *choiceOutput) addChoice(choice weightChoice) {
	wc._mutex.Lock()
	defer wc._mutex.Unlock()

	wc._choices = append(wc._choices, choice)

	wc._summary.WriteString(choice.choiceName)
	wc._summary.WriteString("=")
	wc._summary.WriteFloat64(choice.accuracy1, 4)
	wc._summary.WriteString(" (")
	wc._summary.WriteFloat64(choice.accuracy1Stat, 4)
	wc._summary.WriteString(") ")
}

func (wc *choiceOutput) bestWeightChoice1() (weightChoice, bool) {
	wc._mutex.Lock()
	defer wc._mutex.Unlock()

	best := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range wc._choices {
		if !choice.weight1.IsEmpty() {
			best.Offer(&choice, choice.accuracy1Stat)
			best.Offer(&choice, choice.accuracy1)
		}
	}
	return best.GetBestOptional().GetWithFlag()
}

func (wc *choiceOutput) bestWeightChoiceExtended() (util_collection.Optional[weightChoice], util_collection.Optional[weightChoice]) {
	input := wc.input
	best2 := util_rank.BestCollector1[weightChoice]{}
	best3 := util_rank.BestCollector1[weightChoice]{}

	wc._accuracyMutex.Lock()
	defer wc._accuracyMutex.Unlock()

	for _, choice := range wc.getChoicesSafeCopy() {
		weightOrig := choice.weightOrig
		switch weightCast := weightOrig.(type) {
		case *weight_types.Weight2:
			choice.weight2 = weightCast
			acc2 := wc._accuracyPrepBasic.EvaluateWeight2(weightCast)
			acc2St := wc._accuracyPrepStat.EvaluateWeight2(weightCast)
			best2.Offer(&choice, acc2)
			best2.Offer(&choice, acc2St)
		case *weight_types.Weight3:
			choice.weight3 = weightCast
			weightConvert2 := weightCast.ConvertToWeight2(input.dataAll)
			choice.weight2 = weightConvert2

			acc3 := wc._accuracyPrepBasic.EvaluateWeight3(weightCast)
			acc3St := wc._accuracyPrepStat.EvaluateWeight3(weightCast)
			best3.Offer(&choice, acc3)
			best3.Offer(&choice, acc3St)

			acc2 := wc._accuracyPrepBasic.EvaluateWeight2(weightConvert2)
			acc2St := wc._accuracyPrepStat.EvaluateWeight2(weightConvert2)
			best2.Offer(&choice, acc2)
			best2.Offer(&choice, acc2St)
		}
	}
	return best2.GetBestOptional(), best3.GetBestOptional()
}

func (wc *choiceOutput) handleWeightError(choiceName string, err error) {
	wc.printer.Printf("Weights error %s %s NULL\n", wc.label, err)
	wc.addChoice(weightChoice{choiceName: choiceName})
}

func (wc *choiceOutput) evaluateWeight1(choiceName string, weight1 *weight_types.Weight1Basic) {
	wc.evaluateWeightGeneric(choiceName, weight1, weight1, nil)
}

func (wc *choiceOutput) evaluateWeight2(choiceName string, weight2 *weight_types.Weight2) {
	wc.evaluateWeightGeneric(choiceName, weight2.ConvertToWeight1(), weight2, nil)
}

func (wc *choiceOutput) evaluateWeight3(choiceName string, weight3 *weight_types.Weight3) {
	wc.evaluateWeightGeneric(choiceName, weight3.ConvertToWeight2(wc.input.dataAll).ConvertToWeight1(), weight3, nil)
}

func (wc *choiceOutput) evaluateWeightResult1(choiceName string, weightResult *weight_types.WeightResult1) {
	wc.evaluateWeightGeneric(choiceName, weightResult.Weight, weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (wc *choiceOutput) evaluateWeightResult2(choiceName string, weightResult *weight_types.WeightResult2) {
	wc.evaluateWeightGeneric(choiceName, weightResult.AsWeight1(wc.input.dataAll), weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (wc *choiceOutput) evaluateWeightResult3(choiceName string, weightResult *weight_types.WeightResult3) {
	wc.evaluateWeightGeneric(choiceName, weightResult.AsWeight1(wc.input.dataAll), weightResult.WeightInterface, &weightResult.WeightResultCommon)
}

func (wc *choiceOutput) evaluateWeightResultFuture1(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult1]) {
	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
		wc.evaluateWeightResult1(choiceName, &weightResult)
	}
}

func (wc *choiceOutput) evaluateWeightResultFuture2(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult2]) {
	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
		wc.evaluateWeightResult2(choiceName, &weightResult)
	}
}

//func (wc *weightChoiceCollection) evaluateWeightResultFuture3(choiceName string, futureResult *util_async.FutureCancellable[weight_types.WeightResult3]) {
//	if weightResult, hasResult := futureResult.WaitForResult(); hasResult {
//		spec.evaluateWeightResult3(choiceName, &weightResult)
//	}
//}

func (wc *choiceOutput) evaluateWeightGeneric(choiceName string, weight1 *weight_types.Weight1Basic, weightOrig weight_types.IWeight, weightResult *weight_types.WeightResultCommon) {
	if weight1 == nil || weightOrig == nil {
		wc.printer.Printf("Weights accuracy %s %s NIL\n", wc.label, choiceName)
		wc.addChoice(weightChoice{choiceName: choiceName, weightResult: weightResult})
		return
	}

	accuracy1, accuracy1Stat, accuracyX, accuracyXStat, hasExtended := wc.evaluateAccuracy(weightOrig, weight1)
	wc.addChoice(weightChoice{choiceName, *weight1, nil, nil, hasExtended,
		accuracy1, accuracy1Stat, accuracyX, accuracyXStat,
		weightResult, weightOrig})

	isGood := wc.logChoice(choiceName, weight1, weightOrig, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
	if isGood {
		wc.runFollowupTweaker(choiceName, weight1)
	}
}

func (wc *choiceOutput) logChoice(choiceName string, weight1 *weight_types.Weight1Basic, weightOrig weight_types.IWeight, accuracy1 float64, accuracy1Stat float64, accuracyX float64, accuracyXStat float64) bool {
	tools.WritePawnString(*weight1, wc.printer)
	wc.printer.Println(weightOrig.String())
	tools.WriteWeightString(weightOrig, wc.printer)

	isGood := false
	if weight1.IsEmpty() || weightOrig.IsEmpty() {
		wc.printer.Printf("Weights accuracy %s %s EMPTY a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
	} else if weight1.IsOverlySimple() {
		wc.printer.Printf("Weights accuracy %s %s OVERLY SIMPLE a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
	} else {
		wc.printer.Printf("Weights accuracy %s %s a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		isGood = true
	}

	return isGood
}

func (wc *choiceOutput) evaluateAccuracy(weightOrig weight_types.IWeight, weight1 *weight_types.Weight1Basic) (accuracy1, accuracy1Stat, accuracyX, accuracyXStat float64, hasExtended bool) {
	wc._accuracyMutex.Lock()
	defer wc._accuracyMutex.Unlock()

	if _, isOne := weightOrig.(*weight_types.Weight1Basic); isOne {
		hasExtended = false
	} else if two, isTwo := weightOrig.(*weight_types.Weight2); isTwo {
		accuracyX = wc._accuracyPrepBasic.EvaluateWeight2(two)
		accuracyXStat = wc._accuracyPrepStat.EvaluateWeight2(two)
		hasExtended = true
	} else if three, isThree := weightOrig.(*weight_types.Weight3); isThree {
		accuracyX = wc._accuracyPrepBasic.EvaluateWeight3(three)
		accuracyXStat = wc._accuracyPrepStat.EvaluateWeight3(three)
		hasExtended = true
	}
	accuracy1 = wc._accuracyPrepBasic.EvaluateWeight1(weight1)
	accuracy1Stat = wc._accuracyPrepStat.EvaluateWeight1(weight1)
	return accuracy1, accuracy1Stat, accuracyX, accuracyXStat, hasExtended
}

func (wc *choiceOutput) getSummary() util.StringBuild2 {
	wc._mutex.Lock()
	defer wc._mutex.Unlock()
	return wc._summary.Clone()
}

func (wc *choiceOutput) getChoicesSafeCopy() []weightChoice {
	wc._mutex.Lock()
	defer wc._mutex.Unlock()
	return slices.Clone(wc._choices)
}
