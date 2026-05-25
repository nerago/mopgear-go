package model

import (
	"maps"
	. "paladin_gearing_go/db"
	. "paladin_gearing_go/stats"
)

type GemChoice map[SocketType]GemInfo

func (gems GemChoice) Equals(other GemChoice) bool {
	return maps.Equal(gems, other)
}

// TODO alternate gems stuff for hit etc

func (gems GemChoice) GetChoice(socket SocketType) *GemInfo {
	info, ok := gems[socket]
	if ok {
		return &info
	} else {
		return nil
	}
}

func GemChoice_ForSpec(spec SpecType, goal OptimiseGoal) map[SocketType]GemInfo {
	result := make(map[SocketType]GemInfo)
	switch spec {

	case Spec_PaladinProt:
		switch goal {
		case OptimiseGoal_Mitigation:
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
