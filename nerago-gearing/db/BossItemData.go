package db

import (
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/items"
	"os"
	"strings"
)

var BossItemData_NamesInOrder = []string{
	"Dogs MSV",
	"Feng MSV",
	"Gara'jal MSV",
	"Kings MSV",
	"Elegon MSV",
	"Will MSV",
	"Vizier HOF",
	"Blade Lord HOF",
	"Garalon HOF",
	"Wind Lord HOF",
	"Amber-Shaper HOF",
	"Empress HOF",
	"Protectors ToES",
	"Tsulong ToES",
	"Lei Shi ToES",
	"Sha ToES",
	"Jinrokh ToT",
	"Horridon ToT",
	"Council ToT",
	"Tortos ToT",
	"Megaera ToT",
	"Ji-Kun ToT",
	"Durumu ToT",
	"Primordius ToT",
	"Dark Animus ToT",
	"Iron Qon ToT",
	"Twin Consorts ToT",
	"Lei Shen ToT",
	"Raden ToT",
	"SoO Shared",
	"SoO Immerseus",
	"SoO FallenProtectors",
	"SoO Norushen",
	"SoO ShaofPride",
	"SoO Galakras",
	"SoO IronJuggernaut",
	"SoO DarkShaman",
	"SoO Nazgrim",
	"SoO Malkorok",
	"SoO Spoils",
	"SoO Thok",
	"SoO Blackfuse",
	"SoO Paragons",
	"SoO Garrosh"}

var g_itemNameToBoss map[string]string

func load() {
	bytes, err := os.ReadFile(files.BossLookup)
	if err != nil {
		panic(err)
	}
	fullStr := string(bytes)

	itemNameToBoss := make(map[string]string)
	for line := range strings.SplitSeq(fullStr, "\r\n") {
		parts := strings.Split(line, "\t")
		if len(parts) > 2 {
			// itemIdStr := parts[0]
			itemName := parts[1]
			// slot := parts[2]
			bossName := parts[3]
			itemNameToBoss[itemName] = bossName
		}
	}
	g_itemNameToBoss = itemNameToBoss
}

func BossItemData_BossForItem(item *items.FullItem) string {
	if g_itemNameToBoss == nil {
		load()
	}

	return g_itemNameToBoss[item.BaseName()]
}

func BossItemData_BossForItemId(itemId items.ItemId) string {
	if g_itemNameToBoss == nil {
		load()
	}

	itemName := LookupItemNameByItemId(itemId)
	return g_itemNameToBoss[itemName]
}
