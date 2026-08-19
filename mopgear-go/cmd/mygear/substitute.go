package mygear

import (
	"slices"

	"github.com/nerago/mopgear-go/items"
)

var extrasSetSpecific = []items.ItemId{
	86979,
	95142,
	95153,
	95281,
	95291,
	96375,
	96550,
	100644,
	104993,
	86955,
	96657,
	96667,
	104938,
	105090,
	94773,
	95140,
	96394,
	96447,
	96468,
	103791,
	95535,
	96478,
	96533,
	105033,
}

var substituteItemsCommon = slices.Concat(
	extrasSetSpecific,
	LegendCloaks, MiscOtherP3,
	RetT15,
	RetT16,
	ProtT15,
	ProtT16,
	Celestial,
	CelestialRaden,
	OrgRaidDrops,
)
var SubstituteItemsRet = slices.Concat(substituteItemsCommon)

var SubstituteItemsProt = slices.Concat(substituteItemsCommon, OrgOneHandAndShield)

var IgnoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat
