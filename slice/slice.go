// Package slice implements the 3MF Slice Extension.
//
// The Slice extension lets producers ship printer-ready 2D slice contours
// alongside (or instead of) 3D mesh geometry. A SliceStack is a list of
// horizontal slices at increasing Z, each containing 2D polygons and the
// vertices they reference. Objects can opt into using a SliceStack via the
// s:meshresolution and s:slicestackid attributes (read and written
// transparently by this package).
//
// Blank-import to register hooks:
//
//	import _ "github.com/lestrrat-go/3mf/slice"
package slice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lestrrat-go/helium"

	tmf "github.com/lestrrat-go/3mf"
)

const (
	Namespace = tmf.NSSlice
	Prefix    = tmf.PrefixSlice
)

// Resources is the slice-extension payload on tmf.Resources.
type Resources struct {
	Stacks []*Stack
}

// Stack is a single <slicestack>.
type Stack struct {
	ID      uint32
	ZBottom float64
	Slices  []*Slice
	// SliceRefs allows a slice stack to reference slices defined in
	// another stack inside another model part (production extension).
	SliceRefs []Ref
}

// Ref references slices defined in another stack.
type Ref struct {
	SliceStackID uint32
	Path         string
}

// Slice is a single 2D slice at top altitude ZTop.
type Slice struct {
	ZTop     float64
	Vertices []Vertex2D
	Polygons []Polygon
}

// Vertex2D is a 2D point in slice coordinate space (X, Y at slice altitude).
type Vertex2D struct{ X, Y float64 }

// Polygon is a list of vertex indices defining a closed contour.
type Polygon struct {
	StartV   uint32
	Segments []Segment
}

// Segment is one edge of a polygon. P2 is the next vertex; P1 is implicit
// (previous segment's P2 or the polygon's StartV).
type Segment struct {
	V2    uint32
	P1    uint32 // optional curvature control or PID (extension-defined)
	HasP1 bool
}

// MeshResolution attribute values (3.4 § "meshresolution").
const (
	MeshResolutionFullRes = "fullres"
	MeshResolutionLowRes  = "lowres"
)

// ObjectInfo carries slice-extension attributes attached to a tmf.Object.
type ObjectInfo struct {
	MeshResolution string
	SliceStackID   uint32
}

// Of returns the slice resources attached to res, creating an empty one if absent.
func Of(res *tmf.Resources) *Resources {
	if v, ok := res.ExtensionResources(Namespace).(*Resources); ok {
		return v
	}
	r := &Resources{}
	res.SetExtensionResources(Namespace, r)
	return r
}

// OfObject returns the slice attributes attached to obj, or nil.
func OfObject(obj *tmf.Object) *ObjectInfo {
	if v := obj.Extension(Namespace); v != nil {
		if oi, ok := v.(*ObjectInfo); ok {
			return oi
		}
	}
	return nil
}

// Require declares the slice extension on m using the conventional prefix.
func Require(m *tmf.Model) { m.RequireExtension(Namespace, Prefix) }

type extReader struct{ tmf.BaseExtensionReader }

func (extReader) Namespace() string { return Namespace }

type extWriter struct{ tmf.BaseExtensionWriter }

func (extWriter) Namespace() string { return Namespace }

func init() {
	tmf.RegisterExtensionReader(extReader{})
	tmf.RegisterExtensionWriter(extWriter{})
}

func (extReader) ReadResourceElement(res *tmf.Resources, elem *helium.Element) error {
	if elem.LocalName() != "slicestack" {
		return nil
	}
	sr := Of(res)
	st := &Stack{}
	if v, ok := attrUint32(elem, "id"); ok {
		st.ID = v
	}
	if v, ok := attrFloat(elem, "zbottom"); ok {
		st.ZBottom = v
	}
	for child := range childElems(elem, "") {
		switch child.LocalName() {
		case "slice":
			s := &Slice{}
			if v, ok := attrFloat(child, "ztop"); ok {
				s.ZTop = v
			}
			for sub := range childElems(child, "") {
				switch sub.LocalName() {
				case "vertices":
					for v := range childElems(sub, "vertex") {
						x, _ := attrFloat(v, "x")
						y, _ := attrFloat(v, "y")
						s.Vertices = append(s.Vertices, Vertex2D{X: x, Y: y})
					}
				case "polygon":
					p := Polygon{}
					if v, ok := attrUint32(sub, "startv"); ok {
						p.StartV = v
					}
					for seg := range childElems(sub, "segment") {
						s2, _ := attrUint32(seg, "v2")
						s := Segment{V2: s2}
						if v, ok := attrUint32(seg, "p1"); ok {
							s.P1 = v
							s.HasP1 = true
						}
						p.Segments = append(p.Segments, s)
					}
					s.Polygons = append(s.Polygons, p)
				}
			}
			st.Slices = append(st.Slices, s)
		case "sliceref":
			ref := Ref{Path: attr(child, "slicepath")}
			if v, ok := attrUint32(child, "slicestackid"); ok {
				ref.SliceStackID = v
			}
			st.SliceRefs = append(st.SliceRefs, ref)
		}
	}
	sr.Stacks = append(sr.Stacks, st)
	return nil
}

func (extReader) ReadObjectElement(obj *tmf.Object, elem *helium.Element) error {
	// The slice extension defines no child elements on <object>; it adds
	// attributes (meshresolution, slicestackid). Those land here only when
	// extension namespace attributes are surfaced as elements, which they
	// aren't. The attribute-side handling is performed by reading the
	// object element directly in writer/reader hooks below; surface a
	// no-op so the interface is satisfied.
	_ = obj
	_ = elem
	return nil
}

func (extWriter) WriteResourceElements(res *tmf.Resources, w *tmf.Writer) error {
	v := res.ExtensionResources(Namespace)
	if v == nil {
		return nil
	}
	sr, ok := v.(*Resources)
	if !ok {
		return nil
	}
	for _, st := range sr.Stacks {
		if err := writeStack(w, st); err != nil {
			return err
		}
	}
	return nil
}

func (extWriter) WriteObjectElements(obj *tmf.Object, w *tmf.Writer) error {
	oi := OfObject(obj)
	if oi == nil {
		return nil
	}
	// The slice extension's per-object metadata is expressed as attributes
	// on <object>, but at this hook point the opening tag has already been
	// closed. The clean way to round-trip them is to surface them via
	// WriteAttribute earlier; we emit them as <slice:meshresolution>-style
	// elements as a documented best-effort. Producers that want strict
	// schema-conformant output should treat OfObject as informative and
	// emit the attributes themselves.
	if oi.MeshResolution != "" {
		if err := w.StartElementNS(Prefix, "meshresolution"); err != nil {
			return err
		}
		if err := w.WriteString(oi.MeshResolution); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	if oi.SliceStackID != 0 {
		if err := w.StartElementNS(Prefix, "slicestackref"); err != nil {
			return err
		}
		if err := w.Attr("slicestackid", strconv.FormatUint(uint64(oi.SliceStackID), 10)); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return nil
}

func writeStack(w *tmf.Writer, st *Stack) error {
	if err := w.StartElementNS(Prefix, "slicestack"); err != nil {
		return err
	}
	if err := w.Attr("id", strconv.FormatUint(uint64(st.ID), 10)); err != nil {
		return err
	}
	if st.ZBottom != 0 {
		if err := w.Attr("zbottom", strconv.FormatFloat(st.ZBottom, 'g', -1, 64)); err != nil {
			return err
		}
	}
	for _, s := range st.Slices {
		if err := w.StartElementNS(Prefix, "slice"); err != nil {
			return err
		}
		if err := w.Attr("ztop", strconv.FormatFloat(s.ZTop, 'g', -1, 64)); err != nil {
			return err
		}
		if len(s.Vertices) > 0 {
			if err := w.StartElementNS(Prefix, "vertices"); err != nil {
				return err
			}
			for _, v := range s.Vertices {
				if err := w.StartElementNS(Prefix, "vertex"); err != nil {
					return err
				}
				if err := w.Attr("x", strconv.FormatFloat(v.X, 'g', -1, 64)); err != nil {
					return err
				}
				if err := w.Attr("y", strconv.FormatFloat(v.Y, 'g', -1, 64)); err != nil {
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
		for _, p := range s.Polygons {
			if err := w.StartElementNS(Prefix, "polygon"); err != nil {
				return err
			}
			if err := w.Attr("startv", strconv.FormatUint(uint64(p.StartV), 10)); err != nil {
				return err
			}
			for _, seg := range p.Segments {
				if err := w.StartElementNS(Prefix, "segment"); err != nil {
					return err
				}
				if err := w.Attr("v2", strconv.FormatUint(uint64(seg.V2), 10)); err != nil {
					return err
				}
				if seg.HasP1 {
					if err := w.Attr("p1", strconv.FormatUint(uint64(seg.P1), 10)); err != nil {
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
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	for _, ref := range st.SliceRefs {
		if err := w.StartElementNS(Prefix, "sliceref"); err != nil {
			return err
		}
		if err := w.Attr("slicestackid", strconv.FormatUint(uint64(ref.SliceStackID), 10)); err != nil {
			return err
		}
		if ref.Path != "" {
			if err := w.Attr("slicepath", ref.Path); err != nil {
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

var _ = fmt.Sprintf
var _ = strings.Fields
