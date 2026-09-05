package tools

import (
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/ratings"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func StatRatingsWeightsExtended_ReadFile(filename string, requiredStats []stats.StatType, sampleInputs []weight_types.WeightInput) (*ratings.StatRatingsWeightsExtended, error) {
	weight1Compatible, weight1Solvable, err1 := readEitherWeight1Format(filename, sampleInputs, requiredStats)
	if err1 != nil {
		return nil, err1
	}

	weight2, _ := ReadWeight2File(files.NameToWeight2(filename))
	weight3, _ := ReadWeight3File(files.NameToWeight3(filename))

	if weight1Solvable.IsEmpty() {
		panic("missing weight")
	}

	return &ratings.StatRatingsWeightsExtended{
		Weight1Compatible: *weight1Compatible,
		Weight1Scaled:     *weight1Solvable,
		Weight2:           weight2,
		Weight3:           weight3,
	}, nil
}

func readEitherWeight1Format(filename string, sampleInputs []weight_types.WeightInput, requiredStats []stats.StatType) (*weight_types.Weight1_CompatibleExternal, *weight_types.Weight1_ScaledSolvable, error) {
	sniffType, err := util.ReadFileSniffTen(filename)
	if err != nil {
		return nil, nil, err
	}

	if sniffType == 'P' {
		pawnBlock, err := readPawnWeightAsBlock(filename)
		if err != nil {
			return nil, nil, err
		}

		weight1Compatible := weight_types.Weight1Basic_Make_CompatibleExternal_FromBlock(pawnBlock)
		weight1Solvable, err := weight1Compatible.ConvertToSolvable(sampleInputs)
		if err != nil {
			return nil, nil, err
		}

		return weight1Compatible, &weight1Solvable, nil
	} else if sniffType == 'J' {
		weight1Solvable, err := ReadWeight1File(filename)
		if err != nil {
			return nil, nil, err
		} else if weight1Solvable == nil {
			return nil, nil, util.ErrorTracedNew("missing weights file " + filename)
		}

		weight1Compatible := weight1Solvable.ConvertToExternal(requiredStats)

		return &weight1Compatible, weight1Solvable, nil
	} else {
		return nil, nil, util.ErrorTracedNew("unknown weight file type: " + filename)
	}
}
