package multi

import (
	"fmt"
	"maps"

	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/solver"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type workingGroup struct {
	job        *MainJob
	task       *multi_types.JobInputTask
	workers    map[string]*specWorker
	weightType weight_types.WeightType
}

type specWorker struct {
	// copied from specItemPrep
	label                        string
	model                        *gear_model.SpecModel
	task                         *multi_types.JobInputTask
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
	return work.model
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

func (work *specWorker) init(printer *util.PrintRecorder) error {
	return work.setupAlternateGems(work.task.AlternateGemList, printer)
}

func (job *MainJob) prepareWorkingGroups(cancel util_async.CancelSignal) (<-chan *workingGroup, *util_async.PossibleFutureErrors, error) {
	job.workGroups = make([]*workingGroup, 0)
	for task := range util_collection.ForPointer(job.tasks) {
		for _, weightType := range task.WeightTypeList {
			group := &workingGroup{
				job:        job,
				task:       task,
				weightType: weightType,
				workers:    make(map[string]*specWorker),
			}

			for label, prep := range job.itemPrep {
				work := &specWorker{
					label:                        prep.label,
					model:                        prep.model,
					seenInSolutions:              prep.seenInSolutions,
					itemOptionsWork:              prep.itemOptions.Clone(),
					expectAllBonusItemsAvailable: prep.inputs.ExpectAllBonusItemsAvailable,
					weightType:                   weightType,
					task:                         task,
				}
				if err := work.init(job.printer); err != nil {
					return nil, nil, err
				}
				group.workers[label] = work
			}
			job.workGroups = append(job.workGroups, group)
		}
	}

	prepErrors := util_async.PossibleFutureErrorMake()
	groupChannel := util_async.MapOptional_SliceToChannel(c_prepThreadCount, job.workGroups, func(group **workingGroup) (*workingGroup, bool) {
		if err := (*group).prepareWorkersBaseline(&job.input, job.printer); err != nil {
			prepErrors.AddError(err)
			return nil, false
		}

		if (*group).task.RunDecimate {
			err := (*group).runDecimate(cancel)
			if err != nil {
				prepErrors.AddError(err)
				return nil, false
			}
		}
		return *group, true
	})
	return groupChannel, prepErrors, nil
}

func (group *workingGroup) prepareWorkersBaseline(input *multi_types.JobInputs, printer *util.PrintRecorder) error {
	workChannel := util_async.SeqToChannel(maps.Values(group.workers))

	err := util_async.ForEach_Channel_PassError(c_prepThreadCount, workChannel, func(work *specWorker) error {
		return work.runBaseline(printer, input.TimeLimitEachSolve, nil)
	})
	if err != nil {
		return err
	}

	group.prepareRatingMultipliersGroup(input, printer)

	return nil
}

func (work *specWorker) runBaseline(printer *util.PrintRecorder, timeout int, cancel util_async.CancelSignal) error {
	printer.Printf("BASELINE for %s %d\n", work.Label(), work.weightType)
	work.baselineResult = solver.Solver(
		work.ItemOptions(),
		work.Model(),
		printer,
		work.weightType,
		timeout,
		cancel,
	)

	if work.baselineResult.Error != nil {
		return fmt.Errorf("failed to find baseline for %s: %w", work.Label(), work.baselineResult.Error)
	} else if !work.baselineResult.Success {
		return fmt.Errorf("failed to find baseline for %s: unknown", work.Label())
	}

	work.baselineResult.Report(printer)
	work.AddSeen(work.baselineResult.FullSet.Items())
	return nil
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
