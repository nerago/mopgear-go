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
	mutex    sync.Mutex
	label    string
	_choices []weightChoice
	_summary util.StringBuild2
	input    *updateInputs
	printer  *util.PrintRecorder
}

type weightChoice struct {
	choiceName    string
	weight        weight_types.Weight1Basic
	weight2       *weight_types.Weight2Extended
	weight3       *weight_types.Weight3ExtendedRanged
	hadExtended   bool
	accuracy1     float64
	accuracy1Stat float64
	accuracyX     float64
	accuracyXStat float64
	pawnString    string
	weightResult  *weight_types.WeightResultCommon
	weightOrig    weight_types.IWeight
}

func (wc *choiceOutput) startReport() {
	wc._summary.WriteString("Weights Accuracy Summary ::::: ")
	wc._summary.WriteString(wc.label)
	wc._summary.WriteString(" ::::: ")
}

func (wc *choiceOutput) addChoice(choice weightChoice) {
	wc._choices = append(wc._choices, choice)
	wc.addToSummary(choice)
}

func (wc *choiceOutput) addToSummary(option weightChoice) {
	wc._summary.WriteString(option.choiceName)
	wc._summary.WriteString("=")
	wc._summary.WriteFloat64(option.accuracy1, 4)
	wc._summary.WriteString(" (")
	wc._summary.WriteFloat64(option.accuracy1Stat, 4)
	wc._summary.WriteString(") ")
}

func (wc *choiceOutput) bestWeightChoice1() (weightChoice, bool) {
	best := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range wc._choices {
		best.Offer(&choice, choice.accuracy1Stat)
		best.Offer(&choice, choice.accuracy1)
	}
	return best.GetBestOptional().GetWithFlag()
}

func (wc *choiceOutput) bestWeightChoiceExtended() (util_collection.Optional[weightChoice], util_collection.Optional[weightChoice]) {
	input := wc.input
	best2 := util_rank.BestCollector1[weightChoice]{}
	best3 := util_rank.BestCollector1[weightChoice]{}
	for _, choice := range wc._choices {
		weightOrig := choice.weightOrig
		switch weightCast := weightOrig.(type) {
		case *weight_types.Weight2Extended:
			choice.weight2 = weightCast
			acc2 := weightfind.EvaluateAccuracyBasic(weightCast, input.simTypes, &input.targetRatio, input.dataAll)
			acc2St := weightfind.EvaluateAccuracyStatisticalExtended(weightCast, input.simTypes, &input.targetRatio, input.dataAll)
			best2.Offer(&choice, acc2)
			best2.Offer(&choice, acc2St)
		case *weight_types.Weight3ExtendedRanged:
			choice.weight3 = weightCast
			weightConvert2 := weightCast.ConvertToWeight2(input.dataAll)
			choice.weight2 = weightConvert2

			acc3 := weightfind.EvaluateAccuracyBasic(weightCast, input.simTypes, &input.targetRatio, input.dataAll)
			acc3St := weightfind.EvaluateAccuracyStatisticalExtended(weightCast, input.simTypes, &input.targetRatio, input.dataAll)
			best3.Offer(&choice, acc3)
			best3.Offer(&choice, acc3St)

			acc2 := weightfind.EvaluateAccuracyBasic(weightConvert2, input.simTypes, &input.targetRatio, input.dataAll)
			acc2St := weightfind.EvaluateAccuracyStatisticalExtended(weightConvert2, input.simTypes, &input.targetRatio, input.dataAll)
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

func (wc *choiceOutput) evaluateWeight2(choiceName string, weight2 *weight_types.Weight2Extended) {
	wc.evaluateWeightGeneric(choiceName, weight2.ConvertToWeight1(), weight2, nil)
}

func (wc *choiceOutput) evaluateWeight3(choiceName string, weight3 *weight_types.Weight3ExtendedRanged) {
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
		wc.printer.Printf("Weights accuracy %s %s NULL\n", wc.label, choiceName)
		wc.addChoice(weightChoice{choiceName: choiceName, weightResult: weightResult})
		return
	}

	input := wc.input

	var accuracyX, accuracyXStat float64
	var hadExtended bool
	if _, isOne := weightOrig.(*weight_types.Weight1Basic); isOne {
		hadExtended = false
	} else {
		accuracyX = weightfind.EvaluateAccuracyBasic(weightOrig, input.simTypes, &input.targetRatio, input.dataAll)
		accuracyXStat = weightfind.EvaluateAccuracyStatisticalExtended(weightOrig, input.simTypes, &input.targetRatio, input.dataAll)
		hadExtended = true
	}
	accuracy1 := weightfind.EvaluateAccuracyBasic(weight1, input.simTypes, &input.targetRatio, input.dataAll)
	accuracy1Stat := weightfind.EvaluateAccuracyStatisticalExtended(weight1, input.simTypes, &input.targetRatio, input.dataAll)

	pawnString := tools.WritePawnString(*weight1, wc.printer)
	wc.printer.Println(weightOrig.String())
	tools.WriteWeightString(weightOrig, wc.printer)

	if weight1.IsEmpty() || weightOrig.IsEmpty() {
		wc.printer.Printf("Weights accuracy %s %s EMPTY a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		wc.addChoice(weightChoice{choiceName: choiceName, weight: *weight1, pawnString: pawnString, weightResult: weightResult})
	} else if weight1.IsOverlySimple() {
		wc.printer.Printf("Weights accuracy %s %s OVERLY SIMPLE a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		wc.addChoice(weightChoice{choiceName: choiceName, weight: *weight1, pawnString: pawnString, weightResult: weightResult})
	} else {
		wc.printer.Printf("Weights accuracy %s %s a1=%f a1s=%f aX=%f aXs=%f\n", wc.label, choiceName, accuracy1, accuracy1Stat, accuracyX, accuracyXStat)
		wc.addChoice(weightChoice{choiceName, *weight1, nil, nil, hadExtended,
			accuracy1, accuracy1Stat,
			accuracyX, accuracyXStat,
			pawnString, weightResult, weightOrig})
	}
}

func (wc *choiceOutput) getSummary() util.StringBuild2 {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()
	return wc._summary.Clone()
}

func (wc *choiceOutput) getChoices() []weightChoice {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()
	return slices.Clone(wc._choices)
}

func (spec *weightSpecInternal) tweakEachWeight() {
	currentChoiceSize := len(spec.choices)
	for i := range currentChoiceSize {
		choice := spec.choices[i]
		if !choice.weight.IsEmpty() {
			weightsTweaked, _ := weightfind.WeightTweakerWithLogging(choice.weight, spec.statTypes, &spec.targetRatio, spec.dataAll, spec.process.printer)
			spec.evaluateWeight1(choice.choiceName+"_TWEAK", &weightsTweaked)
		}
	}
}
