package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"paladin_gearing_go/files"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	. "paladin_gearing_go/setup"
	"paladin_gearing_go/util"
	. "paladin_gearing_go/util"
	"runtime"
	"runtime/pprof"
	"strconv"
	"syscall"
	"time"
)

const (
	enableProfiling = true
)

var printer *util.PrintRecorder

func main() {
	lowerPriority()

	printer = PrintRecorder_CreateLogFile()
	defer printer.Close()

	log.SetOutput(io.Discard) // ignore wowsim's internal progress logs

	if enableProfiling {
		f, err := os.Create(files.ProfileDir + "main.pgo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	startTime := time.Now()

	core(printer)

	timeTaken := time.Since(startTime)
	printer.Println("Duration = " + timeTaken.String())

	if enableProfiling {
		f, err := os.Create(files.ProfileDir + "main-memory.pgo")
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

func core(printer *PrintRecorder) {
	itemOptions, model := setupPallyMitigation()
	// itemOptions, model := setupPallyDps()

	// slotRating(itemOptions[Equip_Chest], &model, printer)
	basicReforge(&itemOptions, &model, printer)

	// PaladinMultiRun(printer)
	// testSim(printer)
}

func setupPallyMitigation() (FullOptionsMap, Model) {
	model := Model_PallyProtMitigation()
	return OptionsSetup_FromGearFile(files.GearFileProtMitigation, &model, printer), model
}

func setupPallyDps() (FullOptionsMap, Model) {
	model := Model_PallyProtDps()
	return OptionsSetup_FromGearFile(files.GearFileProtDps, &model, printer), model
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
