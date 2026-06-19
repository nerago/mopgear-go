package loaders

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
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

func BagsFile_PlusPaladinGear_Read() EquippedArray {
	equippedItems := BagsFileReader_Read()
	gearFiles := []string{files.GearFileProtMitigationWithSet, files.GearFileProtMitigationNoSet, files.GearFileProtCompromise, files.GearFileProtHeal, files.GearFileProtDps, files.GearFileRet}
	for _, filename := range gearFiles {
		gear := GearFileReader_Read(filename)
		equippedItems = append(equippedItems, gear...)
	}
	equippedItems = util.RemoveDuplicatesFunc(equippedItems, (*EquippedItem).Equals)
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

func BagsFileItemSetExtraDefaults(equip *EquippedItem, upgradeLevelSuggested items.UpgradeLevel) {
	// shouldn't be needed now we had an alternate export
	if equip.UpgradeStepOrItemLevel == 0 {
		equip.UpgradeStepOrItemLevel = int32(upgradeLevelSuggested)
	}
}
