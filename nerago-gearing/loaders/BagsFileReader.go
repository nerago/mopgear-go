package loaders

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
)

func BagsFileReader_Read() EquippedArray {
	equippedItems := make(EquippedArray, 0)

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

type EquippedArray []EquippedItem

func (equippedArray *EquippedArray) HasAnyWithItemId(itemId items.ItemId) bool {
	for _, item := range *equippedArray {
		if item.ItemId == itemId {
			return true
		}
	}
	return false
}

func (equippedArray *EquippedArray) GetWithItemId(itemId items.ItemId) *EquippedItem {
	for _, item := range *equippedArray {
		if item.ItemId == itemId {
			return &item
		}
	}
	return nil
}
