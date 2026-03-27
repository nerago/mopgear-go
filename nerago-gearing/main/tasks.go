package main

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

func basicReforge(itemOptions *items.FullOptionsMap, model *model.Model, printer *util.PrintRecorder) {
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         itemOptions,
		Model:               model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Long,
		Printer:             nil})
	output.Report(printer)
}

func testSim() {
	itemOptions, model := setupPallyMitigation()
	output := solver.Solver(solver.SolveInput{
		ItemOptions:         &itemOptions,
		Model:               &model,
		PhasedAcceptable:    false,
		EnableTrackProgress: true,
		SolveSize:           solver.SolveSize_Medium,
		Printer:             nil})
	resultStats := simulate.WowSim_Execute(simulate.RunSize_QuickDirty, model.Spec, output.FullSet.Items(), &model, nil, util.TrackProgress_Start())
	resultStats.Print(printer)
}

func slotRating(itemArray []items.FullItem, model *model.Model, printer *util.PrintRecorder) {
	printer.Println("RATINGS")
	// printer.Println(model.StatRatings.(ratings.StatRatingsWeights).Weights())
	printer.Println(model.StatRatings.Weights().CreateString())
	printer.Println0()

	best := util_rank.BestCollector1[items.FullItem]{}
	for _, item := range itemArray {
		rate := model.CalcRatingFullItem(&item)
		printer.Println(item.CreateString())
		printer.Printf("%d\n\n", rate)
		best.Offer(&item, rate)
	}

	printer.Println0()
	printer.Println("BEST")
	printer.Println(best.BestObject.CreateString())
}
