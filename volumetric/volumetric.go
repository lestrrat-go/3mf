// Package volumetric implements the 3MF Volumetric Extension.
//
// The Volumetric extension adds <function> and <volumetric> resources that
// describe implicit and field-based geometry (signed distance functions,
// gradient fields, etc.) suitable for slicers that support voxel/volumetric
// printing. The extension is more recent than the others and its schema is
// still evolving; this package implements the v1 (2022/01) schema with a
// generic-payload fallback for elements it does not yet recognize.
//
// Blank-import to register hooks:
//
//	import _ "github.com/lestrrat-go/3mf/volumetric"
package volumetric

import (
	"strconv"

	"github.com/lestrrat-go/helium"

	tmf "github.com/lestrrat-go/3mf"
)

const (
	Namespace = tmf.NSVolumetric
	Prefix    = tmf.PrefixVolumetric
)

// Resources is the volumetric payload attached to tmf.Resources.
type Resources struct {
	Functions   []*Function
	Volumetrics []*Volumetric
}

// Function is a v:function resource: a named, evaluable scalar/vector field.
type Function struct {
	ID         uint32
	DisplayName string
	// Body holds the parsed children of the <function> element. Because the
	// volumetric schema includes a large taxonomy of node types
	// (constants, references, math ops, lookups, etc.), this package
	// preserves them as a tree of generic Nodes so that round-trip is
	// faithful even when consumers don't understand every node kind.
	Body []*Node
}

// Volumetric is a v:volumetric resource: an explicit volumetric material
// mapping that references a Function by id and selects a channel.
type Volumetric struct {
	ID         uint32
	FunctionID uint32
	Channel    string
	Inputs     map[string]string
}

// Node is a generic XML-like node used to preserve unknown volumetric
// elements. It records the local name and attributes, plus any nested
// children.
type Node struct {
	Name       string
	Attributes map[string]string
	Children   []*Node
	Text       string
}

// Of returns the volumetric resources attached to res, creating an empty one
// when absent.
func Of(res *tmf.Resources) *Resources {
	if v := res.ExtensionResources(Namespace); v != nil {
		return v.(*Resources)
	}
	r := &Resources{}
	res.SetExtensionResources(Namespace, r)
	return r
}

// Require declares the volumetric extension on m.
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
	vr := Of(res)
	switch elem.LocalName() {
	case "function":
		fn := &Function{
			DisplayName: attr(elem, "displayname"),
		}
		if v, ok := attrUint32(elem, "id"); ok {
			fn.ID = v
		}
		for c := range childElems(elem, "") {
			fn.Body = append(fn.Body, readNode(c))
		}
		vr.Functions = append(vr.Functions, fn)
	case "volumetric":
		v := &Volumetric{Inputs: map[string]string{}}
		if id, ok := attrUint32(elem, "id"); ok {
			v.ID = id
		}
		if id, ok := attrUint32(elem, "functionid"); ok {
			v.FunctionID = id
		}
		v.Channel = attr(elem, "channel")
		for _, a := range elem.Attributes() {
			if a == nil {
				continue
			}
			name := a.Name()
			switch name {
			case "id", "functionid", "channel":
				continue
			}
			v.Inputs[name] = a.Value()
		}
		vr.Volumetrics = append(vr.Volumetrics, v)
	}
	return nil
}

func readNode(elem *helium.Element) *Node {
	n := &Node{
		Name:       elem.LocalName(),
		Attributes: map[string]string{},
	}
	for _, a := range elem.Attributes() {
		if a == nil {
			continue
		}
		n.Attributes[a.Name()] = a.Value()
	}
	for c := range helium.Children(elem) {
		switch v := c.(type) {
		case *helium.Element:
			n.Children = append(n.Children, readNode(v))
		case *helium.Text:
			n.Text += string(v.Content())
		}
	}
	return n
}

func (extWriter) WriteResourceElements(res *tmf.Resources, w *tmf.Writer) error {
	v := res.ExtensionResources(Namespace)
	if v == nil {
		return nil
	}
	vr, ok := v.(*Resources)
	if !ok {
		return nil
	}
	for _, fn := range vr.Functions {
		if err := w.StartElementNS(Prefix, "function"); err != nil {
			return err
		}
		if err := w.Attr("id", strconv.FormatUint(uint64(fn.ID), 10)); err != nil {
			return err
		}
		if fn.DisplayName != "" {
			if err := w.Attr("displayname", fn.DisplayName); err != nil {
				return err
			}
		}
		for _, n := range fn.Body {
			if err := writeNode(w, n); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	for _, vol := range vr.Volumetrics {
		if err := w.StartElementNS(Prefix, "volumetric"); err != nil {
			return err
		}
		if err := w.Attr("id", strconv.FormatUint(uint64(vol.ID), 10)); err != nil {
			return err
		}
		if vol.FunctionID != 0 {
			if err := w.Attr("functionid", strconv.FormatUint(uint64(vol.FunctionID), 10)); err != nil {
				return err
			}
		}
		if vol.Channel != "" {
			if err := w.Attr("channel", vol.Channel); err != nil {
				return err
			}
		}
		for k, vv := range vol.Inputs {
			if err := w.Attr(k, vv); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return nil
}

func writeNode(w *tmf.Writer, n *Node) error {
	if err := w.StartElementNS(Prefix, n.Name); err != nil {
		return err
	}
	for k, v := range n.Attributes {
		if err := w.Attr(k, v); err != nil {
			return err
		}
	}
	if n.Text != "" {
		if err := w.WriteString(n.Text); err != nil {
			return err
		}
	}
	for _, c := range n.Children {
		if err := writeNode(w, c); err != nil {
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
