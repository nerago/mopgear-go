package mygear

import "github.com/nerago/mopgear-go/items"

const (
	TrinketZandSpark = 96398
	TrinketFortZand  = 96793
	TrinketPrimRage  = 94519
	TrinketTwinsGaze = 94529

	TrinketFusionCoreHeroic     = 104463
	TrinketThokTailCelestial    = 105111
	TrinketVialCorruptNormal    = 102306
	TrinketRookUnluckyHeroic    = 104442
	TrinketEyeGalakrasCelestial = 104993
	TrinketSkeerBloodCelestial  = 105134
	TrinketJuggFocusCelestial   = 105016

	LegendMeleeCloak = 102249
	LegendTankCloak  = 102250
)

// TIER
var RetT15 = []items.ItemId{
	//95282, // ret tier15 normal head
	//96658, // ret tier15 shoulder heroic
}
var RetT16 = []items.ItemId{
	99052, // ret t16 chest celestial
	99002, // ret t16 hand celestial
	98985, // ret t16 head celestial
	98987, // ret t16 shoulder celestial
	99139, // ret t16 legs normal
}
var ProtT15 = []items.ItemId{
	//95291, // prot tier15 hand normal
	//96664, // prot tier15 chest heroic
	//96666,  // prot tier15 head heroic - ADDED TO INDIVIDUAL
	//96667, // prot tier15 leg heroic
	//96668, // prot tier15 shoulder heroic
}
var ProtT16 = []items.ItemId{
	99126, // prot t16 chest normal
	99128, // prot t16 head normal
	99129, // prot t16 legs normal
	99130, // prot t16 shoulder normal
	99026, // prot t16 legs celestial
	99027, // prot t16 shoulder celestial
	99028, // prot t16 hand celestial
	99368, // prot t16 chest heroic
}

// TRINKET
var TrinketsDpsP3 = []items.ItemId{
	TrinketZandSpark,
}
var TrinketsTankP3 = []items.ItemId{
	TrinketFortZand,
}

var NewTrinketsDamage = []items.ItemId{
	TrinketThokTailCelestial,
	TrinketFusionCoreHeroic,
	TrinketSkeerBloodCelestial,
}
var NewTrinketsTank = []items.ItemId{
	TrinketVialCorruptNormal,
	TrinketRookUnluckyHeroic,
}

// REMAINING P3
var MiscOtherP3 = []items.ItemId{
	96420, // talisman of angry spirits
	96373, // cloudbreaker belt heroic
	96542, // tidal force treads
	96500, // scaled tyrant heroic
}

// ORGRIMMAR
var Timeless = []items.ItemId{
	101882, // cliffbreaker helm exp/mastery
	101887, // timeless ring haste/mastery. Cliffbreaker Seal of the Faultline. 549 (is upgraded)
	// Cliffbreaker Seal of the Landslide. hit/expertise. 535 (not upgraded)
	// 101947, //  Elder Tortoiseshell Seal of the Mountainbed. 549 (is upgraded)
}
var Celestial = []items.ItemId{
	105011, // Demolisher's Reinforced Belt
}
var CelestialRaden = []items.ItemId{
	95011, // lighting clawfeet
	95022, // Ra-den's Ruinous Ring
}
var OrgRaidDrops = []items.ItemId{
	103787, // poisonbinder girth
	103738, // bubble bracers
	105785, // vanguard burly bracer
	103734, // zoid gauntlets
	103735, // tar-coated gauntlets
	103916, // jugg ignition keys
	104461, // rage-blind greathelm
	104415, // bubble bracer heroic
	103892, // tharnok helm
	105767, // hoodrych chest ordos
	104417, // corruption-rotted gauntlets
	104416, // chest congealed corruption heroic
	103796, // seal kings norm
	103798, // bloodclaw band
	105761, // Partik's Purified Legplates
	104494, // krugruk shoulderplates
	104515, // tar-coated gauntlets heroic
	104440, // sorrowpath signet
	104513, // demo belt heroic
}
var OrgOneHandAndShield = []items.ItemId{
	103826, // xifeng weapon
	103872, // bulwurk of fallen general

	104485, // shield of mockery
	103972, // kilruk sword
	104464, // xifeng heroic
	104560, // bulwurk of fallen general heroic
}

var LegendCloaks = []items.ItemId{LegendTankCloak, LegendMeleeCloak}
