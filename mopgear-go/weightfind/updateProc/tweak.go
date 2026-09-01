package updateProc

import (
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func (wc *choiceOutput) runFollowupTweaker(choiceName string, weight1 *weight_types.Weight1Basic) {
	weight1B := wc.runTweakJob(weight1, &wc._accuracyPrepBasic)
	accuracy1B, accuracy1BStat, _, _, _ := wc.evaluateAccuracy(&weight1B, &weight1B)
	wc.addChoice(weightChoice{
		choiceName:    choiceName + "_TWEAK_BASIC",
		weight1:       weight1B,
		weightOrig:    &weight1B,
		accuracy1:     accuracy1B,
		accuracy1Stat: accuracy1BStat,
	})

	weight1S := wc.runTweakJob(weight1, &wc._accuracyPrepStat)
	accuracy1S, accuracy1SStat, _, _, _ := wc.evaluateAccuracy(&weight1S, &weight1S)
	wc.addChoice(weightChoice{
		choiceName:    choiceName + "_TWEAK_STAT",
		weight1:       weight1S,
		weightOrig:    &weight1S,
		accuracy1:     accuracy1S,
		accuracy1Stat: accuracy1SStat,
	})
}

func (wc *choiceOutput) runTweakJob(weight1 *weight_types.Weight1Basic, evaluate *weightfind.EvaluateAccuracyPrepared) weight_types.Weight1Basic {
	wc._accuracyMutex.Lock()
	defer wc._accuracyMutex.Unlock()

	weight, _ := weightfind.WeightTweaker_FastCached(*weight1, 0.1, wc.input.statTypes, evaluate)
	return weight
}
