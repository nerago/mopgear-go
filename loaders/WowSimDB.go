package loaders

import (
	"encoding/json"
	"fmt"
	"os"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	. "paladin_gearing_go/types/stats"
	"strconv"
)

var loaded = false
var itemsById map[uint32][]FullItem = make(map[uint32][]FullItem)
var itemsByRef map[ItemRef]FullItem = make(map[ItemRef]FullItem)
var reforgeById map[uint16]ReforgeRecipe = make(map[uint16]ReforgeRecipe)
var reforgeByObj map[ReforgeRecipe]uint16 = make(map[ReforgeRecipe]uint16)

func WowSimDB_Read() {
	filename := `C:\Users\nicholas\Dropbox\prog\paladin_gearing\src\main\resources\wowsimdb.json`

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

func WowSimDB_ByIdAndUpgrade(itemId uint32, upgradeLevel int16) *FullItem {
	if !loaded {
		WowSimDB_Read()
	}

	known := itemsById[itemId]
	for _, item := range known {
		if item.Ref.UpgradeLevel() == upgradeLevel {
			return &item
		}
	}

	return nil
}

func convertItems(itemArray []any) {
	for _, element := range itemArray {
		itemObj := element.(map[string]any)
		addItem(itemObj)
	}
}

func addItem(itemObj map[string]any) {
	itemId := getUInt32OrPanic(itemObj, "id")
	name := itemObj["name"].(string)
	phase := int8(getIntOrDefault(itemObj, "phase", -1))
	itemType := getIntOrDefault(itemObj, "type", -1)
	if itemType == -1 {
		return
	}

	weaponType := getIntOrDefault(itemObj, "weaponType", 0)
	handType := getIntOrDefault(itemObj, "handType", 0)
	slot := mapSlot(itemType, weaponType, handType)

	armorType := convertArmorType(getIntOrDefault(itemObj, "armorType", -1))
	socketSlots := convertSockets(itemObj["gemSockets"].([]any))
	socketBonus := convertStatsFromFlat(itemObj["socketBonus"].([]any))

	scalingOptions := itemObj["scalingOptions"].(map[string]any)
	baseItemLevel := getUInt16OrPanic(scalingOptions["0"].(map[string]any), "ilvl")
	for _, entry := range scalingOptions {
		scaleEntry := entry.(map[string]any)
		itemLevel := getUInt16OrPanic(scaleEntry, "ilvl")
		scaleStats := convertStatsFromMap(scaleEntry)

		itemRef := ItemRef{
			ItemId:        itemId,
			ItemLevel:     itemLevel,
			ItemLevelBase: baseItemLevel}
		item := FullItem_FromWowSim(itemRef, slot, name, scaleStats, armorType, socketSlots, socketBonus, phase)
		itemsById[itemId] = append(itemsById[itemId], item)
		itemsByRef[itemRef] = item
	}
}

func convertStatsFromFlat(input []any) StatBlock {
	block := StatBlock{}
	for indexNum, value := range input {
		stat := simBlockIndexToStat(indexNum)
		if stat != Stat_Invalid {
			block[stat] = uint32(value.(float64))
		}
	}
	return block
}

func convertStatsFromMap(input map[string]any) StatBlock {
	block := StatBlock{}
	for indexStr, value := range input {
		indexNum, err := strconv.Atoi(indexStr)
		if err != nil {
			panic(err)
		}

		stat := simBlockIndexToStat(indexNum)
		if stat != Stat_Invalid {
			block[stat] = uint32(value.(float64))
		}
	}
	return block
}

func simBlockIndexToStat(num int) StatType {
	// this may be a one-to-one for now, rather not rely on it
	switch num {
	case 0:
		return Stat_Strength
	case 1:
		return Stat_Agility
	case 3:
		return Stat_Intellect
	case 2:
		return Stat_Stamina
	case 4:
		return Stat_Spirit
	case 5:
		return Stat_Hit
	case 6:
		return Stat_Crit
	case 7:
		return Stat_Haste
	case 8:
		return Stat_Expertise
	case 9:
		return Stat_Dodge
	case 10:
		return Stat_Parry
	case 11:
		return Stat_Mastery
	case 15, 16:
		return Stat_Invalid
	default:
		panic("unknown stat index " + fmt.Sprint(num))
	}
}

func convertSockets(jsonSockets []any) []SocketType {
	gemSockets := make([]SocketType, 0, len(jsonSockets))
	for _, num := range jsonSockets {
		sock := convertSocket(num)
		gemSockets = append(gemSockets, sock)
	}
	return gemSockets
}

func convertSocket(num any) SocketType {
	return SocketType(num.(float64))
}

func convertArmorType(i int32) ArmorType {
	panic("unimplemented")
}

func mapSlot(itemType, weaponType, handType int32) SlotItem {
	panic("unimplemented")
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
	fromStat := simBlockIndexToStat(from)

	to := getAnyIntOrPanic(reforgeObj, "toStat")
	toStat := simBlockIndexToStat(to)

	reforge := ReforgeRecipe{From: fromStat, To: toStat}
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
