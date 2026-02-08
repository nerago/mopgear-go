package model

import (
	. "paladin_gearing_go/db"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/stats"
)

type GemChoice map[SocketType]GemInfo

// TODO alternate gems stuff for hit etc

func GemChoice_ForSpec(spec SpecType) map[SocketType]GemInfo {
	result := make(map[SocketType]GemInfo)
	switch spec {

	case Spec_PaladinProtMitigation:
		result[Socket_Red] = GemData_ById(76667)
		result[Socket_Blue] = GemData_ById(76588)
		result[Socket_Yellow] = GemData_ById(76699)
		result[Socket_General] = GemData_ById(76699)
		result[Socket_Meta] = GemData_ById(95344)
		result[Socket_Engineering] = GemData_ById(77542)
		result[Socket_Sha] = GemData_ById(89881)

	case Spec_PaladinProtDps, Spec_PaladinRet:
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
