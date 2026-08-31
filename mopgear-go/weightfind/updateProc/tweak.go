package updateProc

import (
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func (wc *choiceOutput) runFollowupTweaker(choiceName string, weight1 *weight_types.Weight1Basic) {
	weight1Tweaked, _ := weightfind.WeightTweakerWithLogging(*weight1,
		wc.input.statTypes, &wc.input.targetRatio, wc.input.dataAll, wc.printer)

	accuracy1, accuracy1Stat, _, _, _ := wc.evaluateAccuracy(&weight1Tweaked, &weight1Tweaked)
	wc.addChoice(weightChoice{
		choiceName:    choiceName + "_TWEAK",
		weight1:       weight1Tweaked,
		accuracy1:     accuracy1,
		accuracy1Stat: accuracy1Stat,
		weightOrig:    &weight1Tweaked,
	})
}
