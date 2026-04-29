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
	dataInternal []float64
}

func (row *constraintRow) add(value float64) {
	row.dataInternal = append(row.dataInternal, value)
}

func (row *constraintRow) getDataChecked() []float64 {
	nonZeros := 0
	for _, val := range row.dataInternal {
		if val != 0.0 {
			nonZeros++
		}
	}
	if nonZeros == 0 {
		panic("row of all zeros")
	}
	return row.dataInternal
}

type constraintRowSparse struct {
	constMatrix *[]highs.Nonzero
	row         int
	currCol     int
}

func (row *constraintRowSparse) init(m *highs.Model, lowerBound float64, upperBound float64) {
	row.constMatrix = &m.ConstMatrix
	row.row = len(m.RowLower)
	row.currCol = 0
	m.RowLower = append(m.RowLower, lowerBound)
	m.RowUpper = append(m.RowUpper, upperBound)
}

func (row *constraintRowSparse) add(value float64) {
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
