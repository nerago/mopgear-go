package fitting1

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	util_weight2 "github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type FittingSingleStatSegmentsProcess struct {
	printer *util.PrintRecorder

	timeout                  int
	onlyComputeSingleSegment bool

	samplesOriginal       []util_weight2.FittingSample
	samplesRemainingParts map[weight_types.StatRangeFloat][]util_weight2.FittingSample

	foundSegments map[weight_types.StatRangeFloat]FittingSingleStatResult
}

func (fg *FittingSingleStatSegmentsProcess) Init(printer *util.PrintRecorder, timeout int) {
	fg.printer = printer
	fg.timeout = timeout
	fg.foundSegments = make(map[weight_types.StatRangeFloat]FittingSingleStatResult)
	fg.samplesRemainingParts = make(map[weight_types.StatRangeFloat][]util_weight2.FittingSample)
}

func (fg *FittingSingleStatSegmentsProcess) SetOnlyComputeSingleSegment(lazy bool) {
	fg.onlyComputeSingleSegment = lazy
}

func (fg *FittingSingleStatSegmentsProcess) SupplyData(inputData []util_weight2.FittingSample) {
	fg.samplesOriginal = slices.Clone(inputData)
}

// let's say a standard output is about 4 line segments
// each segment thus should cover on average 25%
// initial range is a stronger requirement of at least 35%, assume remaining actually 60% (20% each)
// next ones we want to give them some slack but hoping for 15-30%
func (fg *FittingSingleStatSegmentsProcess) Run(cancel util_async.CancelSignal) (map[weight_types.StatRangeFloat]FittingSingleStatResult, error) {
	err := fg.runInitial(cancel)
	if err != nil {
		return nil, err
	}

	if fg.onlyComputeSingleSegment {
		return fg.foundSegments, nil
	}

	overallSize := len(fg.samplesOriginal)
	for len(fg.samplesRemainingParts) > 0 {
		fg.mergeAnyPossibleRemainingSamples()

		nextRange, nextData := util_collection.MapFirstEntry(fg.samplesRemainingParts)
		delete(fg.samplesRemainingParts, nextRange)

		processRatioOfOverall := float64(len(nextData)) / float64(overallSize)
		targetInclude := 0.0

		if processRatioOfOverall < c_fitting_dropped_range_threshold || len(nextData) <= 8 {
			// drop it, can't get good results from small sample
			continue
		} else if processRatioOfOverall >= c_fitting_large_range_threshold {
			targetRatioOfOverall := c_fitting_large_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_medium_range_threshold {
			targetRatioOfOverall := c_fitting_medium_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_small_range_threshold {
			targetRatioOfOverall := c_fitting_small_range_required
			targetInclude = targetRatioOfOverall / processRatioOfOverall
		} else if processRatioOfOverall >= c_fitting_tiny_range_threshold {
			targetInclude = c_fitting_tiny_range_over_required
		} else {
			targetInclude = c_fitting_tiny_range_under_required
		}

		err = fg.runNextSegment(nextData, nextRange, targetInclude, cancel)
		if err != nil {
			return fg.foundSegments, err
		}
	}

	return fg.foundSegments, nil
}

func (fg *FittingSingleStatSegmentsProcess) mergeAnyPossibleRemainingSamples() {
	type statRangeFlagged struct {
		statRange weight_types.StatRangeFloat
		remain    bool
	}
	flaggedSegments := make([]statRangeFlagged, 0)
	for key := range fg.samplesRemainingParts {
		flaggedSegments = append(flaggedSegments, statRangeFlagged{
			key,
			true,
		})
	}
	for key := range fg.foundSegments {
		flaggedSegments = append(flaggedSegments, statRangeFlagged{
			key,
			false,
		})
	}

	slices.SortFunc(flaggedSegments, func(a, b statRangeFlagged) int { return cmp.Compare(a.statRange.Minimum, b.statRange.Minimum) })
	for i := range len(flaggedSegments) - 1 {
		a := flaggedSegments[i]
		b := flaggedSegments[i+1]
		if a.remain && b.remain {
			combinedData := slices.Concat(fg.samplesRemainingParts[a.statRange], fg.samplesRemainingParts[b.statRange])
			combinedRange := weight_types.StatRangeFloat{
				Minimum: min(a.statRange.Minimum, b.statRange.Minimum),
				Maximum: max(a.statRange.Maximum, b.statRange.Maximum),
			}

			delete(fg.samplesRemainingParts, a.statRange)
			delete(fg.samplesRemainingParts, b.statRange)
			fg.samplesRemainingParts[combinedRange] = combinedData
		}
	}
}

func (fg *FittingSingleStatSegmentsProcess) runInitial(cancel util_async.CancelSignal) error {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fg.printer, fg.timeout)
	fit.SetMinimumIncludeRate(c_fitting_initial_range_required)
	fit.SupplySamples(fg.samplesOriginal)

	resultOptionalFuture, err := fit.Run()
	if err != nil {
		return err
	}
	err = util_async.ChainCancel(cancel, resultOptionalFuture)
	if err != nil {
		return errors.Join(err, resultOptionalFuture.Cancel())
	}

	resultOptional := resultOptionalFuture.WaitForResultAsOptional()
	if segmentResult, hasResult := resultOptional.GetWithFlag(); hasResult {
		statRange := weight_types.StatRangeFloat{Minimum: segmentResult.Minimum, Maximum: segmentResult.Maximum}
		segmentResult.BuiltSequence = 0
		fg.foundSegments[statRange] = segmentResult

		totalRange := weight_types.StatRangeFloat{Minimum: 0, Maximum: c_fitting_statScaledRangeHigh}
		fg.addToRemainingData(fg.samplesOriginal, totalRange, statRange)
		return nil
	} else {
		return errors.New("ERROR failed to get any useful stat fit")
	}
}

func (fg *FittingSingleStatSegmentsProcess) runNextSegment(inputData []util_weight2.FittingSample, inputRange weight_types.StatRangeFloat, includeRate float64, cancel util_async.CancelSignal) error {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fg.printer, fg.timeout)
	fit.SetMinimumIncludeRate(includeRate)
	fit.SupplySamples(inputData)

	resultOptionalFuture, err := fit.Run()
	if err != nil {
		return err
	}
	err = util_async.ChainCancel(cancel, resultOptionalFuture)
	if err != nil {
		return errors.Join(err, resultOptionalFuture.Cancel())
	}

	resultOptional := resultOptionalFuture.WaitForResultAsOptional()
	if segmentResult, hasResult := resultOptional.GetWithFlag(); hasResult {
		minimum := max(inputRange.Minimum, segmentResult.Minimum)
		maximum := min(inputRange.Maximum, segmentResult.Maximum)

		statRange := weight_types.StatRangeFloat{Minimum: minimum, Maximum: maximum}
		segmentResult.Minimum = minimum
		segmentResult.Maximum = maximum
		segmentResult.BuiltSequence = len(fg.foundSegments)
		fg.foundSegments[statRange] = segmentResult

		fg.addToRemainingData(inputData, inputRange, statRange)
	}
	return nil
}

func (fg *FittingSingleStatSegmentsProcess) addToRemainingData(processedData []util_weight2.FittingSample, inputRange weight_types.StatRangeFloat, removeRange weight_types.StatRangeFloat) {
	if !inputRange.ContainsOtherRangeFloatAllowance(removeRange) {
		panic(fmt.Sprintf("range isn't within bounds %e %e %e %e", removeRange.Minimum, inputRange.Minimum, removeRange.Maximum, inputRange.Maximum))
	}
	if !removeRange.IsValid() {
		panic("invalid remove range")
	}
	if !inputRange.IsValid() {
		panic("invalid input range")
	}

	loData := make([]util_weight2.FittingSample, 0)
	hiData := make([]util_weight2.FittingSample, 0)
	for _, input := range processedData {
		stat := input.StatValue
		if util.FloatsBetween(inputRange.Minimum, stat, removeRange.Minimum) {
			loData = append(loData, input)
		} else if util.FloatsBetween(removeRange.Maximum, stat, inputRange.Maximum) {
			hiData = append(hiData, input)
		} else if util.FloatsBetween(inputRange.Minimum, stat, inputRange.Maximum) {
			// drop
		} else {
			panic(fmt.Sprintf("sample isn't within bounds %e %e %e", inputRange.Minimum, stat, inputRange.Maximum))
		}
	}

	if len(loData) > 0 {
		loRange := weight_types.StatRangeFloat{Minimum: inputRange.Minimum, Maximum: removeRange.Minimum - c_fitting_statScaledUnequalDelta}
		fg.samplesRemainingParts[loRange] = loData
	}

	if len(hiData) > 0 {
		hiRange := weight_types.StatRangeFloat{Minimum: removeRange.Maximum + c_fitting_statScaledUnequalDelta, Maximum: inputRange.Maximum}
		fg.samplesRemainingParts[hiRange] = hiData
	}
}
