var apl={
	"type": "TypeAPL",
	"simple": {
	"cooldowns": {
		"hpPercentForDefensives": 0.3
	}
	},
	"prepullActions": [
	// Sacred shield 
	{"action":{"castSpell":{"spellId":{"spellId":20925}}},"doAtValue":{"const":{"val":"-1.6s"}}},

	// potion
	{"action":{"castSpell":{"spellId":{"otherId":"OtherActionPotion"}}},"doAtValue":{"const":{"val":"-0.1s"}}}

	],
	"priorityList": [
	// AvengingWrath
	{"action":{"condition":{"cmp":{"op":"OpGt","lhs":{"currentTime":{}},"rhs":{"const":{"val":"36s"}}}},"castSpell":{"spellId":{"spellId":31884}}}},

	// DivineProtection
	{"action":{"condition":{"and":
		{"vals":[{"or":
			{"vals":[{"and":
				{"vals":[{"bossCurrentTarget":{"targetUnit":{"type":"Target"}}},{"cmp":{"op":"OpLe","lhs":{"bossSpellTimeToReady":{"targetUnit":{"type":"Target"},"spellId":{"spellId":136767}}},"rhs":{"const":{"val":"1s"}}}}]}},{"bossSpellIsCasting":{"targetUnit":{"type":"Target"},"spellId":{"spellId":137458}}},{"cmp":{"op":"OpLe","lhs":{"bossSpellTimeToReady":{"targetUnit":{"type":"Target","index":1},"spellId":{"spellId":136817}}},"rhs":{"const":{"val":"1s"}}}}]}}]}},"castSpell":{"spellId":{"spellId":498}}}},

	// Guardian
	{"action":{"condition":{"and":{"vals":[{"not":{"val":{"spellIsReady":{"spellId":{"spellId":498}}}}},{"auraIsInactive":{"auraId":{"spellId":498},"includeReactionTime":true}},{"cmp":{"op":"OpLe","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%"}}}},{"or":{"vals":[{"and":{"vals":[{"bossCurrentTarget":{"targetUnit":{"type":"Target"}}},{"cmp":{"op":"OpEq","lhs":{"bossSpellTimeToReady":{"targetUnit":{"type":"Target"},"spellId":{"spellId":136767}}},"rhs":{"const":{"val":"0s"}}}}]}},{"cmp":{"op":"OpEq","lhs":{"bossSpellTimeToReady":{"targetUnit":{"type":"Target","index":1},"spellId":{"spellId":136817}}},"rhs":{"const":{"val":"0s"}}}}]}}]}},"castSpell":{"spellId":{"spellId":86659}}}},

	// ArdentDefender
	{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpLe","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"60%"}}}},{"or":{"vals":[{"bossCurrentTarget":{"targetUnit":{"type":"Target"}}},{"bossCurrentTarget":{"targetUnit":{"type":"Target","index":1}}}]}},{"auraIsInactive":{"auraId":{"spellId":86659},"includeReactionTime":true}},{"not":{"val":{"spellIsReady":{"spellId":{"spellId":86659}}}}}]}},"castSpell":{"spellId":{"spellId":31850}}}},

	// potion
	{"action":{"schedule":{"schedule":"42s","innerAction":{"castSpell":{"spellId":{"otherId":"OtherActionPotion"}}}}}},

	// old trinket
	{"action":{"schedule":{"schedule":"0s,60s","innerAction":{"castSpell":{"spellId":{"itemId":89079}}}}}},

	// Vigilance (via druid?)
	{"action":{"condition":{"and":{"vals":[{"auraIsInactive":{"auraId":{"spellId":132403},"includeReactionTime":true}},{"cmp":{"op":"OpGe","lhs":{"currentTime":{}},"rhs":{"const":{"val":"76s"}}}}]}},"castSpell":{"spellId":{"spellId":114030,"tag":-1}}}},

	// Healthstone
	{"action":{"condition":{"and":{"vals":[{"or":{"vals":[{"bossCurrentTarget":{"targetUnit":{"type":"Target"}}},{"bossCurrentTarget":{"targetUnit":{"type":"Target","index":1}}}]}},{"cmp":{"op":"OpLt","lhs":{"currentHealthPercent":{}},"rhs":{"const":{"val":"50%"}}}}]}},"castSpell":{"spellId":{"itemId":5512}}}},

	// "other cooldowns"
	{"action":{"autocastOtherCooldowns":{}}},

	// Use Word of Glory when low
	// Improves Death numbers nicely, but overall DTPS, DPS and TMI worse
	// Could be better on Dark Animus zerg though
 	{
		"action": {
			"condition": {
				"and": {
					"vals": [
						{
							"cmp": {
								"op": "OpLt",
								"lhs": { "currentHealthPercent": {} },
								"rhs": { "const": { "val": "50%" } }
							}
						},
						{
							"cmp": {
								"op": "OpGe",
								"lhs": { "currentGenericResource": {} },
								"rhs": { "const": { "val": "3" } }
							}
						}
					]
				}
			},
			"castSpell": { "spellId": { "spellId": 85673 } }
		}
	},

	// ShieldOfTheRighteous
	{"action":{"castSpell":{"spellId":{"spellId":53600}}}},

	// HolyAvenger
	{"action":{"condition":{"cmp":{"op":"OpGe","lhs":{"currentTime":{}},"rhs":{"const":{"val":"14s"}}}},"castSpell":{"spellId":{"spellId":105809}}}},

	// CrusaderStrike
	{"action":{"castSpell":{"spellId":{"spellId":35395}}}},
	{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGt","lhs":{"spellTimeToReady":{"spellId":{"spellId":35395}}},"rhs":{"const":{"val":"0"}}}},{"cmp":{"op":"OpLe","lhs":{"spellTimeToReady":{"spellId":{"spellId":35395}}},"rhs":{"const":{"val":"0.5s"}}}}]}},"wait":{"duration":{"spellTimeToReady":{"spellId":{"spellId":35395}}}}}},

	// Judgement
	{"action":{"castSpell":{"spellId":{"spellId":20271}}}},
	{"action":{"condition":{"and":{"vals":[{"cmp":{"op":"OpGt","lhs":{"spellTimeToReady":{"spellId":{"spellId":20271}}},"rhs":{"const":{"val":"0"}}}},{"cmp":{"op":"OpLe","lhs":{"spellTimeToReady":{"spellId":{"spellId":20271}}},"rhs":{"const":{"val":"0.5s"}}}},{"cmp":{"op":"OpGe","lhs":{"math":{"op":"OpSub","lhs":{"spellTimeToReady":{"spellId":{"spellId":35395}}},"rhs":{"spellTimeToReady":{"spellId":{"spellId":20271}}}}},"rhs":{"const":{"val":"0.5s"}}}}]}},"wait":{"duration":{"spellTimeToReady":{"spellId":{"spellId":20271}}}}}},

	// Avengers shield
	{"action":{"castSpell":{"spellId":{"spellId":31935}}}},

	// Sacred shield
	{"action":{"condition":{"cmp":{"op":"OpLt","lhs":{"auraRemainingTime":{"auraId":{"spellId":20925}}},"rhs":{"const":{"val":"5s"}}}},"castSpell":{"spellId":{"spellId":20925}}}},

	// HolyWrath
	{"action":{"castSpell":{"spellId":{"spellId":119072}}}},

	// ExecutionSentence
	{"action":{"castSpell":{"spellId":{"spellId":114916}}}},

	// HammerOfWrath
	{"action":{"castSpell":{"spellId":{"spellId":24275}}}},

	// Consecration
	{"action":{"condition":{"not":{"val":{"dotIsActive":{"spellId":{"spellId":26573}}}}},"castSpell":{"spellId":{"spellId":26573}}}},

	// Sacred shield
	{"action":{"castSpell":{"spellId":{"spellId":20925}}}}
	]
}
