package tmf

import (
	"fmt"
	"strconv"
	"strings"
)

// BaseMaterials is a resource that groups one or more named, colored Base
// materials. Triangles and Objects reference a BaseMaterials by its id and
// an index into the slice returned by Materials.
type BaseMaterials struct {
	id        uint32
	materials []BaseMaterial
}

// BaseMaterial is a single named material with an optional display color.
type BaseMaterial struct {
	Name         string
	DisplayColor Color
}

// NewBaseMaterials constructs a BaseMaterials resource.
func NewBaseMaterials(id uint32, mats ...BaseMaterial) *BaseMaterials {
	return &BaseMaterials{id: id, materials: append([]BaseMaterial(nil), mats...)}
}

// ID returns the resource id.
func (b *BaseMaterials) ID() uint32 { return b.id }

// SetID sets the resource id.
func (b *BaseMaterials) SetID(id uint32) { b.id = id }

// Materials returns the contained materials slice.
func (b *BaseMaterials) Materials() []BaseMaterial { return b.materials }

// Append appends a single material and returns its zero-based index.
func (b *BaseMaterials) Append(m BaseMaterial) uint32 {
	b.materials = append(b.materials, m)
	return uint32(len(b.materials) - 1)
}

// Color is an sRGB color with 8-bit channels and an alpha channel. A zero
// Color (R=G=B=A=0) is treated as "unspecified" so that 3MF's distinction
// between "no color" and "fully transparent black" can be preserved.
type Color struct {
	R, G, B, A uint8
	Set        bool
}

// NewColor returns an opaque Color with the given channels marked as set.
func NewColor(r, g, b uint8) Color { return Color{R: r, G: g, B: b, A: 0xff, Set: true} }

// NewColorRGBA returns a Color with the given channels marked as set.
func NewColorRGBA(r, g, b, a uint8) Color { return Color{R: r, G: g, B: b, A: a, Set: true} }

// String formats the color as a 3MF #RRGGBB or #RRGGBBAA string. An unset
// color formats to the empty string.
func (c Color) String() string {
	if !c.Set {
		return ""
	}
	if c.A == 0xff {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// ParseColor parses a 3MF color string of the form "#RRGGBB" or "#RRGGBBAA"
// (case-insensitive). An empty string returns an unset Color.
func ParseColor(s string) (Color, error) {
	if s == "" {
		return Color{}, nil
	}
	if s[0] != '#' || (len(s) != 7 && len(s) != 9) {
		return Color{}, fmt.Errorf("tmf: invalid color %q", s)
	}
	hex := strings.ToUpper(s[1:])
	parseByte := func(off int) (uint8, error) {
		v, err := strconv.ParseUint(hex[off:off+2], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("tmf: invalid color %q: %w", s, err)
		}
		return uint8(v), nil
	}
	r, err := parseByte(0)
	if err != nil {
		return Color{}, err
	}
	g, err := parseByte(2)
	if err != nil {
		return Color{}, err
	}
	b, err := parseByte(4)
	if err != nil {
		return Color{}, err
	}
	a := uint8(0xff)
	if len(hex) == 8 {
		a, err = parseByte(6)
		if err != nil {
			return Color{}, err
		}
	}
	return Color{R: r, G: g, B: b, A: a, Set: true}, nil
}
