package tmf

import "fmt"

// Unit identifies the unit of measure used by a 3MF model. The 3MF core
// specification fixes the set of legal values; any other string is rejected
// when parsing.
type Unit int

const (
	UnitUnknown Unit = iota
	UnitMicron
	UnitMillimeter
	UnitCentimeter
	UnitInch
	UnitFoot
	UnitMeter
)

// String returns the canonical 3MF spelling of the unit.
func (u Unit) String() string {
	switch u {
	case UnitMicron:
		return "micron"
	case UnitMillimeter:
		return "millimeter"
	case UnitCentimeter:
		return "centimeter"
	case UnitInch:
		return "inch"
	case UnitFoot:
		return "foot"
	case UnitMeter:
		return "meter"
	}
	return ""
}

// ParseUnit returns the Unit corresponding to s, which must be one of the
// canonical spellings defined by the 3MF core specification.
func ParseUnit(s string) (Unit, error) {
	switch s {
	case "micron":
		return UnitMicron, nil
	case "millimeter":
		return UnitMillimeter, nil
	case "centimeter":
		return UnitCentimeter, nil
	case "inch":
		return UnitInch, nil
	case "foot":
		return UnitFoot, nil
	case "meter":
		return UnitMeter, nil
	}
	return UnitUnknown, fmt.Errorf("tmf: unknown unit %q", s)
}

// ObjectType identifies the role of an Object inside a 3MF model.
type ObjectType int

const (
	ObjectTypeModel ObjectType = iota
	ObjectTypeSolidSupport
	ObjectTypeSupport
	ObjectTypeSurface
	ObjectTypeOther
)

func (t ObjectType) String() string {
	switch t {
	case ObjectTypeSolidSupport:
		return "solidsupport"
	case ObjectTypeSupport:
		return "support"
	case ObjectTypeSurface:
		return "surface"
	case ObjectTypeOther:
		return "other"
	}
	return "model"
}

// ParseObjectType returns the ObjectType corresponding to s. An empty string
// returns ObjectTypeModel (the default per the spec). Unknown strings return
// an error.
func ParseObjectType(s string) (ObjectType, error) {
	switch s {
	case "", "model":
		return ObjectTypeModel, nil
	case "solidsupport":
		return ObjectTypeSolidSupport, nil
	case "support":
		return ObjectTypeSupport, nil
	case "surface":
		return ObjectTypeSurface, nil
	case "other":
		return ObjectTypeOther, nil
	}
	return ObjectTypeModel, fmt.Errorf("tmf: unknown object type %q", s)
}
