package simulate

import (
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/stats/extern_stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"

	"github.com/google/uuid"
	wowsim_core "github.com/wowsims/mop/sim/core"
	wowsim_proto "github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/simsignals"
	wowsim_protojson "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type WowSim_RunSize int32

const (
	RunSize_TestOnly   WowSim_RunSize = 100
	RunSize_QuickDirty WowSim_RunSize = 5000
	RunSize_Common     WowSim_RunSize = 20000
	RunSize_Largish    WowSim_RunSize = 100000
	RunSize_VerySlow   WowSim_RunSize = 500000
)

var WowSimRanDuringCurrentProcess = false

func WowSim_Execute_UseModel(runSize WowSim_RunSize, model *gear_model.SpecModel, equipMap *items.FullEquipMap, bonusStats *stats.StatTypeMap[int32], tracker *util.TrackProgress) stats.SimData {
	return WowSim_Execute_SpecifyAll(runSize, model.SimSpeedUp, model.Spec, model.Goal, model.SimulateAs, model.Professions, equipMap, bonusStats, tracker)
}

func WowSim_Execute_SpecifyAll(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *stats.StatTypeMap[int32], tracker *util.TrackProgress) stats.SimData {
	input, reporter, id := prepareSim(runSize, speedUp, spec, goal, fight, profession, equipMap, bonusStats)
	wowsim_core.RunRaidSimConcurrentAsync(input, reporter, id)

	finalResult, completedIterations := waitForResult(reporter, tracker)
	return convertResult(finalResult, completedIterations)
}

func WowSim_Execute_SpecifyAll_Future(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *stats.StatTypeMap[int32], tracker *util.TrackProgress) *util_async.FutureCancellable[stats.SimData] {
	input, reporter, id := prepareSim(runSize, speedUp, spec, goal, fight, profession, equipMap, bonusStats)
	wowsim_core.RunRaidSimConcurrentAsync(input, reporter, id)

	future := util_async.FutureCancellable_Make[stats.SimData]()
	future.AddCancelHandler(func() {
		simsignals.AbortById(id)
	})

	go func() {
		finalResult, completedIterations := waitForResult(reporter, tracker)
		converted := convertResult(finalResult, completedIterations)
		future.SetResult(converted)
	}()

	return future
}

func prepareSim(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *stats.StatTypeMap[int32]) (*wowsim_proto.RaidSimRequest, chan *wowsim_proto.ProgressMetrics, string) {
	if speedUp != 0 {
		runSize /= WowSim_RunSize(speedUp)
	}

	infile := files.SimFileFor(spec, fight)
	input := inputRequestFromTemplate(infile, equipMap, profession, bonusStats, spec, fight, runSize, goal)

	reporter := make(chan *wowsim_proto.ProgressMetrics, 10)

	id := uuid.NewString()
	input.RequestId = id
	return input, reporter, id
}

func inputRequestFromTemplate(infile string, equipMap *items.FullEquipMap, profession gear_model.ProfessionInfo, bonusStats *stats.StatTypeMap[int32], spec stats.SpecType, fight stats.WowSim_Fight, runSize WowSim_RunSize, goal stats.OptimiseGoal) *wowsim_proto.RaidSimRequest {
	var input wowsim_proto.RaidSimRequest
	loadAnyProtoFile(&input, infile)

	updateGear(&input, equipMap, profession)
	updateBonus(&input, bonusStats)
	updateRotation(&input, spec, fight)
	updateFight(&input, fight, spec)
	updateTalents(&input, spec, fight, goal)
	input.SimOptions.Iterations = int32(runSize)
	input.SimOptions.RandomSeed = 0
	return &input
}

func updateFight(input *wowsim_proto.RaidSimRequest, fight stats.WowSim_Fight, spec stats.SpecType) {

	switch fight {
	case stats.Fight_Horridon_HighHeal:
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 45000

	case stats.Fight_Horridon_LowHeal:
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 0
		for _, target := range input.Encounter.Targets {
			target.MinBaseDamage *= 2.0
		}

	case stats.Fight_Animus:
		for _, target := range input.Encounter.Targets {
			target.SwingSpeed = 0.5
			target.MinBaseDamage *= 1.3
		}
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 220000

	case stats.Fight_Juggernaut_HighHeal:
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 600000

	case stats.Fight_Juggernaut_NoExternalHeal:
		if spec == stats.Spec_PaladinProt && input.Raid.Parties[0].Players[0].TalentsString != "113213" {
			panic("unexpected talent setup")
		}
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 0
		for _, target := range input.Encounter.Targets {
			target.MinBaseDamage *= 2.5
		}

	case stats.Fight_Juggernaut_OffHealer:

		input.Raid.Parties[0].Players[0].HealingModel.Hps = 0
		for _, target := range input.Encounter.Targets {
			target.MinBaseDamage *= 2
		}

	default:
		panic("unknown fight")
	}
}

func updateTalents(input *wowsim_proto.RaidSimRequest, spec stats.SpecType, fight stats.WowSim_Fight, goal stats.OptimiseGoal) {
	switch spec {
	case stats.Spec_PaladinProt:
		if fight == stats.Fight_Juggernaut_OffHealer || goal == stats.OptimiseGoal_Healing || goal == stats.OptimiseGoal_HalfMitiHeal {
			// sacred shield -> eternal flame, execution sentence -> light's hammer
			input.Raid.Parties[0].Players[0].TalentsString = "112212"
		} else {
			input.Raid.Parties[0].Players[0].TalentsString = "113213"
		}

		// my usual
		input.Raid.Parties[0].Players[0].Glyphs = &wowsim_proto.Glyphs{
			Major1: 41104, // Final Wrath
			Major2: 45744, // Alabaster Shield
			Major3: 41096, // Divine Protection
		}

		// wowsim's suggested version
		//input.Raid.Parties[0].Players[0].Glyphs = &wowsim_proto.Glyphs{
		//	Major1: 41101, // Focused Shield
		//	Major2: 45744, // Alabaster Shield
		//	Major3: 41096, // Divine Protection
		//}

		// wowsim suggests minor glyph focused wrath, has no relevant effect in sims
	case stats.Spec_PaladinRet:
	default:
		panic("don't know spec")
	}
}

func updateRotation(input *wowsim_proto.RaidSimRequest, spec stats.SpecType, fight stats.WowSim_Fight) {
	var rotation wowsim_proto.APLRotation
	switch spec {
	case stats.Spec_PaladinProt:
		if fight == stats.Fight_Horridon_HighHeal || fight == stats.Fight_Horridon_LowHeal {
			loadAnyProtoFile(&rotation, files.PaladinProtRotationT15)
		} else {
			loadAnyProtoFile(&rotation, files.PaladinProtRotation)
		}
	case stats.Spec_PaladinRet:
		loadAnyProtoFile(&rotation, files.PaladinRetRotation)
	default:
		panic("don't know rotation")
	}
	input.Raid.Parties[0].Players[0].Rotation = &rotation
}

func updateGear(input *wowsim_proto.RaidSimRequest, equipMap *items.FullEquipMap, professions gear_model.ProfessionInfo) {
	if equipMap == nil {
		return
	}

	itemSpecArray := make([]*wowsim_proto.ItemSpec, 0, items.ITEM_SLOT_COUNT)

	for item := range equipMap.AllItemSeq() {
		spec := wowsim_proto.ItemSpec{}
		spec.Id = int32(item.ItemId())
		spec.UpgradeStep = wowsim_proto.ItemLevelState(item.UpgradeLevel())
		if !item.Reforge().IsEmpty() {
			spec.Reforging = int32(db.WowSimDB_ReforgeToId(item.Reforge()))
		}

		spec.Gems = make([]int32, len(item.GemChoice()))
		for i := range item.GemChoice() {
			spec.Gems[i] = int32(item.GemChoice()[i].Id)
		}

		if item.EnchantChoice() != 0 {
			spec.Enchant = int32(item.EnchantChoice())
		}

		if item.RandomSuffix() != 0 {
			spec.RandomSuffix = int32(item.RandomSuffix())
		}

		if item.SlotItem() == items.Item_Hand && professions.IsEngineer {
			spec.Tinker = 4898
		}

		itemSpecArray = append(itemSpecArray, &spec)
	}

	input.Raid.Parties[0].Players[0].Equipment.Items = itemSpecArray
}

func updateBonus(input *wowsim_proto.RaidSimRequest, bonusStats *stats.StatTypeMap[int32]) {
	if bonusStats == nil {
		return
	}

	unitStats := extern_stats.GearStatMapToUnitStats(*bonusStats)
	input.Raid.Parties[0].Players[0].BonusStats = unitStats
}

func waitForResult(reporter chan *wowsim_proto.ProgressMetrics, tracker *util.TrackProgress) (*wowsim_proto.RaidSimResult, int32) {
	if tracker != nil {
		push := tracker.PrepareForPush()
		for v := range reporter {
			if v.FinalRaidResult != nil {
				tracker.SetDone()
				return v.FinalRaidResult, v.CompletedIterations
			}
			progress := float64(v.CompletedIterations) / float64(v.TotalIterations)
			push(progress)
		}
	} else {
		for v := range reporter {
			if v.FinalRaidResult != nil {
				return v.FinalRaidResult, v.CompletedIterations
			}
		}
	}
	panic("no final result")
}

func convertResult(finalResult *wowsim_proto.RaidSimResult, completedIterations int32) stats.SimData {
	if finalResult.Error != nil {
		panic("sim fail = " + finalResult.Error.Message)
	} else if finalResult != nil && finalResult.RaidMetrics != nil && finalResult.RaidMetrics.Parties != nil && finalResult.RaidMetrics.Parties[0] != nil && finalResult.RaidMetrics.Parties[0].Players != nil && finalResult.RaidMetrics.Parties[0].Players[0] != nil {
		WowSimRanDuringCurrentProcess = true
		playerMetrics := finalResult.RaidMetrics.Parties[0].Players[0]

		simData := stats.SimData{}
		simData.SetDetailed(stats.Sim_DPS, playerMetrics.Dps.Avg, playerMetrics.Dps.Min, playerMetrics.Dps.Max, playerMetrics.Dps.Stdev)
		simData.SetDetailed(stats.Sim_TPS, playerMetrics.Threat.Avg, playerMetrics.Threat.Min, playerMetrics.Threat.Max, playerMetrics.Threat.Stdev)
		simData.SetDetailed(stats.Sim_DTPS, playerMetrics.Dtps.Avg, playerMetrics.Dtps.Min, playerMetrics.Dtps.Max, playerMetrics.Dtps.Stdev)
		simData.SetDetailed(stats.Sim_TMI, playerMetrics.Tmi.Avg, playerMetrics.Tmi.Min, playerMetrics.Tmi.Max, playerMetrics.Tmi.Stdev)
		simData.SetDetailed(stats.Sim_HPS, playerMetrics.Hps.Avg, playerMetrics.Hps.Min, playerMetrics.Hps.Max, playerMetrics.Hps.Stdev)
		simData.Set(stats.Sim_DEATH, playerMetrics.ChanceOfDeath)
		simData.SimIterations = playerMetrics.Dps.AggregatorData.N
		return simData
	} else {
		panic("incomplete sim result")
	}
}

func loadAnyProtoFile[T proto.Message](object T, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = wowsim_protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, object)
	if err != nil {
		panic(err)
	}
}
