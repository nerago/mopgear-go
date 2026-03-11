package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	. "paladin_gearing_go/util"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	gearFileProtMitigation = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-defence.json`
	gearFileProtDps        = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-prot-dps.json`
	gearFileRet            = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\gear-ret.json`
	bagsFile               = `C:\Users\nicholas\Dropbox\prog\paladin_gearing\bags-gear-bags.json`
	enableProfiling        = true
)

var printer *util.PrintRecorder

func main() {
	lowerPriority()

	printer = PrintRecorder_CreateLogFile()
	defer printer.Close()

	log.SetOutput(io.Discard) // ignore wowsim's internal progress logs

	if enableProfiling {
		f, err := os.Create("main.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	startTime := time.Now()

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
	itemOptions, model := setupPallyDps()

	// slotRating(itemOptions[Equip_Chest], &model)
	basicReforge(&itemOptions, &model, printer)

	// PaladinMultiRun()
	// testSim()

	// fmt.Println(cpu.X86.HasAVX2)

	// a := stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// b := stats.StatBlock{1,2,3,4,5,6,7,8,9,10,11,12}
	// c := stats.StatBlock{1, 2, 1, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// d := stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 12}
	// fmt.Println(stats.StatBlock_Equals_Assem(&a, &b))
	// fmt.Println(stats.StatBlock_Equals_Assem(&a, &c))
	// fmt.Println(stats.StatBlock_Equals_Assem(&a, &d))

	// a := stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// c := stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// fmt.Println(stats.StatBlock_Equals(&a, &c))
	// fmt.Println(StringFull(&a))
	// fmt.Println(StringFull(&c))
	// fmt.Println()
	// fmt.Println()

	// a = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// c = stats.StatBlock{1, 2, 666, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// fmt.Println(stats.StatBlock_Equals(&a, &c))
	// fmt.Println(StringFull(&a))
	// fmt.Println(StringFull(&c))
	// fmt.Println()
	// fmt.Println()

	// a = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// c = stats.StatBlock{1, 2, 3, 4, 665, 6, 7, 8, 9, 10, 11, 12}
	// fmt.Println(stats.StatBlock_Equals(&a, &c))
	// fmt.Println(StringFull(&a))
	// fmt.Println(StringFull(&c))
	// fmt.Println()
	// fmt.Println()

	// a = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	// c = stats.StatBlock{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 664}
	// fmt.Println(stats.StatBlock_Equals(&a, &c))
	// fmt.Println(StringFull(&a))
	// fmt.Println(StringFull(&c))
	// fmt.Println()
	// fmt.Println()

	// a = stats.StatBlock{5, 5, 4, 55674, 1123, 88, 123, 44, 4, 1, 0, 10}
	// c = stats.StatBlock{5, 5, 4, 55674, 1123, 88, 123, 44, 4, 1, 0, 10}
	// fmt.Println(stats.StatBlock_Equals(&a, &c))
	// fmt.Println(StringFull(&a))
	// fmt.Println(StringFull(&c))
	// fmt.Println()
	// fmt.Println()
}

func StringFull(block *stats.StatBlock) string {
	build := strings.Builder{}
	build.WriteString("{")
	for _, value := range block {
		build.WriteString(strconv.FormatUint(uint64(value), 10))
		build.WriteString(" ")
	}

	build.WriteString("}")
	return build.String()
}

func setupPallyMitigation() (FullOptionsMap, Model) {
	model := Model_PallyProtMitigation()
	return OptionsSetup_FromGearFile(gearFileProtMitigation, &model, printer), model
}

func setupPallyDps() (FullOptionsMap, Model) {
	model := Model_PallyProtDps()
	return OptionsSetup_FromGearFile(gearFileProtDps, &model, printer), model
}

func lowerPriority() {
	// NOTE go command mangles the double quote in priority if allowed to build command line
	pid := strconv.Itoa(os.Getpid())
	cmd := exec.Command(`C:\Windows\System32\wbem\wmic.exe`)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `C:\\Windows\\System32\\wbem\\wmic.exe process where ProcessId=` + pid + ` CALL setpriority "below normal"`}
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
