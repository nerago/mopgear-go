package main

import (
	"os"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	. "paladin_gearing_go/util"
	"runtime"
	"runtime/pprof"
	"time"
)

const (
	gearFileProtMitigation = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
	gearFileProtDps        = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-dps.json`
	gearFileRet            = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-ret.json`
	bagsFile               = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\bags-gear-bags.json`
	enableProfiling        = true
)

var printer = PrintRecorder{}

func main() {
	// a := stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// b := stats.StatBlock{7, 7, 7, 7, 7, 9, 9, 9, 9, 11, 11, 11}
	// fmt.Println(a)
	// fmt.Println(b)
	// stats.StatBlock_Increment_Reference(&a, &b)
	// fmt.Println(a)
	// fmt.Println(b)
	// fmt.Println()

	// a = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// b = stats.StatBlock{7, 7, 7, 7, 7, 9, 9, 9, 9, 11, 11, 11}
	// fmt.Println(a)
	// fmt.Println(b)
	// stats.StatBlock_Increment_Assem(&a, &b)
	// fmt.Println(a)
	// fmt.Println(b)
	// fmt.Println()

	// a = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// b = stats.StatBlock{7, 7, 7, 7, 7, 9, 9, 9, 9, 11, 11, 11}
	// fmt.Println(a)
	// fmt.Println(b)
	// stats.StatBlock_Increment_Assem_Vec(&a, &b)
	// fmt.Println(a)
	// fmt.Println(b)
	// fmt.Println()

	startTime := time.Now()

	if enableProfiling {
		f, err := os.Create("main.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	core()

	timeTaken := time.Since(startTime)
	printer.Println("Duration = " + timeTaken.String())

	if enableProfiling {
		f, err := os.Create("main-memory.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
            panic(err)
        }
	}
}

func core() {
	// itemOptions, model := setupPallyMitigation()
	// itemOptions, model := setupPallyDps()

	// slotRating(itemOptions[Equip_Chest], &model)
	// basicReforge(&itemOptions, &model, &printer)

	PaladinMultiRun()
}

func setupPallyMitigation() (FullOptionsMap, Model) {
	model := Model_PallyProtMitigation()
	return OptionsSetup_FromGearFile(gearFileProtMitigation, &model, &printer), model
}

func setupPallyDps() (FullOptionsMap, Model) {
	model := Model_PallyProtDps()
	return OptionsSetup_FromGearFile(gearFileProtDps, &model, &printer), model
}
