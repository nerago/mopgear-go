package solve_highs

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_highs"
)

func debugPrint(solution *util_highs.Solution2, build *util_highs.LinearBuilder, allColumns []*columnInfo, printer *util.PrintRecorder) {
	if !util_highs.C_DebugHighs {
		return
	}

	printer.Printf("OBJECTIVE VALUE = %f\n", solution.Objective()*c_scaled_ratings)

	activeBonus := ""
	activeBonusWeight := 0.0

	for colIndex, outputValue := range solution.ColValuesSeq() {
		if !debugPrintColumn(allColumns, colIndex, outputValue, &activeBonus, &activeBonusWeight, printer) {
			text := build.DebugTextFor(colIndex)
			printer.Printf("%d %f %s\n", colIndex, outputValue, text)
		}
	}

	printer.Printf("ACTIVE highs Bonus = %s %f\n", activeBonus, activeBonusWeight)
}

func debugPrintColumn(allColumns []*columnInfo, columnIndex util_highs.ColumnIndex, outputValue float64, activeBonus *string, activeBonusWeight *float64, printer *util.PrintRecorder) bool {
	var colEntry *columnInfo
	found := false
	for _, col := range allColumns {
		if col.columnIndex == columnIndex {
			colEntry = col
			found = true
			break
		}
	}

	if found {
		debugPrintColumnEntry(colEntry, columnIndex, outputValue, activeBonus, activeBonusWeight, printer)
	}
	return found
}

func (colEntry columnInfo) DebugText() string {
	strBuild := util.StringBuild2{}
	switch colEntry.entryType {
	case entry_item:
		strBuild.WriteString("item ")
		strBuild.WriteString(colEntry.itemSlot.Name())
		strBuild.WriteRune(' ')
		strBuild.WriteUint32(uint32(colEntry.item.ItemId()))
		strBuild.WriteRune(' ')
		colEntry.item.Total().AppendString(&strBuild)
	case entry_set_total_count:
		strBuild.WriteString("set total count ")
		strBuild.WriteString(colEntry.set.Name())
	case entry_set_exact_count:
		strBuild.WriteString("set exact count flag ")
		strBuild.WriteString(colEntry.set.Name())
		strBuild.WriteRune(' ')
		strBuild.WriteInt64(int64(colEntry.itemCount))
	case entry_sum_rating:
		strBuild.WriteString("initial item rating sum")
	case entry_permutation_active:
		strBuild.WriteString("permutation active ")
		strBuild.WriteString(colEntry.permutation.debugStr())
	case entry_permutation_output_weighted:
		strBuild.WriteString("permutation weighted output ")
		strBuild.WriteString(colEntry.permutation.debugStr())
		strBuild.WriteRune(' ')
		strBuild.WriteFloat64(colEntry.weight, 2)
	case entry_main_output:
		strBuild.WriteString("final value ")
	case entry_multi_enable_forge:
		strBuild.WriteString("multi enable forge ")
		strBuild.WriteUint32(uint32(colEntry.itemFull.ItemId()))
		strBuild.WriteRune(' ')
		colEntry.itemFull.Total().AppendString(&strBuild)
	case entry_multi_output:
		strBuild.WriteString("multi output ")
	default:
		panic("unknown column")
	}
	return strBuild.String()
}

func debugPrintColumnEntry(colEntry *columnInfo, columnIndex util_highs.ColumnIndex, outputValue float64, activeBonus *string, activeBonusWeight *float64, printer *util.PrintRecorder) {
	switch colEntry.entryType {
	case entry_item:
		printer.Printf("%d %f %s %s %d\n", columnIndex, outputValue, "item", colEntry.itemSlot.Name(), colEntry.item.ItemId())
	case entry_set_total_count:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "set total count", colEntry.set.Name())
	case entry_set_exact_count:
		printer.Printf("%d %f %s %s %d\n", columnIndex, outputValue, "set exact count flag", colEntry.set.Name(), colEntry.itemCount)
	case entry_sum_rating:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "initial item rating sum")
	case entry_permutation_active:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "permutation active", colEntry.permutation.debugStr())
		if util.FloatEqualsOne(outputValue) && activeBonus != nil {
			*activeBonus += colEntry.permutation.debugStr()
		}
	case entry_permutation_output_weighted:
		printer.Printf("%d %f %s %s %f\n", columnIndex, outputValue, "permutation weighted output", colEntry.permutation.debugStr(), colEntry.weight)
		if !util.FloatEqualsZero(outputValue) && activeBonusWeight != nil {
			*activeBonusWeight += colEntry.weight
		}
	case entry_main_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "final value")
	case entry_multi_enable_forge:
		printer.Printf("%d %f %s %d\n", columnIndex, outputValue, "multi enable forge", colEntry.itemFull.ItemId())
	case entry_multi_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "multi output")
	default:
		panic("unknown column")
	}
}
