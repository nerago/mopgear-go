package tools

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
)

func WowSimJson_Write(equip *items.FullEquipMap, model *gear_model.SpecModel, printer *util.PrintRecorder) string {
	inputFile := model.ReferenceGearFile
	allBytes, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	var mainObject map[string]any
	err = json.Unmarshal(allBytes, &mainObject)
	if err != nil {
		panic(err)
	}

	itemArray := make([]any, 0, items.ITEM_SLOT_COUNT)
	for item := range equip.AllItemSeqPairedConsistent() {
		itemArray = append(itemArray, makeItemObject(item, model.Professions))
	}

	// gear data, wowsim friendly
	gearObject := mainObject["gear"].(map[string]any)
	gearObject["items"] = itemArray

	// equipment data, reforgelite friendly
	equipmentData := make(map[string]any)
	equipmentData["items"] = itemArray
	playerData := make(map[string]any)
	playerData["equipment"] = equipmentData
	mainObject["player"] = playerData

	allBytes, err = json.Marshal(mainObject)
	if err != nil {
		panic(err)
	}

	asText := string(allBytes)
	printer.Println(asText)
	return asText
}

func makeItemObject(item *items.FullItem, profession gear_model.ProfessionInfo) map[string]any {
	object := make(map[string]any)

	object["id"] = item.ItemId()
	object["upgrade_step"] = item.UpgradeLevel()
	if !item.Reforge().IsEmpty() {
		reforgeId := db.WowSimDB_ReforgeToId(item.Reforge())
		object["reforging"] = reforgeId
	}

	if len(item.GemChoice()) > 0 {
		gemArray := make([]any, len(item.GemChoice()))
		for i, gemInfo := range item.GemChoice() {
			gemArray[i] = gemInfo.Id
		}
		object["gems"] = gemArray
	}

	if item.EnchantChoice() != 0 {
		object["enchant"] = item.EnchantChoice()
	}

	//     StatBlock expectedEnchants = GemData.process(item.gemChoice, item.enchantChoice, item.shared.socketSlots(), item.shared.socketBonus(), item.shared.name(), item.slot().possibleBlacksmith());
	//     if (!expectedEnchants.equalsStats(item.statEnchant)) {
	//         throw new RuntimeException("enchant details don't match");
	//     }

	if item.RandomSuffix() != 0 {
		object["random_suffix"] = item.RandomSuffix()
	}

	if item.SlotItem() == items.Item_Hand && profession.IsEngineer {
		object["tinker"] = 4898
	}

	return object
}
