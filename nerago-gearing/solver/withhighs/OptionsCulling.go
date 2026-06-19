package withhighs

import (
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/util"
	"sync"
	"sync/atomic"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_cullThreadCount       = 10
	c_cullMinimumSlotSize   = 2
	c_cullAddTasksPerTask   = 2
	c_cullAddBlocksEachTime = 5
)

type cullTask struct {
	blockedItems map[items.ItemId]bool
}

func (task *cullTask) withMoreBlocks(removeItems []items.ItemId) cullTask {
	newBlocks := make(map[items.ItemId]bool, len(task.blockedItems)+len(removeItems))
	for itemId := range task.blockedItems {
		newBlocks[itemId] = true
	}
	for _, itemId := range removeItems {
		newBlocks[itemId] = true
	}
	return cullTask{newBlocks}
}

type OptionsCulling struct {
	label             string
	targetResultCount int32
	itemOptions       items.SolvableOptionsMap
	model             *gear_model.Model
	printer           *util.PrintRecorder

	tasksCompleted atomic.Int32
	queueHighwater atomic.Int64

	didRemoveLock sync.Mutex
	didRemove     map[items.ItemId]bool
}

func (process *OptionsCulling) Init(label string, targetResultCount int32, itemOptions items.SolvableOptionsMap, model *gear_model.Model, printer *util.PrintRecorder) {
	process.label = label
	process.targetResultCount = targetResultCount
	process.itemOptions = itemOptions
	process.model = model
	process.printer = printer
	process.didRemove = make(map[items.ItemId]bool)
}

func (process *OptionsCulling) Run() <-chan items.SolvableItemSet {
	process.printer.Printf("Running culling\n")

	resultChannel := make(chan items.SolvableItemSet, 8)
	taskChannel := make(chan cullTask, 0x1000000)

	taskChannel <- cullTask{make(map[items.ItemId]bool)}

	waitGroup := sync.WaitGroup{}
	for range c_cullThreadCount {
		waitGroup.Go(func() {
			for process.tasksCompleted.Load() < process.targetResultCount - c_cullThreadCount {
				task := <-taskChannel
				process.runTask(task, taskChannel, resultChannel)
				process.tasksCompleted.Add(1)
			}
			process.printer.Println("exit thread")
		})
	}
	go func() {
		waitGroup.Wait()
		process.reportHowManyTried()
		close(resultChannel)
		close(taskChannel)
	}()

	return resultChannel
}

func (process *OptionsCulling) reportHowManyTried() {
	itemIdOptions := distinctItemIdsAll(&process.itemOptions)
	process.printer.Printf("CULLING NUMS %s options=%d didRemove=%d highwater=%d\n", process.label, len(itemIdOptions), len(process.didRemove), process.queueHighwater.Load())
}

func (process *OptionsCulling) runTask(task cullTask, taskChannel chan cullTask, resultChannel chan<- items.SolvableItemSet) {
	linearBuild := utilhighs.LinearBuilder{}
	linearBuild.Solver = utilhighs.Solver_MIP_Interior
	linearBuild.NoOutput = true

	itemOptions, isTerminal := process.makeRestrictedItemOptions(task)
	if isTerminal {
		return
	}

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

		process.deriveNewTasks(result, itemOptions, task, taskChannel)
	}
}

func (process *OptionsCulling) makeRestrictedItemOptions(task cullTask) (items.SolvableOptionsMap, bool) {
	stringBuild := util.StringBuild2{}
	itemOptions := process.itemOptions.Clone()
	for removeItemId := range task.blockedItems {
		makesSlotEmpty := itemOptions.RemoveItemIdFromAll(removeItemId)
		if makesSlotEmpty {
			// terminal leaf point of algorithm, should try to avoid ending up here in first place
			process.printer.Println("CULL task reached terminal")
			return items.SolvableOptionsMap{}, true
		}
		stringBuild.WriteUint32(uint32(removeItemId))
		stringBuild.WriteRune(' ')
	}
	process.printer.Printf("TASK with removed: %s\n", stringBuild.String())

	process.recordDidRemove(task.blockedItems)

	return itemOptions, false
}

func (process *OptionsCulling) recordDidRemove(items map[items.ItemId]bool) {
	process.didRemoveLock.Lock()
	defer process.didRemoveLock.Unlock()
	for itemId := range items {
		process.didRemove[itemId] = true
	}
}

func (process *OptionsCulling) deriveNewTasks(chosenSet items.SolvableItemSet, itemOptions items.SolvableOptionsMap, task cullTask, taskChannel chan cullTask) {
	availableToRemove := process.selectItemsCanRemove(chosenSet, itemOptions)
	if len(availableToRemove) > 0 {
		for range c_cullAddTasksPerTask {
			removeItems := randomSampleSlice(availableToRemove, c_cullAddBlocksEachTime)
			taskChannel <- task.withMoreBlocks(removeItems)
		}

		channelSize := int64(len(taskChannel))
		currHighwater := process.queueHighwater.Load()
		for channelSize > currHighwater {
			if process.queueHighwater.CompareAndSwap(currHighwater, channelSize) {
				break
			}
			currHighwater = process.queueHighwater.Load()
		}
	}
}

func (*OptionsCulling) selectItemsCanRemove(chosenSet items.SolvableItemSet, itemOptions items.SolvableOptionsMap) []items.ItemId {
	availableToRemove := make([]items.ItemId, 0, items.ITEM_SLOT_COUNT)
	for slot := range chosenSet.Items() {
		chosenSlotItem := chosenSet.Items()[slot]
		if chosenSlotItem != nil {
			slotOptions := itemOptions.Get(items.SlotEquip(slot))
			slotItemIds := distinctItemIds(slotOptions)
			if len(slotItemIds) > c_cullMinimumSlotSize {
				availableToRemove = append(availableToRemove, chosenSlotItem.ItemId())
			}
		}
	}
	return availableToRemove
}

func distinctItemIds(slotOptions []items.SolvableItem) map[items.ItemId]bool {
	distinct := make(map[items.ItemId]bool)
	for _, item := range slotOptions {
		distinct[item.ItemId()] = true
	}
	return distinct
}

func distinctItemIdsAll(itemOptions *items.SolvableOptionsMap) map[items.ItemId]bool {
	distinct := make(map[items.ItemId]bool)
	for item := range itemOptions.AllItemSeq() {
		distinct[item.ItemId()] = true
	}
	return distinct
}

func randomSampleSlice(availableToRemove []items.ItemId, sampleCount int) []items.ItemId {
	if len(availableToRemove) <= sampleCount {
		return availableToRemove
	}

	util.Shuffle(availableToRemove)
	return availableToRemove[0:sampleCount]
}
