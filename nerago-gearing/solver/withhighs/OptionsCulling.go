package withhighs

import (
	"math/rand/v2"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/util"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_cullThreadCount     = 16
	c_cullMinimumSlotSize = 2
)

type OptionsCulling struct {
	label       string
	itemOptions items.SolvableOptionsMap
	model       *gear_model.Model
	printer     *util.PrintRecorder

	allItemIds []items.ItemId

	targetResultCount int64
	tasksCompleted    atomic.Int64

	didRemoveLock sync.Mutex
	didRemove     map[items.ItemId]bool
}

func (process *OptionsCulling) Init(label string, targetResultCount int64, itemOptions items.SolvableOptionsMap, model *gear_model.Model, printer *util.PrintRecorder) {
	process.label = label
	process.targetResultCount = targetResultCount
	process.itemOptions = itemOptions
	process.allItemIds = distinctItemIdsAll(&itemOptions)
	process.model = model
	process.printer = printer
	process.didRemove = make(map[items.ItemId]bool)
}

func (process *OptionsCulling) Run(shouldRun func() bool) <-chan items.SolvableItemSet {
	process.printer.Printf("Running culling\n")

	resultChannel := make(chan items.SolvableItemSet, 8)

	waitGroup := sync.WaitGroup{}
	for range c_cullThreadCount {
		waitGroup.Go(func() {
			for process.tasksCompleted.Load() < process.targetResultCount-c_cullThreadCount && shouldRun() {
				process.runTask(resultChannel)
			}
			process.printer.Println("exit thread")
		})
	}
	go func() {
		waitGroup.Wait()
		process.reportHowManyTried()
		close(resultChannel)
	}()

	return resultChannel
}

func (process *OptionsCulling) reportHowManyTried() {
	itemIdOptions := process.allItemIds
	process.printer.Printf("CULLING NUMS %s options=%d didRemove=%d\n", process.label, len(itemIdOptions), len(process.didRemove))
}

func (process *OptionsCulling) runTask(resultChannel chan<- items.SolvableItemSet) {
	blockedItems := process.chooseRandomToRemove()
	itemOptions, isUnusable := process.makeRestrictedItemOptions(blockedItems)
	if isUnusable {
		return
	}

	linearBuild := utilhighs.LinearBuilder{}
	linearBuild.Solver = utilhighs.Solver_MIP_Interior
	linearBuild.NoOutput = true

	setup := setupGearSet(&linearBuild, process.model, &itemOptions, 1)
	solution, _ := linearBuild.RunHighs()
	if solution.Status == highs.ModelStatusOptimal {
		percent := float64(process.tasksCompleted.Load()) / float64(process.targetResultCount) * 100
		process.printer.Printf("TASK OK %s %.0f\n", process.label, percent)
	} else {
		process.printer.Printf("TASK status = %s\n", solution.Status.String())
	}

	if solution.HasSolution() {
		result := setup.buildResultSet(solution, &itemOptions, process.model)
		resultChannel <- result
		process.tasksCompleted.Add(1)
	}
}

// bit inefficient since not slot aware could frequently empty out slots
func (process *OptionsCulling) chooseRandomToRemove() []items.ItemId {
	minRemove := 3
	maxRemove := len(process.allItemIds) - items.ITEM_SLOT_COUNT - 3

	toRemove := minRemove
	if maxRemove > minRemove {
		toRemove += rand.IntN(maxRemove - minRemove + 1)
	}

	return randomSampleSlice(process.allItemIds, toRemove)
}

func (process *OptionsCulling) makeRestrictedItemOptions(blockedItems []items.ItemId) (items.SolvableOptionsMap, bool) {
	stringBuild := util.StringBuild2{}
	itemOptions := process.itemOptions.Clone()
	for _, removeItemId := range blockedItems {
		makesSlotEmpty := itemOptions.RemoveItemIdFromAll(removeItemId)
		if makesSlotEmpty {
			// process.printer.Println("CULL task unusable")
			return items.SolvableOptionsMap{}, true
		}
		stringBuild.WriteUint32(uint32(removeItemId))
		stringBuild.WriteRune(' ')
	}
	process.printer.Printf("TASK: %s\n", stringBuild.String())

	process.recordDidRemove(blockedItems)

	return itemOptions, false
}

func (process *OptionsCulling) recordDidRemove(blockedItems []items.ItemId) {
	process.didRemoveLock.Lock()
	defer process.didRemoveLock.Unlock()
	for _, itemId := range blockedItems {
		process.didRemove[itemId] = true
	}
}

func distinctItemIdsAll(itemOptions *items.SolvableOptionsMap) []items.ItemId {
	distinct := make(map[items.ItemId]bool)
	for item := range itemOptions.AllItemSeq() {
		distinct[item.ItemId()] = true
	}
	return util.KeysToSlice(distinct)
}

func randomSampleSlice(sharedSlice []items.ItemId, sampleCount int) []items.ItemId {
	if len(sharedSlice) <= sampleCount {
		return sharedSlice
	}

	copiedSlice := slices.Clone(sharedSlice)
	util.Shuffle(copiedSlice)
	return copiedSlice[0:sampleCount]
}
