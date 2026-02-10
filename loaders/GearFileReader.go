package loaders

import (
	"encoding/json"
	"os"
)

func GearFileReader_Read(filename string) []EquippedItem {
	equippedItems := make([]EquippedItem, 0)

	allBytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var inputObject map[string]any
	json.Unmarshal(allBytes, &inputObject)

	gearObject := inputObject["gear"].(map[string]any)
	itemArray := gearObject["items"].([]any)
	for _, element := range itemArray {
		itemObject := element.(map[string]any)
		equip := readEquipped(itemObject)
		equippedItems = append(equippedItems, equip)
	}

	return equippedItems
}

func readEquipped(itemObject map[string]any) EquippedItem {
	itemId := uint32(itemObject["id"].(float64))

	gems := make([]uint32, 0)
	if itemObject["gems"] != nil {
		for _, num := range itemObject["gems"].([]any) {
			gems = append(gems, uint32(num.(float64)))
		}
	}

	var enchant uint32
	if itemObject["enchant"] != nil {
		enchant = uint32(itemObject["enchant"].(float64))
	}

	var upgradeStep int16
	if itemObject["upgrade_step"] != nil {
		upgradeStep = int16(itemObject["upgrade_step"].(float64))
	}

	var reforging uint16
	if itemObject["reforging"] != nil {
		reforging = uint16(itemObject["reforging"].(float64))
	}

	var randomSuffix int32
	if itemObject["random_suffix"] != nil {
		randomSuffix = int32(itemObject["random_suffix"].(float64))
	}

	return EquippedItem{
		ItemId:        itemId,
		GemChoice:     gems,
		EnchantChoice: enchant,
		RandomSuffix:  randomSuffix,
		UpgradeStep:   upgradeStep,
		Reforging:     reforging,
	}
}
