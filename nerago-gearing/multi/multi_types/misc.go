package multi_types

type ForceItemMode int8

const (
	Force_UnknownTODO              ForceItemMode = iota
	Force_ForbiddenTODO            ForceItemMode = iota
	Force_OptionalTODO             ForceItemMode = iota
	Force_FixedWhereAvailableTODO  ForceItemMode = iota
	Force_RequireAtLeastOneUseTODO ForceItemMode = iota
	Force_RequiredAlwaysTODO       ForceItemMode = iota
)
