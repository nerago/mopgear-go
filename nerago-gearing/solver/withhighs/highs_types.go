package withhighs

import (
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"

	"github.com/lanl/highs"
)

type entryType int8

const (
	entry_item           entryType = iota
	entry_item_extracopy entryType = iota
	entry_set_item_count entryType = iota
)

type lookupEntry struct {
	purpose  entryType
	itemSlot items.SlotEquip
	item     *items.SolvableItem
	set      gear_model.ActiveSet
}

type constraintRow struct {
	constraintRowRaw
}

func constraintRow_make(high_model *highs.RawModel, lowerBound float64, upperBound float64) constraintRow {
	return constraintRow{constraintRowRaw_make(high_model, lowerBound, upperBound)}
}

func constraintRow_make_nil() constraintRow {
	return constraintRow{constraintRowRaw_make_nil()}
}

type constraintRowRaw struct {
	high_model *highs.RawModel
	lowerBound float64
	upperBound float64

	insertColumn  int
	columnNumbers []int
	values        []float64
}

func (row *constraintRowRaw) finish() {
	// expected nil row
	if row.high_model == nil {
		return
	}

	var err error
	if len(row.values) > 0 {
		err = row.high_model.AddCompSparseRows(
			[]float64{row.lowerBound},
			[]int{0}, 
			row.columnNumbers,
			row.values,
			[]float64{row.upperBound},
		)
	} else {
		// need to set an explicit zero value so array isn't empty
		// i'd argue this is a bug in go/highs binding library, 
		// empty array should be acceptable to lower level code
		err = row.high_model.AddCompSparseRows(
			[]float64{row.lowerBound},
			[]int{0}, 
			[]int{0},
			[]float64{0.0},
			[]float64{row.upperBound},
		)
	}
	
	if err != nil {
		panic(err)
	}
}

func (row *constraintRowRaw) add(value float64) {
	if row.high_model == nil && value != 0.0 {
		panic("expected nil row recived a non zero")
	}

	if value != 0.0 {
		row.columnNumbers = append(row.columnNumbers, row.insertColumn)
		row.values = append(row.values, value)
	}
	row.insertColumn++
}

func constraintRowRaw_make(high_model *highs.RawModel, lowerBound float64, upperBound float64) constraintRowRaw {
	return constraintRowRaw{
		high_model:   high_model,
		lowerBound:   lowerBound,
		upperBound:   upperBound,
		insertColumn: 0,
	}
}

func constraintRowRaw_make_nil() constraintRowRaw {
	return constraintRowRaw{}
}

type constraintRowDirect struct {
	constMatrix *[]highs.Nonzero
	row         int
	currCol     int
}

func constraintRowDirect_make(m *highs.Model, lowerBound float64, upperBound float64) constraintRowDirect {
	row := len(m.RowLower)
	m.RowLower = append(m.RowLower, lowerBound)
	m.RowUpper = append(m.RowUpper, upperBound)
	return constraintRowDirect{
		constMatrix: &m.ConstMatrix,
		row:         row,
		currCol:     0,
	}
}

func constraintRowDirect_make_null() constraintRowDirect {
	return constraintRowDirect{
		constMatrix: nil,
	}
}

func (row *constraintRowDirect) add(value float64) {
	if row.constMatrix == nil && value != 0.0 {
		panic("expected nil row recived a non zero")
	}

	if value != 0.0 {
		nz := highs.Nonzero{
			Row: row.row,
			Col: row.currCol,
			Val: value,
		}
		*row.constMatrix = append(*row.constMatrix, nz)
	}
	row.currCol++
}

const c_permuteCodeShift = 3
const c_maxSetItems = 5 // fundamental in MoP gear sets

type multiSetPermutation struct {
	totalBonus         float64
	permutationCode    float64 // define a permutation code like 201 positionally. can't set a nice range, but equality constraint will cover us
	eachCount          []float64
	slotsEachEqualCode [16]constraintRow
}
