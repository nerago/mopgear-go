package db

import (
	"encoding/json"
	"iter"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/stats/extern_stats"
	"paladin_gearing_go/util"
	"strconv"
)

var loaded = false
var itemsById map[items.ItemId][]items.FullItem = make(map[items.ItemId][]items.FullItem)
var itemsByRef map[items.ItemRef]items.FullItem = make(map[items.ItemRef]items.FullItem)
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

func WowSimDB_ByIdAndUpgrade(itemId items.ItemId, upgradeLevel int8) *items.FullItem {
	known := itemsById[itemId]
	for _, item := range known {
		if item.Ref.UpgradeLevel == upgradeLevel {
			return &item
		}
	}

	return nil
}

func WowSimDB_ByIdAndUpgrade_AllowFallback(itemId items.ItemId, upgradeLevel int8, printer *util.PrintRecorder) *items.FullItem {
	storedItem := WowSimDB_ByIdAndUpgrade(itemId, upgradeLevel)

	if storedItem == nil && upgradeLevel > 0 {
		storedItem = WowSimDB_ByIdAndUpgrade(itemId, 0)
		if storedItem != nil {
			printer.Printf("NOT FOUND at specified upgrade %d = %s\n", upgradeLevel, storedItem.CreateString())
		}
	}

	if storedItem == nil {
		panic("NOT FOUND at any upgrade level itemid=" + strconv.FormatUint(uint64(itemId), 10))
	}

	return storedItem
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
	slot := extern_stats.MapSlotToGear(itemType, handType)

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

		var itemRef items.ItemRef
		if scaleGroup == "-1" {
			itemRef = items.ItemRef_Challenge(itemId, itemLevel)
		} else {
			itemRef = items.ItemRef_Make(itemId, itemLevel, baseItemLevel)
		}
		item := items.FullItem_FromWowSim(itemRef, slot, name, scaleStats, armorType, socketSlots, socketBonus, phase)
		itemsById[itemId] = append(itemsById[itemId], item)
		itemsByRef[itemRef] = item
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
