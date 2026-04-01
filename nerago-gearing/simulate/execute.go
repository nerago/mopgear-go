package simulate

import (
	"fmt"
	"log"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	gear_stat "paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
	wowsim_core "github.com/wowsims/mop/sim/core"
	wowsim_proto "github.com/wowsims/mop/sim/core/proto"
	wowsim_stat "github.com/wowsims/mop/sim/core/stats"
	wowsim_protojson "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type WowSim_RunSize int32

const (
	RunSize_TestOnly     WowSim_RunSize = 100
	RunSize_QuickDirty   WowSim_RunSize = 20000
	RunSize_Medium       WowSim_RunSize = 100000
	RunSize_SlowAccurate WowSim_RunSize = 500000
)

func WowSim_Execute(runSize WowSim_RunSize, spec gear_stat.SpecType, equipMap *items.FullEquipMap, model *model.Model, bonusStats *gear_stat.StatBlock, tracker *util.TrackProgress) SimResultStats {
	infile := files.SimFileFor(spec)
	var input wowsim_proto.RaidSimRequest
	loadAnyProtoFile(&input, infile)

	updateGear(&input, equipMap, model)
	updateBonus(&input, bonusStats)
	// updateRotation(&input, spec)
	// updateHealRate(&input, spec)
	input.SimOptions.Iterations = int32(runSize)

	reporter := make(chan *wowsim_proto.ProgressMetrics, 10)

	id := uuid.NewString()
	input.RequestId = id

	wowsim_core.RunRaidSimConcurrentAsync(&input, reporter, "gearing-"+id)

	finalResult := waitForResult(reporter, tracker)
	return convertResult(finalResult)
}

func updateHealRate(input *wowsim_proto.RaidSimRequest, spec gear_stat.SpecType) {
	var healRate float64
	if spec == gear_stat.Spec_PaladinProtDps {
		healRate = 45000
	} else {
		healRate = 5000
	}
	input.Raid.Parties[0].Players[0].HealingModel.Hps = healRate
}

func updateRotation(input *wowsim_proto.RaidSimRequest, spec gear_stat.SpecType) {
	if spec == gear_stat.Spec_PaladinProtDps || spec == gear_stat.Spec_PaladinProtMitigation {
		var rotation wowsim_proto.APLRotation
		loadAnyProtoFile(&rotation, files.PaladinProtRotation)
		input.Raid.Parties[0].Players[0].Rotation = &rotation
	}
}

func updateGear(input *wowsim_proto.RaidSimRequest, equipMap *items.FullEquipMap, model *model.Model) {
	if equipMap == nil {
		return
	}

	itemSpecArray := make([]*wowsim_proto.ItemSpec, 0, 16)

	for item := range equipMap.AllItemSeq() {
		spec := wowsim_proto.ItemSpec{}
		spec.Id = int32(item.Ref.ItemId)
		spec.UpgradeStep = wowsim_proto.ItemLevelState(item.Ref.UpgradeLevel)
		if !item.Reforge.IsEmpty() {
			spec.Reforging = int32(db.WowSimDB_ReforgeToId(item.Reforge))
		}

		spec.Gems = make([]int32, len(item.GemChoice))
		for i := range item.GemChoice {
			spec.Gems[i] = int32(item.GemChoice[i].Id)
		}

		if item.EnchantChoice != 0 {
			spec.Enchant = int32(item.EnchantChoice)
		}

		if item.RandomSuffix != 0 {
			spec.RandomSuffix = item.RandomSuffix
		}

		if item.Slot == items.Item_Hand && model.IsEngineer {
			spec.Tinker = 4898
		}

		itemSpecArray = append(itemSpecArray, &spec)
	}

	input.Raid.Parties[0].Players[0].Equipment.Items = itemSpecArray
}

func updateBonus(input *wowsim_proto.RaidSimRequest, bonusStats *gear_stat.StatBlock) {
	if bonusStats == nil {
		return
	}

	unitStats := wowsim_proto.UnitStats{}

	for index := range bonusStats {
		stat := gear_stat.StatType(index)
		theirIndex := mapStat(stat)
		unitStats.Stats[theirIndex] = float64(bonusStats[stat])
	}

	input.Raid.Parties[0].Players[0].BonusStats = &unitStats
}

func mapStat(stat gear_stat.StatType) wowsim_stat.Stat {
	switch stat {
	case gear_stat.Stat_Strength:
		return wowsim_stat.Strength
	case gear_stat.Stat_Agility:
		return wowsim_stat.Agility
	case gear_stat.Stat_Stamina:
		return wowsim_stat.Stamina
	case gear_stat.Stat_Intellect:
		return wowsim_stat.Intellect
	case gear_stat.Stat_Spirit:
		return wowsim_stat.Spirit
	case gear_stat.Stat_Hit:
		return wowsim_stat.HitRating
	case gear_stat.Stat_Crit:
		return wowsim_stat.CritRating
	case gear_stat.Stat_Haste:
		return wowsim_stat.HasteRating
	case gear_stat.Stat_Expertise:
		return wowsim_stat.ExpertiseRating
	case gear_stat.Stat_Dodge:
		return wowsim_stat.DodgeRating
	case gear_stat.Stat_Parry:
		return wowsim_stat.ParryRating
	case gear_stat.Stat_Mastery:
		return wowsim_stat.MasteryRating
	default:
		panic("unknown that")
	}
}

func waitForResult(reporter chan *wowsim_proto.ProgressMetrics, tracker *util.TrackProgress) *wowsim_proto.RaidSimResult {
	if tracker != nil {
		push := tracker.PrepareForPush()
		for v := range reporter {
			if v.FinalRaidResult != nil {
				tracker.Stop()
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

func printResult(finalResult *wowsim_proto.RaidSimResult) {
	output, err := wowsim_protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(finalResult)
	if err != nil {
		log.Fatalf("failed to marshal final results: %s", err)
	}

	fmt.Print(string(output))
}

func convertResult(finalResult *wowsim_proto.RaidSimResult) SimResultStats {
	if finalResult.Error != nil {
		panic("sim fail = " + finalResult.Error.Message)
	} else if finalResult != nil && finalResult.RaidMetrics != nil && finalResult.RaidMetrics.Parties != nil && finalResult.RaidMetrics.Parties[0] != nil && finalResult.RaidMetrics.Parties[0].Players != nil && finalResult.RaidMetrics.Parties[0].Players[0] != nil {
		playerMetrics := finalResult.RaidMetrics.Parties[0].Players[0]
		return SimResultStats{DPS: playerMetrics.Dps.Avg, TPS: playerMetrics.Threat.Avg, DTPS: playerMetrics.Dtps.Avg, TMI: playerMetrics.Tmi.Avg, HPS: playerMetrics.Hps.Avg, DEATH: playerMetrics.ChanceOfDeath}
	} else {
		panic("incomplete sim result")
	}
}

func loadAnyProtoFile[T proto.Message](object T, filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to load input json file %q: %v", filename, err)
	}

	err = wowsim_protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, object)
	if err != nil {
		log.Fatalf("failed to load input json file: %s", err)
	}
}
