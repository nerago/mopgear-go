package tools

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

// reporting

func ReportSetFewerParams(model *model.Model, fullSet *items.FullItemSet, printer *util.PrintRecorder) {
	ReportSet(model, fullSet, model.CalcRatingFullAsFloat(fullSet), printer)
}

func ReportSet(model_obj *model.Model, fullSet *items.FullItemSet, rating float64, printer *util.PrintRecorder) {
	printer.Printf("SET rating %.0f\n", rating)
	printer.Printf("BONUS counts %s\n", model.AllBonusesText(fullSet.Items()))
	printer.Printf("BONUS multiply %f\n", model_obj.SetBonus.CalcBonusFull(fullSet.Items()))
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
