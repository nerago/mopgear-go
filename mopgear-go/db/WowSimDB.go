package db

import (
	"encoding/json"
	"iter"
	"os"

	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/stats/extern_stats"
	"github.com/nerago/mopgear-go/util"
)

var loaded = false
var itemsById map[items.ItemId][]items.FullItem = make(map[items.ItemId][]items.FullItem)
var reforgeById map[uint16]stats.ReforgeRecipe = make(map[uint16]stats.ReforgeRecipe)
var reforgeByObj map[stats.ReforgeRecipe]uint16 = make(map[stats.ReforgeRecipe]uint16)

func WowSimDB_Read() {
	filename := files.WowSimDB
	allBytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var inputObject map[string]any
	json.Unmarshal(allBytes, &inputObject)

	convertItems(inputObject["items"].([]any))
	convertReforge(inputObject["reforgeStats"].([]any))

	loaded = true
}

func WowSimDB_HasItemId(itemId items.ItemId) bool {
	_, found := itemsById[itemId]
	return found
}

func WowSimDB_LoadItemById(itemId items.ItemId, upgradeLevel int32) *items.FullItem {
	known := itemsById[itemId]
	for _, item := range known {
		// cater to bags file which currently doesn't use upgrade level as such, just current item level
		if int32(item.UpgradeLevel()) == upgradeLevel {
			return &item
		} else if item.ItemLevel() == uint16(upgradeLevel) {
			return &item
		}
	}

	return nil
}

func WowSimDB_LoadItemById_AllowFallback(itemId items.ItemId, upgradeLevel int32, printer *util.PrintRecorder) *items.FullItem {
	storedItem := WowSimDB_LoadItemById(itemId, upgradeLevel)

	if storedItem == nil && upgradeLevel > 0 {
		storedItem = WowSimDB_LoadItemById(itemId, 0)
		if storedItem != nil {
			printer.Printf("NOT FOUND at specified upgrade %d = %s\n", upgradeLevel, storedItem.CreateString())
		}
	}

	if storedItem == nil {
		panic("NOT FOUND at any upgrade level itemid=" + itemId.String())
	}

	return storedItem
}

func LookupItemNameByItemId(itemId items.ItemId) string {
	known := itemsById[itemId]
	for _, item := range known {
		return item.BaseName()
	}
	return "Unknown Item"
}

func WowSimDB_AllItems() iter.Seq[*items.FullItem] {
	return func(yield func(*items.FullItem) bool) {
		for _, subList := range itemsById {
			for i := range subList {
				if !yield(&subList[i]) {
					return
				}
			}
		}
	}
}

func WowSimDB_ReforgeById(reforgeId uint16) stats.ReforgeRecipe {
	recipe, ok := reforgeById[reforgeId]
	if !ok {
		panic("reforge not found")
	}
	return recipe
}

func WowSimDB_ReforgeToId(recipe stats.ReforgeRecipe) uint16 {
	id, ok := reforgeByObj[recipe]
	if !ok {
		panic("reforge not found")
	}
	return id
}

func convertItems(itemArray []any) {
	for _, element := range itemArray {
		itemObj := element.(map[string]any)
		addItem(itemObj)
	}
}

func addItem(itemObj map[string]any) {
	itemId := items.ItemId(getUInt32OrPanic(itemObj, "id"))
	name := itemObj["name"].(string)
	phase := int8(getIntOrDefault(itemObj, "phase", -1))
	itemType := getIntOrDefault(itemObj, "type", -1)
	if itemType == -1 {
		return
	}

	handType := getIntOrDefault(itemObj, "handType", 0)
	slot := items.MapSlotToGear(itemType, handType)

	armorType := convertArmorType(getIntOrDefault(itemObj, "armorType", -1))

	var socketSlots []stats.SocketType
	if itemObj["gemSockets"] != nil {
		socketSlots = convertSockets(itemObj["gemSockets"].([]any))
	}

	var socketBonus stats.StatBlock
	if itemObj["socketBonus"] != nil {
		socketBonus = extern_stats.SimJsonArrayToGearStatBlock(itemObj["socketBonus"].([]any))
	}

	scalingOptions := itemObj["scalingOptions"].(map[string]any)
	baseItemLevel := getUInt16OrPanic(scalingOptions["0"].(map[string]any), "ilvl")
	for scaleGroup, entry := range scalingOptions {

		scaleEntry := entry.(map[string]any)
		itemLevel := getUInt16OrPanic(scaleEntry, "ilvl")

		var scaleStats stats.StatBlock
		if scaleEntry["stats"] != nil {
			scaleStats = extern_stats.SimJsonMapToGearStatBlock(scaleEntry["stats"].(map[string]any))
		}

		var upgradeLevel items.UpgradeLevel
		if scaleGroup == "-1" {
			upgradeLevel = -1
		} else {
			upgradeLevel = items.CalcUpgradeLevel(itemLevel, baseItemLevel)
		}

		item := items.FullItem_FromWowSim(itemId, itemLevel, baseItemLevel, upgradeLevel, slot, name, scaleStats, armorType, socketSlots, socketBonus, phase)
		itemsById[itemId] = append(itemsById[itemId], item)
	}
}

func convertSockets(jsonSockets []any) []stats.SocketType {
	gemSockets := make([]stats.SocketType, 0, len(jsonSockets))
	for _, num := range jsonSockets {
		sock := convertSocket(num)
		gemSockets = append(gemSockets, sock)
	}
	return gemSockets
}

func convertSocket(num any) stats.SocketType {
	return stats.SocketType(num.(float64))
}

func convertArmorType(num int32) stats.ArmorType {
	return stats.ArmorType(num)
}

func convertReforge(reforgeArray []any) {
	for _, element := range reforgeArray {
		reforegeObj := element.(map[string]any)
		addReforge(reforegeObj)
	}
}

func addReforge(reforgeObj map[string]any) {
	id := getUInt16OrPanic(reforgeObj, "id")

	from := getAnyIntOrPanic(reforgeObj, "fromStat")
	fromStat := extern_stats.SimStatIndexToGearStatThrows(from)

	to := getAnyIntOrPanic(reforgeObj, "toStat")
	toStat := extern_stats.SimStatIndexToGearStatThrows(to)

	reforge := stats.ReforgeRecipe_of(fromStat, toStat)
	reforgeById[id] = reforge
	reforgeByObj[reforge] = id
}

func getUInt32OrPanic(obj map[string]any, key string) uint32 {
	value, ok := obj[key]
	if ok {
		return uint32(value.(float64))
	} else {
		panic("json key not found " + key)
	}
}

func getUInt16OrPanic(obj map[string]any, key string) uint16 {
	value, ok := obj[key]
	if ok {
		return uint16(value.(float64))
	} else {
		panic("json key not found " + key)
	}
}

func getAnyIntOrPanic(obj map[string]any, key string) int {
	value, ok := obj[key]
	if ok {
		return int(value.(float64))
	} else {
		panic("json key not found " + key)
	}
}

func getIntOrDefault(obj map[string]any, key string, defaultValue int32) int32 {
	value, ok := obj[key]
	if ok {
		return int32(value.(float64))
	} else {
		return defaultValue
	}
}
