package main

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/solver"
	. "paladin_gearing_go/util"
)

func basicReforge(itemOptions *FullOptionsMap, model *Model, printer *PrintRecorder) {
	fullSet := Solver(itemOptions, model, printer)
	reportSet(fullSet, model, printer)
}

func slotRating(itemArray []FullItem, model *Model, printer *PrintRecorder) {
	// printer.Println()
	// printer.Println("RATINGS")
	// printer.Println(model.StatRatings.(StatRatingsWeights).Weights())
	// printer.Println()

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

func reportSet(fullSet FullItemSet, model *Model, printer *PrintRecorder) {
	rating := model.CalcRatingFull(&fullSet)
	printer.Printf("SET OUTPUT rating %d\n", rating)
	printer.Printf("RATED %s\n", fullSet.TotalRated.String())
	printer.Printf("CAP %s\n", fullSet.TotalCap.String())
	printEquipMap(&fullSet.Items, printer)
	// TODO set bonus
}

func printEquipMap(fullEquipMap *FullEquipMap, printer *PrintRecorder) {
	for _, item := range fullEquipMap {
		printer.Println(item.String())
	}
}

func UNUSED(x ...interface{}) {}
