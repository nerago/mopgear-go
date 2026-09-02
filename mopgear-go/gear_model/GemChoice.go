package gear_model

import (
	"maps"

	. "github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/items"
	. "github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

type GemChoice map[SocketType]GemInfo

func (gems GemChoice) Equals(other GemChoice) bool {
	return maps.Equal(gems, other)
}

func (gems GemChoice) GetChoice(socket SocketType) *GemInfo {
	info, ok := gems[socket]
	if ok {
		return &info
	} else {
		return nil
	}
}

func (gems GemChoice) ValidateMetaGemInItemSet(itemSet *items.FullItemSet) error {
	head := itemSet.Items().Get(items.Equip_Head)
	return gems.ValidateMetaGemInItem(head)
}

func (gems GemChoice) ValidateMetaGemInItem(head *items.FullItem) error {
	if head == nil {
		return util.ErrorTracedNew("no item in head slot")
	}

	if head.SlotItem() != items.Item_Head {
		return util.ErrorTracedNew("not really a head item")
	}

	headGems := head.GemChoice()
	if len(headGems) == 0 {
		return util.ErrorTracedNew("head item with no gems " + head.CreateString())
	}

	firstGem := headGems[0]
	if !GemData_IsMeta(&firstGem) {
		return util.ErrorTracedNew("first gem is not meta gem in " + head.CreateString())
	}

	expected := gems.GetChoice(Socket_Meta)
	if expected == nil {
		return util.ErrorTracedNew("model doesn't specify meta gem")
	} else if firstGem.Id != expected.Id {
		return util.ErrorTracedNew("meta gem doesn't match spec requirement " + head.CreateString())
	}

	return nil
}

func GemChoice_ForSpec(spec SpecType, goal OptimiseGoal) map[SocketType]GemInfo {
	result := make(map[SocketType]GemInfo)
	switch spec {

	case Spec_PaladinProt:
		switch goal {
		case OptimiseGoal_Mitigation, OptimiseGoal_HalfMitiHeal:
			result[Socket_Red] = GemData_ById(76667)
			result[Socket_Blue] = GemData_ById(76588)
			result[Socket_Yellow] = GemData_ById(76699)
			result[Socket_General] = GemData_ById(76699)
			result[Socket_Meta] = GemData_ById(95344)
			result[Socket_Engineering] = GemData_ById(77542)
			result[Socket_Sha] = GemData_ById(89881)

		case OptimiseGoal_Dps, OptimiseGoal_HalfMitiDps:
			result[Socket_Red] = GemData_ById(76667)
			result[Socket_Blue] = GemData_ById(76588)
			result[Socket_Yellow] = GemData_ById(76699)
			result[Socket_General] = GemData_ById(76699)
			result[Socket_Meta] = GemData_ById(95346)
			result[Socket_Engineering] = GemData_ById(77542)
			result[Socket_Sha] = GemData_ById(89881)

		default:
			panic("GemChoice not known")
		}

	case Spec_PaladinRet:
		result[Socket_Red] = GemData_ById(76667)
		result[Socket_Blue] = GemData_ById(76588)
		result[Socket_Yellow] = GemData_ById(76699)
		result[Socket_General] = GemData_ById(76699)
		result[Socket_Meta] = GemData_ById(95346)
		result[Socket_Engineering] = GemData_ById(77542)
		result[Socket_Sha] = GemData_ById(89881)

	default:
		panic("GemChoice not known")
	}
	return result
}
