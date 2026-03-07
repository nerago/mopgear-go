package simulate

import (
	"fmt"
	"log"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	mystat "paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/google/uuid"
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	theirstat "github.com/wowsims/mop/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

type WowSim_RunSize int32

const (
	QuickDirty   WowSim_RunSize = 20000
	Medium       WowSim_RunSize = 100000
	SlowAccurate WowSim_RunSize = 500000
)

type SimResultStats struct {
	DPS, TPS, DTPS, HPS, TMI, DEATH float64
}

func WowSim_Execute(runSize WowSim_RunSize, spec mystat.SpecType, equipMap *items.FullEquipMap, bonusStats *mystat.StatBlock) SimResultStats {
	verbose := true
	infile := exampleFileFor(spec)
	input := loadExampleFile(infile)

	updateGear(input, equipMap)
	updateBonus(input, bonusStats)

	reporter := make(chan *proto.ProgressMetrics, 10)
	id := uuid.NewString()
	core.RunRaidSimConcurrentAsync(input, reporter, "gearing-"+id)

	finalResult := fetchResult(reporter, verbose)
	printResult(finalResult)
	return convertResult(finalResult)
}

func updateGear(input *proto.RaidSimRequest, equipMap *items.FullEquipMap) {
	if equipMap == nil {
		return
	}

	itemSpecArray := make([]*proto.ItemSpec, 0, 16)

	for item := range equipMap.AllItemSeq() {
		spec := proto.ItemSpec{}
		spec.Id = int32(item.Ref.ItemId)
		spec.UpgradeStep = proto.ItemLevelState(item.Ref.UpgradeLevel())
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

		if item.Slot == items.Item_Hand {
			spec.Tinker = 4898
		}

		itemSpecArray = append(itemSpecArray, &spec)
	}

	input.Raid.Parties[0].Players[0].Equipment.Items = itemSpecArray
}

func updateBonus(input *proto.RaidSimRequest, bonusStats *mystat.StatBlock) {
	if bonusStats == nil {
		return
	}

	unitStats := proto.UnitStats{}

	for index := range bonusStats {
		stat := mystat.StatType(index)
		theirIndex := mapStat(stat)
		unitStats.Stats[theirIndex] = float64(bonusStats[stat])
	}

	input.Raid.Parties[0].Players[0].BonusStats = &unitStats
}

func mapStat(stat mystat.StatType) theirstat.Stat {
	switch stat {
	case mystat.Stat_Strength:
		return theirstat.Strength
	case mystat.Stat_Agility:
		return theirstat.Agility
	case mystat.Stat_Stamina:
		return theirstat.Stamina
	case mystat.Stat_Intellect:
		return theirstat.Intellect
	case mystat.Stat_Spirit:
		return theirstat.Spirit
	case mystat.Stat_Hit:
		return theirstat.HitRating
	case mystat.Stat_Crit:
		return theirstat.CritRating
	case mystat.Stat_Haste:
		return theirstat.HasteRating
	case mystat.Stat_Expertise:
		return theirstat.ExpertiseRating
	case mystat.Stat_Dodge:
		return theirstat.DodgeRating
	case mystat.Stat_Parry:
		return theirstat.ParryRating
	case mystat.Stat_Mastery:
		return theirstat.MasteryRating
	default:
		panic("unknown that")
	}
}

func fetchResult(reporter chan *proto.ProgressMetrics, verbose bool) *proto.RaidSimResult {
	var finalResult *proto.RaidSimResult
	for v := range reporter {
		if v.FinalRaidResult != nil {
			finalResult = v.FinalRaidResult
			break
		}
		if verbose {
			fmt.Printf("Sim Progress: %d / %d\n", v.CompletedIterations, v.TotalIterations)
		}
	}
	return finalResult
}

func printResult(finalResult *proto.RaidSimResult) {
	output, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(finalResult)
	if err != nil {
		log.Fatalf("failed to marshal final results: %s", err)
	}

	fmt.Print(string(output))
}

func convertResult(finalResult *proto.RaidSimResult) SimResultStats {
	playerMetrics := finalResult.RaidMetrics.Parties[0].Players[0]
	return SimResultStats{DPS: playerMetrics.Dps.Avg, TPS: playerMetrics.Threat.Avg, DTPS: playerMetrics.Dtps.Avg, TMI: playerMetrics.Tmi.Avg, HPS: playerMetrics.Hps.Avg, DEATH: playerMetrics.ChanceOfDeath}
}

func loadExampleFile(infile string) *proto.RaidSimRequest {
	data, err := os.ReadFile(infile)
	if err != nil {
		log.Fatalf("failed to load input json file %q: %v", infile, err)
	}

	input := &proto.RaidSimRequest{}
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, input)
	if err != nil {
		log.Fatalf("failed to load input json file: %s", err)
	}
	return input
}

func exampleFileFor(spec mystat.SpecType) string {
	switch spec {
	case mystat.Spec_PaladinProtDps:
		return "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-prot-dps.json"
	case mystat.Spec_PaladinProtMitigation:
		return "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-prot-miti.json"
	case mystat.Spec_PaladinRet:
		return "C:\\Users\\nicholas\\Dropbox\\prog\\wow-sim-mop\\example-ret.json"
	default:
		panic("spec not supported")
	}
}

func (stats SimResultStats) Print(printer *util.PrintRecorder) {
	printer.Printf("DPS\t%.2f\n", stats.DPS)
	printer.Printf("TPS\t%.2f\n", stats.TPS)
	printer.Printf("DTPS\t%.2f\n", stats.DTPS)
	printer.Printf("HPS\t%.2f\n", stats.HPS)
	printer.Printf("TMI\t%.2f\n", stats.TMI)
	printer.Printf("DEATH\t%.2f\n", stats.DEATH*100)
}
