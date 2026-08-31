package weight_types

import (
	"encoding/json/v2"
	"math"
	"os"
	"time"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

func WeightInputReadFile(filename string) ([]WeightInput, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var weightInputs []WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		return nil, err
	}
	return weightInputs, err
}

func WeightInputWriteFile(weightInputs []WeightInput, filename string) error {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		return err
	}
	return nil
}

func WeightInputReadFileOrPanic(filename string) []WeightInput {
	data, err := WeightInputReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func WeightInputWriteFileOrPanic(weightInputs []WeightInput, filename string) {
	err := WeightInputWriteFile(weightInputs, filename)
	if err != nil {
		panic(err)
	}
}

type IWeightResult interface {
	GetWeight() IWeight
	GetSolveTime() time.Duration
	GetStatus() string
	GetNewRatio() *SimPriorityBasic
	GetError() error
	AsWeight1(verificationInputs []WeightInput) *Weight1Basic
	AsWeight2(verificationInputs []WeightInput) *Weight2Extended
	AsWeight3(verificationInputs []WeightInput) *Weight3ExtendedRanged
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

type WeightResult1 weightResultGeneric[*Weight1Basic]

type WeightResult2 weightResultGeneric[*Weight2Extended]

type WeightResult3 weightResultGeneric[*Weight3ExtendedRanged]

type WeightResult4 weightResultGeneric[*Weight4Segmented]

func WeightResult1Make(weight *Weight1Basic, solveTime time.Duration, status highs.ModelStatus) WeightResult1 {
	return WeightResult1{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult2Make(weight *Weight2Extended, solveTime time.Duration, status highs.ModelStatus) WeightResult2 {
	return WeightResult2{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult3Make(weight *Weight3ExtendedRanged, solveTime time.Duration, status highs.ModelStatus) WeightResult3 {
	return WeightResult3{WeightResultCommon: WeightResultCommon{weight, solveTime, status.String(), nil, nil}, Weight: weight}
}
func WeightResult4Make(weight *Weight4Segmented, solveTime time.Duration, status highs.ModelStatus) WeightResult4 {
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
func WeightResult1MakeWithRatio(weight *Weight1Basic, solveTime time.Duration, status highs.ModelStatus, ratio *SimPriorityBasic, err error) WeightResult1 {
	return WeightResult1{WeightResultCommon{weight, solveTime, status.String(), ratio, err}, weight}
}

func (wr *WeightResult1) AsWeight1(_ []WeightInput) *Weight1Basic {
	return wr.Weight
}
func (wr *WeightResult2) AsWeight1(_ []WeightInput) *Weight1Basic {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight1()
	} else {
		return nil
	}
}
func (wr *WeightResult3) AsWeight1(verificationInputs []WeightInput) *Weight1Basic {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2(verificationInputs).ConvertToWeight1()
	} else {
		return nil
	}
}
func (wr *WeightResult4) AsWeight1(verificationInputs []WeightInput) *Weight1Basic {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2().ConvertToWeight1()
	} else {
		return nil
	}
}

func (wr *WeightResult1) AsWeight2(_ []WeightInput) *Weight2Extended {
	return nil
}
func (wr *WeightResult2) AsWeight2(_ []WeightInput) *Weight2Extended {
	return wr.Weight
}
func (wr *WeightResult3) AsWeight2(verificationInputs []WeightInput) *Weight2Extended {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2(verificationInputs)
	} else {
		return nil
	}
}
func (wr *WeightResult4) AsWeight2(_ []WeightInput) *Weight2Extended {
	if wr.Weight != nil {
		return wr.Weight.ConvertToWeight2()
	} else {
		return nil
	}
}

func (wr *WeightResult1) AsWeight3(_ []WeightInput) *Weight3ExtendedRanged {
	return nil
}
func (wr *WeightResult2) AsWeight3(_ []WeightInput) *Weight3ExtendedRanged {
	return nil
}
func (wr *WeightResult3) AsWeight3(_ []WeightInput) *Weight3ExtendedRanged {
	return wr.Weight
}
func (wr *WeightResult4) AsWeight3(_ []WeightInput) *Weight3ExtendedRanged {
	return nil
}

type StatRange struct {
	Minimum uint32
	Maximum uint32
}

func (rn StatRange) Equals(other StatRange) bool {
	return rn.Maximum == other.Maximum && rn.Minimum == other.Minimum
}

func (rn StatRange) Overlap(other StatRange) bool {
	if rn.Minimum > other.Maximum {
		return false
	} else if other.Minimum > rn.Maximum {
		return false
	} else {
		return true
	}
}

func (rn StatRange) RangeSize() uint32 {
	if rn.IsFullRange() {
		return math.MaxUint32
	} else {
		return rn.Maximum - rn.Minimum + 1
	}
}

func (rn StatRange) Contains(value uint32) bool {
	return rn.Minimum <= value && value <= rn.Maximum
}

func (rn StatRange) IsFullRange() bool {
	return rn.Minimum == 0 && rn.Maximum == math.MaxUint32
}

type StatRangeFloat struct {
	Minimum float64
	Maximum float64
}

func (rf StatRangeFloat) IsValid() bool {
	return rf.Minimum < rf.Maximum
}

func (rf StatRangeFloat) ContainsOtherRangeFloatAllowance(other StatRangeFloat) bool {
	return util.FloatApproxLessThanOrEqual(rf.Minimum, other.Minimum) &&
		util.FloatApproxLessThanOrEqual(other.Maximum, rf.Maximum)
}
