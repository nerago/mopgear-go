package main

import (
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	. "paladin_gearing_go/solver"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/util"

	// . "paladin_gearing_go/model/ratings"
	// . "paladin_gearing_go/types/common"
	"fmt"
	"time"
)

const (
	gearFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
)

func main() {
	startTime := time.Now()

	printer := PrintRecorder{}
	model := Model_PallyProtMitigation()
	itemOptions := OptionsSetup_FromGearFile(gearFile, &model, &printer)

	// slotRating(itemOptions[Equip_Chest], &model)

	fullSet := Solver(&itemOptions, &model)
	reportSet(fullSet, &model)

	timeTaken := time.Since(startTime)
	fmt.Println("Duration = " + timeTaken.String())
}

// func slotRating(itemArray []FullItem, model *Model) {
// 	fmt.Println()
// 	fmt.Println("RATINGS")
// 	fmt.Println(model.StatRatings.(StatRatingsWeights).Weights())
// 	fmt.Println()
// 	fmt.Println()

// 	best := BestCollector1[FullItem]{}
// 	for _, item := range itemArray {
// 		rate := model.CalcRatingFullItem(item)
// 		fmt.Println(item.String())
// 		fmt.Printf("%d\n\n", rate)
// 		best.Offer(&item, rate)
// 	}

// 	fmt.Println()
// 	fmt.Println("BEST")
// 	fmt.Println(best.BestObject.String())
// }

func reportSet(fullSet FullItemSet, model *Model) {
	rating := model.CalcRatingFull(&fullSet)
	fmt.Printf("SET OUTPUT rating %d\n", rating)
	fmt.Printf("RATED %s\n", fullSet.TotalRated.String())
	fmt.Printf("CAP %s\n", fullSet.TotalCap.String())
	printEquipMap(&fullSet.Items)
	// TODO set bonus
}

func printEquipMap(fullEquipMap *FullEquipMap) {
	for _, item := range fullEquipMap {
		fmt.Println(item.String())
	}
}
