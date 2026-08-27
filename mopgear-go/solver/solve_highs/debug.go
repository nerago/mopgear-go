package solve_highs

import (
	"errors"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_highs"
)

func debugPrint(solution *util_highs.Solution2, build *util_highs.LinearBuilder, printer *util.PrintRecorder) {
	if !util_highs.C_DebugHighs {
		return
	}

	printer.Printf("OBJECTIVE VALUE = %f\n", solution.Objective())

	activeBonus := ""
	activeBonusWeight := 0.0

	for colIndex, outputValue := range solution.ColValuesSeq() {
		if colInfo, isInfo := build.DebugContextFor(colIndex).(*columnInfo); isInfo {
			debugPrintColumnEntry(colInfo, colIndex, outputValue, &activeBonus, &activeBonusWeight, printer)
		} else {
			text := build.DebugTextFor(colIndex)
			printer.Printf("%d %f %s\n", colIndex, outputValue, text)
		}
	}

	printer.Printf("ACTIVE highs Bonus = %s %f\n", activeBonus, activeBonusWeight)
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
		strBuild.WriteInt32(int32(colEntry.setIndex))
	case entry_set_exact_count:
		strBuild.WriteString("set exact count flag ")
		strBuild.WriteInt32(int32(colEntry.setIndex))
		strBuild.WriteRune(' ')
		strBuild.WriteInt64(int64(colEntry.itemCount))
	case entry_sum_rating:
		strBuild.WriteString("initial item rating sum")
	case entry_combo_active:
		strBuild.WriteString("combo active ")
		strBuild.WriteString(colEntry.combo.debugStr())
	case entry_main_output:
		strBuild.WriteString("final value ")
	case entry_multi_enable_forge:
		strBuild.WriteString("multi enable forge ")
		strBuild.WriteUint32(uint32(colEntry.itemFull.ItemId()))
		strBuild.WriteRune(' ')
		colEntry.itemFull.Total().AppendString(&strBuild)
	case entry_multi_output:
		strBuild.WriteString("multi output ")
	case entry_stat_total:
		strBuild.WriteString("stat total ")
		strBuild.WriteString(colEntry.statType.Name())
	case entry_sim_value:
		strBuild.WriteString("sim value ")
		strBuild.WriteString(colEntry.simType.Name())
	case entry_sim_stat_value:
		strBuild.WriteString("sim stat value ")
		strBuild.WriteString(colEntry.simType.Name())
		strBuild.WriteString(" ")
		strBuild.WriteString(colEntry.statType.Name())
	case entry_sim_stat_value_option:
		strBuild.WriteString("sim stat value option")
		strBuild.WriteString(colEntry.simType.Name())
		strBuild.WriteString(" ")
		strBuild.WriteString(colEntry.statType.Name())
		strBuild.WriteString(" ")
		strBuild.WriteUint32(colEntry.statRange.Minimum)
		strBuild.WriteString(" ")
		strBuild.WriteUint32(colEntry.statRange.Maximum)
	case entry_sim_value_combo:
		strBuild.WriteString("sim value for combo ")
		strBuild.WriteString(colEntry.simType.Name())
	default:
		strBuild.WriteString("unknown column")
	}
	return strBuild.String()
}

func debugPrintColumnEntry(colEntry *columnInfo, columnIndex util_highs.ColumnIndex, outputValue float64, activeBonus *string, activeBonusWeight *float64, printer *util.PrintRecorder) error {
	switch colEntry.entryType {
	case entry_item:
		printer.Printf("%d %f %s %s %d\n", columnIndex, outputValue, "item", colEntry.itemSlot.Name(), colEntry.item.ItemId())
	case entry_set_total_count:
		printer.Printf("%d %f %s %d\n", columnIndex, outputValue, "set total count", colEntry.setIndex)
	case entry_set_exact_count:
		printer.Printf("%d %f %s %d %d\n", columnIndex, outputValue, "set exact count flag", colEntry.setIndex, colEntry.itemCount)
	case entry_sum_rating:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "initial item rating sum")
	case entry_combo_active:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "combo active", colEntry.combo.debugStr())
		if util.FloatEqualsOne(outputValue) && activeBonus != nil {
			*activeBonus += colEntry.combo.debugStr()
		}
	case entry_main_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "final value")
	case entry_multi_enable_forge:
		printer.Printf("%d %f %s %d\n", columnIndex, outputValue, "multi enable forge", colEntry.itemFull.ItemId())
	case entry_multi_output:
		printer.Printf("%d %f %s\n", columnIndex, outputValue, "multi output")
	case entry_stat_total:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "stat total", colEntry.statType.Name())
	case entry_sim_value:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "sim value", colEntry.simType.Name())
	case entry_sim_stat_value:
		printer.Printf("%d %f %s %s %s\n", columnIndex, outputValue, "sim stat value", colEntry.simType.Name(), colEntry.statType.Name())
	case entry_sim_stat_value_option:
		printer.Printf("%d %f %s %s %s %d %d\n", columnIndex, outputValue, "sim stat option", colEntry.simType.Name(), colEntry.statType.Name(), colEntry.statRange.Minimum, colEntry.statRange.Maximum)
	case entry_sim_value_combo:
		printer.Printf("%d %f %s %s\n", columnIndex, outputValue, "sim value for combo", colEntry.simType.Name())
	default:
		return errors.New("unknown column")
	}
	return nil
}
