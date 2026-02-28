package main

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	"paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/solver"
	. "paladin_gearing_go/util"
)

func basicReforge(itemOptions *FullOptionsMap, model *Model, printer *PrintRecorder) {
	output := Solver(itemOptions, model, false)
	output.Report(printer)
}

func slotRating(itemArray []FullItem, model *Model, printer *PrintRecorder) {
	printer.Println("RATINGS")
	printer.Println(model.StatRatings.(ratings.StatRatingsWeights).Weights())
	printer.Println0()

	best := BestCollector1[FullItem]{}
	for _, item := range itemArray {
		rate := model.CalcRatingFullItem(&item)
		printer.Println(item.String())
		printer.Printf("%d\n\n", rate)
		best.Offer(&item, rate)
	}

	printer.Println0()
	printer.Println("BEST")
	printer.Println(best.BestObject.String())
}

func UNUSED(x ...interface{}) {}
