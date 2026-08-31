package updateProc

import (
	"cmp"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

func (spec *weightSpecInternal) tabularReportWriteFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	printer := util.FilePrintableMake(file)
	printer.Printf("Weight detail output at %s\n", time.Now().Format(time.DateTime))
	spec.tabularReport(printer)

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func (spec *weightSpecInternal) tabularReport(print util.Printable) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	tab.AddColumnHeader("acc1", false)
	tab.AddColumnHeader("acc1_stat", false)
	tab.AddColumnHeader("accX", false)
	tab.AddColumnHeader("accX_stat", false)
	tab.AddColumnHeader("time", false)
	tab.AddColumnHeader("status", false)
	tab.AddColumnHeader("pawn", false)

	choices := spec.out.getChoicesSafeCopy()
	slices.SortFunc(choices, func(a, b weightChoice) int {
		return cmp.Compare(max(a.accuracy1, a.accuracy1Stat), max(b.accuracy1, b.accuracy1Stat))
	})
	for choice := range util_collection.ForPointer(choices) {
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, choice.choiceName)
		row = append(row, strconv.FormatFloat(choice.accuracy1, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(choice.accuracy1Stat, 'f', 4, 64))
		if choice.hadExtended {
			row = append(row, strconv.FormatFloat(choice.accuracyX, 'f', 4, 64))
			row = append(row, strconv.FormatFloat(choice.accuracyXStat, 'f', 4, 64))
		} else {
			row = append(row, "", "")
		}
		if choice.weightResult != nil {
			row = append(row, choice.weightResult.SolveTime.String())
			row = append(row, choice.weightResult.Status)
		} else {
			row = append(row, "", "")
		}
		pawnString := tools.FormatPawnString(choice.weight1)
		row = append(row, pawnString)
		tab.AddRow(row)
	}

	print.Printf("TABLE %s\n", spec.param.Label)
	tab.Write(print)
}

func (spec *weightSpecInternal) reportAndWriteWeights() string {
	summary := spec.out.getSummary()

	// OVERWRITE WEIGHT FILE
	if bestChoice, hasBest := spec.out.bestWeightChoice1(); hasBest {
		summary.WriteString(" ::::: ")
		summary.WriteString(" W1(")
		summary.WriteString(bestChoice.choiceName)
		summary.WriteString(" ")
		summary.WriteFloat64(bestChoice.accuracy1, 4)
		summary.WriteString(" ")
		summary.WriteFloat64(bestChoice.accuracy1Stat, 4)
		summary.WriteString(") ")

		pawnString := tools.FormatPawnString(bestChoice.weight1)
		util.WriteStringToFile(spec.param.WeightFile1, pawnString)
	}

	weight2Opt, weight3Opt := spec.out.bestWeightChoiceExtended()

	if weight2Choice, hasWeight2 := weight2Opt.GetWithFlag(); hasWeight2 {
		str := tools.FormatWeight2String(weight2Choice.weight2)
		util.WriteStringToFile(spec.param.WeightFile2, str)

		summary.WriteString("W2(")
		summary.WriteString(weight2Choice.choiceName)
		summary.WriteRune(' ')
		summary.WriteFloat64(weight2Choice.accuracyXStat, 4)
		summary.WriteString(") ")
	}

	if weight3Choice, hasWeight3 := weight3Opt.GetWithFlag(); hasWeight3 {
		str := tools.FormatWeight3String(weight3Choice.weight3)
		util.WriteStringToFile(spec.param.WeightFile3, str)

		summary.WriteString("W3(")
		summary.WriteString(weight3Choice.choiceName)
		summary.WriteRune(' ')
		summary.WriteFloat64(weight3Choice.accuracyXStat, 4)
		summary.WriteString(") ")
	}

	spec.process.printer.PrintlnFromBuild(summary)

	logText := summary.Clone()
	logText.WriteRune('\n')
	logText.WriteString(time.Now().Format(time.DateTime))
	util.WriteStringToFile(spec.param.WeightFile1+"-accuracy.log", logText.String())

	spec.tabularReportWriteFile(spec.param.WeightFile1 + "-detail.log")

	return summary.String()
}
