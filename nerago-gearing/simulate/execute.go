package simulate

import (
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/stats/extern_stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"

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

func WowSim_Execute_UseModel(runSize WowSim_RunSize, model *model.Model, equipMap *items.FullEquipMap, bonusStats *map[stats.StatType]int32, tracker *util.TrackProgress) stats.SimData {
	return WowSim_Execute_SpecifyAll(runSize, model.SimSpeedUp, model.Spec, model.Goal, model.SimulateAs, model.Professions, equipMap, bonusStats, tracker)
}

func WowSim_Execute_SpecifyAll(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *map[stats.StatType]int32, tracker *util.TrackProgress) stats.SimData {
	input, reporter, id := prepareSim(runSize, speedUp, spec, goal, fight, profession, equipMap, bonusStats)
	wowsim_core.RunRaidSimConcurrentAsync(input, reporter, id)

	finalResult := waitForResult(reporter, tracker)
	return convertResult(finalResult)
}

func WowSim_Execute_SpecifyAll_Future(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *map[stats.StatType]int32, tracker *util.TrackProgress) *channel_op.FutureCancellable[stats.SimData] {
	input, reporter, id := prepareSim(runSize, speedUp, spec, goal, fight, profession, equipMap, bonusStats)
	wowsim_core.RunRaidSimConcurrentAsync(input, reporter, id)

	future := channel_op.FutureCancellable_Make[stats.SimData]()
	future.AddCancelHandler(func() {
		simsignals.AbortById(id)
	})

	go func() {
		finalResult := waitForResult(reporter, tracker)
		converted := convertResult(finalResult)
		future.SetResult(converted)
	}()

	return future
}

func prepareSim(runSize WowSim_RunSize, speedUp int, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo, equipMap *items.FullEquipMap, bonusStats *map[stats.StatType]int32) (*wowsim_proto.RaidSimRequest, chan *wowsim_proto.ProgressMetrics, string) {
	if speedUp != 0 {
		runSize /= WowSim_RunSize(speedUp)
	}

	infile := files.SimFileFor(spec, goal, fight)
	input := inputRequestFromTemplate(infile, equipMap, profession, bonusStats, spec, fight, runSize)

	reporter := make(chan *wowsim_proto.ProgressMetrics, 10)

	id := uuid.NewString()
	input.RequestId = id
	return input, reporter, id
}

func inputRequestFromTemplate(infile string, equipMap *items.FullEquipMap, profession model.ProfessionInfo, bonusStats *map[stats.StatType]int32, spec stats.SpecType, fight stats.WowSim_Fight, runSize WowSim_RunSize) *wowsim_proto.RaidSimRequest {
	var input wowsim_proto.RaidSimRequest
	loadAnyProtoFile(&input, infile)

	updateGear(&input, equipMap, profession)
	updateBonus(&input, bonusStats)
	updateRotation(&input, spec)
	updateFight(&input, fight)
	input.SimOptions.Iterations = int32(runSize)
	input.SimOptions.RandomSeed = 0
	return &input
}

func inputRequestFromScratch(equipMap *items.FullEquipMap, profession model.ProfessionInfo, bonusStats *map[stats.StatType]int32, spec stats.SpecType, fight stats.WowSim_Fight, runSize WowSim_RunSize) *wowsim_proto.RaidSimRequest {
	input := wowsim_proto.RaidSimRequest{
		Type: wowsim_proto.SimType_SimTypeIndividual,
		SimOptions: &wowsim_proto.SimOptions{
			Iterations:          int32(runSize),
			RandomSeed:          1485465806, // meaningless number but gets consistent out
			DebugFirstIteration: false,
		},
	}

	input.Encounter = &wowsim_proto.Encounter{
		Duration:             93,
		DurationVariation:    23,
		ExecuteProportion_20: 0.20,
		ExecuteProportion_25: 0.25,
		ExecuteProportion_35: 0.35,
		ExecuteProportion_45: 0.45,
		ExecuteProportion_90: 0.90,
		Targets: []*wowsim_proto.Target{
			{
				Id:            68476,
				Name:          "Horridon 10 H",
				Level:         93,
				MobType:       wowsim_proto.MobType_MobTypeBeast,
				Stats:         []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 24835, 0, 654205500, 0, 0},
				MinBaseDamage: 491480,
				DamageSpread:  0.5508,
				SwingSpeed:    2,
			},
			{
				Id:            69374,
				Name:          "War-God Jalak 10 H",
				Level:         93,
				MobType:       wowsim_proto.MobType_MobTypeHumanoid,
				Stats:         []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 24835, 0, 26168220, 0, 0},
				MinBaseDamage: 423170,
				DamageSpread:  0.4668,
				SwingSpeed:    2,
			},
		},
	}
	updateFight(&input, fight)

	input.Raid = wowsim_core.SinglePlayerRaidProto(
		wowsim_core.WithSpec(
			&wowsim_proto.Player{
				Race:          wowsim_proto.Race_RaceBloodElf,
				Class:         wowsim_proto.Class_ClassPaladin,
				TalentsString: "313213",
				Glyphs: &wowsim_proto.Glyphs{
					Major1: 41101,
					Major2: 45744,
					Major3: 41096,
					Minor1: 80581,
				},
				Consumables: &wowsim_proto.ConsumesSpec{
					PotId:      76090,
					FlaskId:    76087,
					FoodId:     74656,
					ConjuredId: 5512,
				},
				Buffs: &wowsim_proto.IndividualBuffs{
					TricksOfTheTrade:     true,
					DevotionAuraCount:    2,
					VigilanceCount:       2,
					RallyingCryCount:     1,
					ShatteringThrowCount: 1,
				},
				HealingModel: &wowsim_proto.HealingModel{
					Hps:              5000,
					CadenceSeconds:   0.37,
					CadenceVariation: 1.31,
					AbsorbFrac:       0.14,
					BurstWindow:      6,
				},
				Cooldowns: &wowsim_proto.Cooldowns{
					HpPercentForDefensives: 0.3,
				},
				Profession1:        wowsim_proto.Profession_Engineering,
				Profession2:        wowsim_proto.Profession_Blacksmithing,
				DistanceFromTarget: 10,
				InFrontOfTarget:    true,
				ReactionTimeMs:     400,
				ChannelClipDelayMs: 50,
				Equipment:          nil, // added below
				Rotation:           nil, // added below
			},
			&wowsim_proto.ProtectionPaladin_Options{
				ClassOptions: &wowsim_proto.PaladinOptions{
					Seal: wowsim_proto.PaladinSeal_Insight,
				},
			},
		),
		&wowsim_proto.PartyBuffs{},
		&wowsim_proto.RaidBuffs{
			TrueshotAura:        true,
			SerpentsSwiftness:   true,
			ArcaneBrilliance:    true,
			ElementalOath:       true,
			BlessingOfMight:     true,
			BlessingOfKings:     true,
			PowerWordFortitude:  true,
			Bloodlust:           true,
			StormlashTotemCount: 2,
			SkullBannerCount:    2,
		},
		&wowsim_proto.Debuffs{
			WeakenedBlows:         true,
			PhysicalVulnerability: true,
			WeakenedArmor:         true,
			CurseOfElements:       true,
		})

	updateGear(&input, equipMap, profession)
	updateBonus(&input, bonusStats)
	updateRotation(&input, spec)
	updateTalents(&input, spec, fight)

	return &input
}

func updateFight(input *wowsim_proto.RaidSimRequest, fight stats.WowSim_Fight) {

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
		if input.Raid.Parties[0].Players[0].TalentsString != "113213" {
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

func updateTalents(input *wowsim_proto.RaidSimRequest, spec stats.SpecType, fight stats.WowSim_Fight) {
	switch spec {
	case stats.Spec_PaladinProt:
		if fight == stats.Fight_Juggernaut_OffHealer {
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

func updateRotation(input *wowsim_proto.RaidSimRequest, spec stats.SpecType) {
	var rotation wowsim_proto.APLRotation
	switch spec {
	case stats.Spec_PaladinProt:
		loadAnyProtoFile(&rotation, files.PaladinProtRotation)
	case stats.Spec_PaladinRet:
		loadAnyProtoFile(&rotation, files.PaladinRetRotation)
	default:
		panic("don't know rotation")
	}
	input.Raid.Parties[0].Players[0].Rotation = &rotation
}

func updateGear(input *wowsim_proto.RaidSimRequest, equipMap *items.FullEquipMap, professions model.ProfessionInfo) {
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

func updateBonus(input *wowsim_proto.RaidSimRequest, bonusStats *map[stats.StatType]int32) {
	if bonusStats == nil {
		return
	}

	unitStats := extern_stats.GearStatMapToUnitStats(*bonusStats)
	input.Raid.Parties[0].Players[0].BonusStats = unitStats
}

func waitForResult(reporter chan *wowsim_proto.ProgressMetrics, tracker *util.TrackProgress) *wowsim_proto.RaidSimResult {
	if tracker != nil {
		push := tracker.PrepareForPush()
		for v := range reporter {
			if v.FinalRaidResult != nil {
				tracker.SetDone()
				return v.FinalRaidResult
			}
			progress := float64(v.CompletedIterations) / float64(v.TotalIterations)
			push(progress)
		}
	} else {
		for v := range reporter {
			if v.FinalRaidResult != nil {
				return v.FinalRaidResult
			}
		}
	}
	panic("no final result")
}

func convertResult(finalResult *wowsim_proto.RaidSimResult) stats.SimData {
	if finalResult.Error != nil {
		panic("sim fail = " + finalResult.Error.Message)
	} else if finalResult != nil && finalResult.RaidMetrics != nil && finalResult.RaidMetrics.Parties != nil && finalResult.RaidMetrics.Parties[0] != nil && finalResult.RaidMetrics.Parties[0].Players != nil && finalResult.RaidMetrics.Parties[0].Players[0] != nil {
		playerMetrics := finalResult.RaidMetrics.Parties[0].Players[0]
		WowSimRanDuringCurrentProcess = true

		simData := stats.SimData{}
		simData.Set(stats.Sim_DPS, playerMetrics.Dps.Avg)
		simData.Set(stats.Sim_TPS, playerMetrics.Threat.Avg)
		simData.Set(stats.Sim_DTPS, playerMetrics.Dtps.Avg)
		simData.Set(stats.Sim_TMI, playerMetrics.Tmi.Avg)
		simData.Set(stats.Sim_HPS, playerMetrics.Hps.Avg)
		simData.Set(stats.Sim_DEATH, playerMetrics.ChanceOfDeath)
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
