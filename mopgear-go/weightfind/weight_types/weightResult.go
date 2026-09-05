package weight_types

import (
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
)

type IWeightResult interface {
	GetWeight() IWeight
	GetSolveTime() time.Duration
	GetStatus() string
	GetNewRatio() *SimPriorityBasic
	GetError() error
	AsWeight1(sampleInputs []WeightInput) *Weight1_ScaledSolvable
	AsWeight2(sampleInputs []WeightInput) *Weight2
	AsWeight3(sampleInputs []WeightInput) *Weight3
}

type WeightResultCommon struct {
	WeightInterface IWeight
	SolveTime       time.Duration
	Status          string
	NewRatio        *SimPriorityBasic
	Error           error
}

func (w WeightResultCommon) GetWeight() IWeight {
	return w.WeightInterface
}

func (w WeightResultCommon) GetSolveTime() time.Duration {
	return w.SolveTime
}

func (w WeightResultCommon) GetStatus() string {
	return w.Status
}

func (w WeightResultCommon) GetNewRatio() *SimPriorityBasic {
	return w.NewRatio
}

func (w WeightResultCommon) GetError() error {
	return w.Error
}

type weightResultGeneric[W IWeight] struct {
	WeightResultCommon
	Weight W
}

type WeightResult1 weightResultGeneric[*Weight1_ScaledSolvable]

type WeightResult2 weightResultGeneric[*Weight2]

type WeightResult3 weightResultGeneric[*Weight3]

type WeightResult4 weightResultGeneric[*Weight4]

func WeightResult1Make(weight *Weight1_ScaledSolvable, solveTime time.Duration, status highs.ModelStatus) WeightResult1 {
	return WeightResult1{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult2Make(weight *Weight2, solveTime time.Duration, status highs.ModelStatus) WeightResult2 {
	return WeightResult2{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult3Make(weight *Weight3, solveTime time.Duration, status highs.ModelStatus) WeightResult3 {
	return WeightResult3{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult4Make(weight *Weight4, solveTime time.Duration, status highs.ModelStatus) WeightResult4 {
	return WeightResult4{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult1MakeError(solveTime time.Duration, err error) WeightResult1 {
	return WeightResult1{WeightResultCommon: WeightResultCommon{SolveTime: solveTime, Status: "ERROR", Error: err}}
}
func WeightResult2MakeError(solveTime time.Duration, err error) WeightResult2 {
	return WeightResult2{WeightResultCommon: WeightResultCommon{SolveTime: solveTime, Status: "ERROR", Error: err}}
}
func WeightResult3MakeError(solveTime time.Duration, err error) WeightResult3 {
	return WeightResult3{WeightResultCommon: WeightResultCommon{SolveTime: solveTime, Status: "ERROR", Error: err}}
}
func WeightResult4MakeError(solveTime time.Duration, err error) WeightResult4 {
	return WeightResult4{WeightResultCommon: WeightResultCommon{SolveTime: solveTime, Status: "ERROR", Error: err}}
}
func WeightResult1MakeWithRatio(weight *Weight1_ScaledSolvable, solveTime time.Duration, status highs.ModelStatus, ratio *SimPriorityBasic, err error) WeightResult1 {
	return WeightResult1{WeightResultCommon{weight, solveTime, status.String(), ratio, err}, weight}
}

func (wr *WeightResult1) AsWeight1(_ []WeightInput) *Weight1_ScaledSolvable {
	return wr.Weight
}
func (wr *WeightResult2) AsWeight1(sampleInputs []WeightInput) *Weight1_ScaledSolvable {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight1(sampleInputs)
	} else {
		return nil
	}
}
func (wr *WeightResult3) AsWeight1(sampleInputs []WeightInput) *Weight1_ScaledSolvable {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2(sampleInputs).ConvertToWeight1(sampleInputs)
	} else {
		return nil
	}
}
func (wr *WeightResult4) AsWeight1(sampleInputs []WeightInput) (*Weight1_ScaledSolvable, error) {
	if wr.Weight != nil {
		weight2, err := wr.Weight.ConvertToWeight2(sampleInputs)
		if err != nil {
			return nil, err
		}
		return weight2.ConvertToWeight1(sampleInputs), nil
	} else {
		return nil, nil
	}
}

func (wr *WeightResult1) AsWeight2(_ []WeightInput) *Weight2 {
	return nil
}
func (wr *WeightResult2) AsWeight2(_ []WeightInput) *Weight2 {
	return wr.Weight
}
func (wr *WeightResult3) AsWeight2(sampleInputs []WeightInput) *Weight2 {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2(sampleInputs)
	} else {
		return nil
	}
}
func (wr *WeightResult4) AsWeight2(sampleInputs []WeightInput) (*Weight2, error) {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2(sampleInputs)
	} else {
		return nil, nil
	}
}

func (wr *WeightResult1) AsWeight3(_ []WeightInput) *Weight3 {
	return nil
}
func (wr *WeightResult2) AsWeight3(_ []WeightInput) *Weight3 {
	return nil
}
func (wr *WeightResult3) AsWeight3(_ []WeightInput) *Weight3 {
	return wr.Weight
}
func (wr *WeightResult4) AsWeight3(_ []WeightInput) *Weight3 {
	return nil
}
