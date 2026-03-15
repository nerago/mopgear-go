package loaders

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
)

func BagsFileReader_Read() []EquippedItem {
	equippedItems := make([]EquippedItem, 0)

	allBytes, err := os.ReadFile(files.BagsFilename)
	if err != nil {
		panic(err)
	}

	var inputObject map[string]any
	json.Unmarshal(allBytes, &inputObject)

	itemArray := inputObject["items"].([]any)
	for _, element := range itemArray {
		itemObject := element.(map[string]any)
		equip := readEquipped(itemObject)
		equippedItems = append(equippedItems, equip)
	}

	return equippedItems
}
