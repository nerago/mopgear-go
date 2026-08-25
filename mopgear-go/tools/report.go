package tools

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util"
)

// reporting

func ReportSet(model_obj *gear_model.SpecModel, fullSet *items.FullItemSet, printer *util.PrintRecorder) {
	//printer.Printf("SET rating %.0f\n", rating)
	printer.Printf("BONUS counts %s\n", bonus_set.AllBonusesText(fullSet.Items()))
	printer.Printf("BONUS multiply %f\n", model_obj.BonusEnabled.CalcBonusFullFlat(fullSet.Items()))
	fullSet.PrintStats(printer)
	printEquipMap(fullSet.Items(), printer)

	WowSimJson_Write(fullSet.Items(), model_obj, printer)

	fullSet.DebugValidate()
}

func printEquipMap(fullEquipMap *items.FullEquipMap, printer *util.PrintRecorder) {
	for _, item := range fullEquipMap {
		if item != nil {
			printer.Println(item.CreateString())
		}
	}
}
