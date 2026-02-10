package main

import (
	"fmt"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	. "paladin_gearing_go/solver"
	. "paladin_gearing_go/types/items"
)

const (
	gearFile = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
)

func main() {
	model := Model_PallyProtMitigation()
	itemOptions := OptionsSetup_FromGearFile(gearFile, &model)

	fullSet := Solver(&itemOptions, &model)

	reportSet(fullSet, &model)
}

func reportSet(fullSet FullItemSet, model *Model) {
	rating := model.CalcRatingFull(&fullSet)
	fmt.Printf("SET OUTPUT rating %,d\n", rating)
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
