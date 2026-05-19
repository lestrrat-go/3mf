// Package beamlattice implements the 3MF Beam Lattice Extension.
//
// A beam lattice is an alternative geometry attached to a Mesh: instead of
// triangles, it defines beams (line segments between vertices) and
// optional ball joints at the vertices, producing a wireframe structure
// suitable for lightweight, high-strength parts. When a mesh carries a
// beam lattice, the triangle list is typically empty and the lattice is
// the actual printed geometry.
//
// Blank-import the package to register the read/write hooks:
//
//	import _ "github.com/lestrrat-go/3mf/beamlattice"
package beamlattice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lestrrat-go/helium"

	tmf "github.com/lestrrat-go/3mf"
)

// Namespace is the URI of the 3MF Beam Lattice extension.
const Namespace = tmf.NSBeamLattice

// Prefix is the conventional prefix used for the Beam Lattice extension.
const Prefix = tmf.PrefixBeamLattice

// CapMode describes how the end of a beam is closed.
type CapMode int

const (
	CapModeSphere CapMode = iota
	CapModeHemisphere
	CapModeButt
)

func (c CapMode) String() string {
	switch c {
	case CapModeHemisphere:
		return "hemisphere"
	case CapModeButt:
		return "butt"
	}
	return "sphere"
}

func parseCapMode(s string) CapMode {
	switch s {
	case "hemisphere":
		return CapModeHemisphere
	case "butt":
		return CapModeButt
	}
	return CapModeSphere
}

// BeamLattice is the value attached to a Mesh via Mesh.Extension(Namespace).
type BeamLattice struct {
	Radius           float64
	MinLength        float64
	Cap              CapMode
	ClippingMode     string // "none" | "inside" | "outside"
	ClippingMesh     uint32 // object id
	RepresentationMesh uint32 // object id
	Beams            []Beam
	BallSets         []BallSet
}

// Beam is a single beam (line segment between two vertices) with optional
// per-end radius overrides and cap modes.
type Beam struct {
	V1, V2  uint32
	R1, R2  float64 // 0 means "inherit lattice radius"
	Cap1    CapMode
	Cap2    CapMode
	HasCap1 bool
	HasCap2 bool
}

// BallSet groups balls (spheres of a given radius placed at specific
// vertices) into a logical set. Each ball references a vertex index.
type BallSet struct {
	Identifier string
	Name       string
	Balls      []Ball
}

// Ball is a single sphere at the given vertex.
type Ball struct {
	V uint32
	R float64
}

// Of returns the beam lattice attached to mesh, or nil.
func Of(mesh *tmf.Mesh) *BeamLattice {
	if v := mesh.Extension(Namespace); v != nil {
		if bl, ok := v.(*BeamLattice); ok {
			return bl
		}
	}
	return nil
}

// Require declares the beam lattice extension on m using the conventional prefix.
func Require(m *tmf.Model) { m.RequireExtension(Namespace, Prefix) }

type extReader struct{ tmf.BaseExtensionReader }

func (extReader) Namespace() string { return Namespace }

type extWriter struct{ tmf.BaseExtensionWriter }

func (extWriter) Namespace() string { return Namespace }

func init() {
	tmf.RegisterExtensionReader(extReader{})
	tmf.RegisterExtensionWriter(extWriter{})
}

func (extReader) ReadMeshElement(mesh *tmf.Mesh, elem *helium.Element) error {
	if elem.LocalName() != "beamlattice" {
		return nil
	}
	bl := &BeamLattice{Cap: CapModeSphere}
	if v, ok := attrFloat(elem, "radius"); ok {
		bl.Radius = v
	}
	if v, ok := attrFloat(elem, "minlength"); ok {
		bl.MinLength = v
	}
	if s := attr(elem, "cap"); s != "" {
		bl.Cap = parseCapMode(s)
	}
	if s := attr(elem, "clippingmode"); s != "" {
		bl.ClippingMode = s
	}
	if v, ok := attrUint32(elem, "clippingmesh"); ok {
		bl.ClippingMesh = v
	}
	if v, ok := attrUint32(elem, "representationmesh"); ok {
		bl.RepresentationMesh = v
	}
	for child := range childElems(elem, "") {
		switch child.LocalName() {
		case "beams":
			for b := range childElems(child, "beam") {
				beam := Beam{}
				if v, ok := attrUint32(b, "v1"); ok {
					beam.V1 = v
				}
				if v, ok := attrUint32(b, "v2"); ok {
					beam.V2 = v
				}
				if v, ok := attrFloat(b, "r1"); ok {
					beam.R1 = v
				}
				if v, ok := attrFloat(b, "r2"); ok {
					beam.R2 = v
				}
				if s := attr(b, "cap1"); s != "" {
					beam.Cap1 = parseCapMode(s)
					beam.HasCap1 = true
				}
				if s := attr(b, "cap2"); s != "" {
					beam.Cap2 = parseCapMode(s)
					beam.HasCap2 = true
				}
				bl.Beams = append(bl.Beams, beam)
			}
		case "ballsets":
			for bs := range childElems(child, "ballset") {
				set := BallSet{
					Identifier: attr(bs, "identifier"),
					Name:       attr(bs, "name"),
				}
				for ball := range childElems(bs, "ball") {
					b := Ball{}
					if v, ok := attrUint32(ball, "v"); ok {
						b.V = v
					}
					if v, ok := attrFloat(ball, "r"); ok {
						b.R = v
					}
					set.Balls = append(set.Balls, b)
				}
				bl.BallSets = append(bl.BallSets, set)
			}
		}
	}
	mesh.SetExtension(Namespace, bl)
	return nil
}

func (extWriter) WriteMeshElements(mesh *tmf.Mesh, w *tmf.Writer) error {
	bl := Of(mesh)
	if bl == nil {
		return nil
	}
	if err := w.StartElementNS(Prefix, "beamlattice"); err != nil {
		return err
	}
	if bl.Radius != 0 {
		if err := w.Attr("radius", strconv.FormatFloat(bl.Radius, 'g', -1, 64)); err != nil {
			return err
		}
	}
	if bl.MinLength != 0 {
		if err := w.Attr("minlength", strconv.FormatFloat(bl.MinLength, 'g', -1, 64)); err != nil {
			return err
		}
	}
	if bl.Cap != CapModeSphere {
		if err := w.Attr("cap", bl.Cap.String()); err != nil {
			return err
		}
	}
	if bl.ClippingMode != "" {
		if err := w.Attr("clippingmode", bl.ClippingMode); err != nil {
			return err
		}
	}
	if bl.ClippingMesh != 0 {
		if err := w.Attr("clippingmesh", strconv.FormatUint(uint64(bl.ClippingMesh), 10)); err != nil {
			return err
		}
	}
	if bl.RepresentationMesh != 0 {
		if err := w.Attr("representationmesh", strconv.FormatUint(uint64(bl.RepresentationMesh), 10)); err != nil {
			return err
		}
	}
	if len(bl.Beams) > 0 {
		if err := w.StartElementNS(Prefix, "beams"); err != nil {
			return err
		}
		for _, b := range bl.Beams {
			if err := w.StartElementNS(Prefix, "beam"); err != nil {
				return err
			}
			if err := w.Attr("v1", strconv.FormatUint(uint64(b.V1), 10)); err != nil {
				return err
			}
			if err := w.Attr("v2", strconv.FormatUint(uint64(b.V2), 10)); err != nil {
				return err
			}
			if b.R1 != 0 {
				if err := w.Attr("r1", strconv.FormatFloat(b.R1, 'g', -1, 64)); err != nil {
					return err
				}
			}
			if b.R2 != 0 {
				if err := w.Attr("r2", strconv.FormatFloat(b.R2, 'g', -1, 64)); err != nil {
					return err
				}
			}
			if b.HasCap1 {
				if err := w.Attr("cap1", b.Cap1.String()); err != nil {
					return err
				}
			}
			if b.HasCap2 {
				if err := w.Attr("cap2", b.Cap2.String()); err != nil {
					return err
				}
			}
			if err := w.EndElement(); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	if len(bl.BallSets) > 0 {
		if err := w.StartElementNS(Prefix, "ballsets"); err != nil {
			return err
		}
		for _, set := range bl.BallSets {
			if err := w.StartElementNS(Prefix, "ballset"); err != nil {
				return err
			}
			if set.Identifier != "" {
				if err := w.Attr("identifier", set.Identifier); err != nil {
					return err
				}
			}
			if set.Name != "" {
				if err := w.Attr("name", set.Name); err != nil {
					return err
				}
			}
			for _, b := range set.Balls {
				if err := w.StartElementNS(Prefix, "ball"); err != nil {
					return err
				}
				if err := w.Attr("v", strconv.FormatUint(uint64(b.V), 10)); err != nil {
					return err
				}
				if err := w.Attr("r", strconv.FormatFloat(b.R, 'g', -1, 64)); err != nil {
					return err
				}
				if err := w.EndElement(); err != nil {
					return err
				}
			}
			if err := w.EndElement(); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

// ---- helpers ----

func attr(elem *helium.Element, local string) string {
	a, ok := elem.FindAttribute(helium.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return a.Value()
}

func attrUint32(elem *helium.Element, local string) (uint32, bool) {
	s := attr(elem, local)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

func attrFloat(elem *helium.Element, local string) (float64, bool) {
	s := attr(elem, local)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func childElems(parent *helium.Element, local string) func(yield func(*helium.Element) bool) {
	return func(yield func(*helium.Element) bool) {
		for child := range helium.Children(parent) {
			elem, ok := child.(*helium.Element)
			if !ok {
				continue
			}
			if local != "" && elem.LocalName() != local {
				continue
			}
			if !yield(elem) {
				return
			}
		}
	}
}

// joinUint isn't currently used here but keeps the file parallel to the
// other extension packages and provides a convenient helper for any future
// element types that need to emit space-separated index lists.
//
//nolint:unused
func joinUint(v []uint32) string {
	var b strings.Builder
	for i, n := range v {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.FormatUint(uint64(n), 10))
	}
	return b.String()
}

// fmt is referenced indirectly via error wrapping in the readers above.
var _ = fmt.Sprintf
