package simulate

import (
	"fmt"
	"log"
	"os"
	"paladin_gearing_go/db"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/stats/extern_stats"
	"paladin_gearing_go/util"
	"slices"
	"strings"

	"github.com/google/uuid"
	wowsim_core "github.com/wowsims/mop/sim/core"
	wowsim_proto "github.com/wowsims/mop/sim/core/proto"
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

func WowSim_Execute(runSize WowSim_RunSize, spec stats.SpecType, equipMap *items.FullEquipMap, profession model.ProfessionInfo, bonusStats *stats.StatBlock, tracker *util.TrackProgress) SimResultStats {
	infile := files.SimFileFor(spec)
	var input wowsim_proto.RaidSimRequest
	loadAnyProtoFile(&input, infile)

	updateGear(&input, equipMap, profession)
	updateBonus(&input, bonusStats)
	updateRotation(&input, spec)
	updateHealRate(&input, spec)
	input.SimOptions.Iterations = int32(runSize)

	reporter := make(chan *wowsim_proto.ProgressMetrics, 10)

	id := uuid.NewString()
	input.RequestId = id

	wowsim_core.RunRaidSimConcurrentAsync(&input, reporter, "gearing-"+id)

	finalResult := waitForResult(reporter, tracker)
	return convertResult(finalResult)
}

func updateHealRate(input *wowsim_proto.RaidSimRequest, spec stats.SpecType) {
	if spec == stats.Spec_PaladinProtDps || spec == stats.Spec_PaladinRet {
		// old Horridon model:
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 45000
	} else {
		// Dark Animus Model:

		// that is a bit high, with 2 mobs rate i just hit for 0.468675/s (40 events in 18.747s)
		// this would imply rate of 0.05, ten times too high
		// for _, target := range input.Encounter.Targets {
		// 	target.SwingSpeed = 0.1
		// 	target.MinBaseDamage *= 0.3
		// }
		// input.Raid.Parties[0].Players[0].HealingModel.Hps = 170000

		for _, target := range input.Encounter.Targets {
			target.SwingSpeed = 0.5
			target.MinBaseDamage *= 1.3
		}
		input.Raid.Parties[0].Players[0].HealingModel.Hps = 220000
	}
}

func updateRotation(input *wowsim_proto.RaidSimRequest, spec stats.SpecType) {
	var rotation wowsim_proto.APLRotation
	switch spec {
	case stats.Spec_PaladinProtDps, stats.Spec_PaladinProtMitigation:
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

		if item.Slot == items.Item_Hand && professions.IsEngineer {
			spec.Tinker = 4898
		}

		itemSpecArray = append(itemSpecArray, &spec)
	}

	input.Raid.Parties[0].Players[0].Equipment.Items = itemSpecArray
}

func updateBonus(input *wowsim_proto.RaidSimRequest, bonusStats *stats.StatBlock) {
	if bonusStats == nil {
		return
	}

	unitStats := extern_stats.GearStatBlockToUnitStats(bonusStats)
	input.Raid.Parties[0].Players[0].BonusStats = unitStats
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
		// parseLogs(finalResult.Logs)
		// readMetrics(playerMetrics)
		return SimResultStats{DPS: playerMetrics.Dps.Avg, TPS: playerMetrics.Threat.Avg, DTPS: playerMetrics.Dtps.Avg, TMI: playerMetrics.Tmi.Avg, HPS: playerMetrics.Hps.Avg, DEATH: playerMetrics.ChanceOfDeath}
	} else {
		panic("incomplete sim result")
	}
}

func readMetrics(unitMetrics *wowsim_proto.UnitMetrics) {
	spellLookup := make(map[int32]string)
	spellLookup[35395] = "Crusader Strike"
	spellLookup[138248] = "Unknown Holy Power"
	spellLookup[498] = "Divine Protection"
	spellLookup[53600] = "Shield Of The Righteous"
	spellLookup[105427] = "Judgment"
	spellLookup[98057] = "Avenger's Shield"
	spellLookup[105809] = "Holy Avenger"

	totalGain := 0.0
	for _, res := range unitMetrics.Resources {
		if res.Type == wowsim_proto.ResourceType_ResourceTypeGenericResource {
			switch res.Id.RawId.(type) {
			case *wowsim_proto.ActionID_SpellId:
				if res.Gain > 0 {
					totalGain += res.Gain
				}
			}
		}
	}

	for _, res := range unitMetrics.Resources {
		if res.Type == wowsim_proto.ResourceType_ResourceTypeGenericResource {
			switch id := res.Id.RawId.(type) {
			case *wowsim_proto.ActionID_SpellId:
				if res.Gain > 0 {
					fmt.Printf("res +%6.0f spell %6d %s\t%f\n", res.Gain, id.SpellId, spellLookup[id.SpellId], res.Gain/totalGain)
				}
			case *wowsim_proto.ActionID_OtherId:
				fmt.Printf("res other %d\n", id.OtherId)
			case *wowsim_proto.ActionID_ItemId:
				fmt.Printf("res item %d\n", id.ItemId)
			}
		}
	}
}

func parseLogs(logText string) {
	for line := range strings.SplitSeq(logText, "\n") {
		if strings.Contains(line, "Gained") && strings.Contains(line, "SecondaryResourceTypeHolyPower ") {
			fmt.Println(line)
		} else if strings.Contains(line, "498") {
			// fmt.Println(line)
		}
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

func addSelfWordOfGlory(rotation *wowsim_proto.APLRotation) {
	// Improves

	actionCastWord := wowsim_proto.APLListItem{
		Action: &wowsim_proto.APLAction{
			Action: &wowsim_proto.APLAction_CastSpell{
				CastSpell: &wowsim_proto.APLActionCastSpell{
					SpellId: &wowsim_proto.ActionID{
						RawId: &wowsim_proto.ActionID_SpellId{
							SpellId: 85673,
						},
					},
				},
			},
			Condition: &wowsim_proto.APLValue{
				Value: &wowsim_proto.APLValue_And{
					And: &wowsim_proto.APLValueAnd{
						Vals: []*wowsim_proto.APLValue{
							{
								Value: &wowsim_proto.APLValue_Cmp{
									Cmp: &wowsim_proto.APLValueCompare{
										Lhs: &wowsim_proto.APLValue{
											Value: &wowsim_proto.APLValue_CurrentHealthPercent{
												CurrentHealthPercent: &wowsim_proto.APLValueCurrentHealthPercent{
													SourceUnit: &wowsim_proto.UnitReference{
														Type: wowsim_proto.UnitReference_Self,
													},
												},
											},
										},
										Op: wowsim_proto.APLValueCompare_OpLt,
										Rhs: &wowsim_proto.APLValue{
											Value: &wowsim_proto.APLValue_Const{
												Const: &wowsim_proto.APLValueConst{
													Val: "50%",
												},
											},
										},
									},
								},
							},
							{
								Value: &wowsim_proto.APLValue_Cmp{
									Cmp: &wowsim_proto.APLValueCompare{
										Lhs: &wowsim_proto.APLValue{
											Value: &wowsim_proto.APLValue_CurrentGenericResource{
												CurrentGenericResource: &wowsim_proto.APLValueCurrentGenericResource{},
											},
										},
										Op: wowsim_proto.APLValueCompare_OpGe,
										Rhs: &wowsim_proto.APLValue{
											Value: &wowsim_proto.APLValue_Const{
												Const: &wowsim_proto.APLValueConst{
													Val: "3",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	shieldRighteousIndex := slices.IndexFunc(rotation.PriorityList, func(item *wowsim_proto.APLListItem) bool {
		if castSpell := item.Action.Action.(*wowsim_proto.APLAction_CastSpell); castSpell != nil {
			if spellId := castSpell.CastSpell.SpellId.RawId.(*wowsim_proto.ActionID_SpellId); spellId != nil {
				return spellId.SpellId == 53600
			}
		}
		return false
	})
	rotation.PriorityList = slices.Insert(rotation.PriorityList, shieldRighteousIndex, &actionCastWord)
}
