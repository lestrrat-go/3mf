package materials

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lestrrat-go/helium"

	tmf "github.com/lestrrat-go/3mf"
)

type extReader struct{ tmf.BaseExtensionReader }

func (extReader) Namespace() string { return Namespace }

type extWriter struct{ tmf.BaseExtensionWriter }

func (extWriter) Namespace() string { return Namespace }

func init() {
	tmf.RegisterExtensionReader(extReader{})
	tmf.RegisterExtensionWriter(extWriter{})
}

// Require declares the materials extension on m using the conventional prefix.
func Require(m *tmf.Model) { m.RequireExtension(Namespace, Prefix) }

// Of returns the materials Resources attached to res, creating an empty
// one if none exists yet.
func Of(res *tmf.Resources) *Resources {
	if v := res.ExtensionResources(Namespace); v != nil {
		return v.(*Resources)
	}
	r := &Resources{}
	res.SetExtensionResources(Namespace, r)
	return r
}

func (extReader) ReadResourceElement(res *tmf.Resources, elem *helium.Element) error {
	mr := Of(res)
	switch elem.LocalName() {
	case "colorgroup":
		cg, err := readColorGroup(elem)
		if err != nil {
			return err
		}
		mr.ColorGroups = append(mr.ColorGroups, cg)
	case "texture2d":
		t, err := readTexture2D(elem)
		if err != nil {
			return err
		}
		mr.Texture2Ds = append(mr.Texture2Ds, t)
	case "texture2dgroup":
		g, err := readTexture2DGroup(elem)
		if err != nil {
			return err
		}
		mr.Texture2DGroups = append(mr.Texture2DGroups, g)
	case "compositematerials":
		c, err := readCompositeMaterials(elem)
		if err != nil {
			return err
		}
		mr.CompositeMaterials = append(mr.CompositeMaterials, c)
	case "multiproperties":
		mp, err := readMultiProperties(elem)
		if err != nil {
			return err
		}
		mr.MultiProperties = append(mr.MultiProperties, mp)
	}
	return nil
}

func (extWriter) WriteResourceElements(res *tmf.Resources, w *tmf.Writer) error {
	v := res.ExtensionResources(Namespace)
	if v == nil {
		return nil
	}
	mr, ok := v.(*Resources)
	if !ok {
		return nil
	}
	for _, t := range mr.Texture2Ds {
		if err := writeTexture2D(w, t); err != nil {
			return err
		}
	}
	for _, cg := range mr.ColorGroups {
		if err := writeColorGroup(w, cg); err != nil {
			return err
		}
	}
	for _, g := range mr.Texture2DGroups {
		if err := writeTexture2DGroup(w, g); err != nil {
			return err
		}
	}
	for _, c := range mr.CompositeMaterials {
		if err := writeCompositeMaterials(w, c); err != nil {
			return err
		}
	}
	for _, mp := range mr.MultiProperties {
		if err := writeMultiProperties(w, mp); err != nil {
			return err
		}
	}
	return nil
}

// ---- readers ----

func readColorGroup(elem *helium.Element) (*ColorGroup, error) {
	id, _ := attrUint32(elem, "id")
	cg := &ColorGroup{ID: id}
	for child := range childElems(elem, "color") {
		c, err := tmf.ParseColor(attr(child, "color"))
		if err != nil {
			return nil, err
		}
		cg.Colors = append(cg.Colors, c)
	}
	return cg, nil
}

func readTexture2D(elem *helium.Element) (*Texture2D, error) {
	id, _ := attrUint32(elem, "id")
	t := &Texture2D{
		ID:          id,
		Path:        attr(elem, "path"),
		ContentType: attr(elem, "contenttype"),
		TileStyleU:  attr(elem, "tilestyleu"),
		TileStyleV:  attr(elem, "tilestylev"),
		Filter:      attr(elem, "filter"),
	}
	if s := attr(elem, "box"); s != "" {
		fields := strings.Fields(s)
		if len(fields) == 4 {
			u0, _ := strconv.ParseFloat(fields[0], 64)
			v0, _ := strconv.ParseFloat(fields[1], 64)
			u1, _ := strconv.ParseFloat(fields[2], 64)
			v1, _ := strconv.ParseFloat(fields[3], 64)
			t.BoxMin = &Vec2{U: u0, V: v0}
			t.BoxMax = &Vec2{U: u1, V: v1}
		}
	}
	return t, nil
}

func readTexture2DGroup(elem *helium.Element) (*Texture2DGroup, error) {
	id, _ := attrUint32(elem, "id")
	tex, _ := attrUint32(elem, "texid")
	g := &Texture2DGroup{ID: id, TextureID: tex}
	for child := range childElems(elem, "tex2coord") {
		u, _ := attrFloat(child, "u")
		v, _ := attrFloat(child, "v")
		g.Coords = append(g.Coords, TextureCoord{U: u, V: v})
	}
	return g, nil
}

func readCompositeMaterials(elem *helium.Element) (*CompositeMaterials, error) {
	id, _ := attrUint32(elem, "id")
	mat, _ := attrUint32(elem, "matid")
	c := &CompositeMaterials{ID: id, MatID: mat}
	if s := attr(elem, "matindices"); s != "" {
		for _, f := range strings.Fields(s) {
			n, err := strconv.ParseUint(f, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("materials: matindices: %w", err)
			}
			c.MatIndices = append(c.MatIndices, uint32(n))
		}
	}
	for child := range childElems(elem, "composite") {
		values := strings.Fields(attr(child, "values"))
		row := Composite{Values: make([]float64, len(values))}
		for i, f := range values {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return nil, fmt.Errorf("materials: composite values: %w", err)
			}
			row.Values[i] = v
		}
		c.Composites = append(c.Composites, row)
	}
	return c, nil
}

func readMultiProperties(elem *helium.Element) (*MultiProperties, error) {
	id, _ := attrUint32(elem, "id")
	mp := &MultiProperties{ID: id}
	if s := attr(elem, "pids"); s != "" {
		for _, f := range strings.Fields(s) {
			n, err := strconv.ParseUint(f, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("materials: pids: %w", err)
			}
			mp.PIDs = append(mp.PIDs, uint32(n))
		}
	}
	if s := attr(elem, "blendmethods"); s != "" {
		mp.BlendMethods = strings.Fields(s)
	}
	for child := range childElems(elem, "multi") {
		entry := MultiEntry{}
		for _, f := range strings.Fields(attr(child, "pindices")) {
			n, err := strconv.ParseUint(f, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("materials: pindices: %w", err)
			}
			entry.PIndices = append(entry.PIndices, uint32(n))
		}
		mp.Multis = append(mp.Multis, entry)
	}
	return mp, nil
}

// ---- writers ----

func writeColorGroup(w *tmf.Writer, cg *ColorGroup) error {
	if err := w.StartElementNS(Prefix, "colorgroup"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(cg.ID)); err != nil {
		return err
	}
	for _, c := range cg.Colors {
		if err := w.StartElementNS(Prefix, "color"); err != nil {
			return err
		}
		if err := w.Attr("color", c.String()); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeTexture2D(w *tmf.Writer, t *Texture2D) error {
	if err := w.StartElementNS(Prefix, "texture2d"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(t.ID)); err != nil {
		return err
	}
	if err := w.Attr("path", t.Path); err != nil {
		return err
	}
	if err := w.Attr("contenttype", t.ContentType); err != nil {
		return err
	}
	if t.TileStyleU != "" {
		if err := w.Attr("tilestyleu", t.TileStyleU); err != nil {
			return err
		}
	}
	if t.TileStyleV != "" {
		if err := w.Attr("tilestylev", t.TileStyleV); err != nil {
			return err
		}
	}
	if t.Filter != "" {
		if err := w.Attr("filter", t.Filter); err != nil {
			return err
		}
	}
	if t.BoxMin != nil && t.BoxMax != nil {
		box := fmt.Sprintf("%g %g %g %g", t.BoxMin.U, t.BoxMin.V, t.BoxMax.U, t.BoxMax.V)
		if err := w.Attr("box", box); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeTexture2DGroup(w *tmf.Writer, g *Texture2DGroup) error {
	if err := w.StartElementNS(Prefix, "texture2dgroup"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(g.ID)); err != nil {
		return err
	}
	if err := w.Attr("texid", formatU32(g.TextureID)); err != nil {
		return err
	}
	for _, c := range g.Coords {
		if err := w.StartElementNS(Prefix, "tex2coord"); err != nil {
			return err
		}
		if err := w.Attr("u", strconv.FormatFloat(c.U, 'g', -1, 64)); err != nil {
			return err
		}
		if err := w.Attr("v", strconv.FormatFloat(c.V, 'g', -1, 64)); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeCompositeMaterials(w *tmf.Writer, c *CompositeMaterials) error {
	if err := w.StartElementNS(Prefix, "compositematerials"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(c.ID)); err != nil {
		return err
	}
	if err := w.Attr("matid", formatU32(c.MatID)); err != nil {
		return err
	}
	if len(c.MatIndices) > 0 {
		if err := w.Attr("matindices", joinUint(c.MatIndices)); err != nil {
			return err
		}
	}
	for _, comp := range c.Composites {
		if err := w.StartElementNS(Prefix, "composite"); err != nil {
			return err
		}
		if err := w.Attr("values", joinFloat(comp.Values)); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeMultiProperties(w *tmf.Writer, mp *MultiProperties) error {
	if err := w.StartElementNS(Prefix, "multiproperties"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(mp.ID)); err != nil {
		return err
	}
	if len(mp.PIDs) > 0 {
		if err := w.Attr("pids", joinUint(mp.PIDs)); err != nil {
			return err
		}
	}
	if len(mp.BlendMethods) > 0 {
		if err := w.Attr("blendmethods", strings.Join(mp.BlendMethods, " ")); err != nil {
			return err
		}
	}
	for _, entry := range mp.Multis {
		if err := w.StartElementNS(Prefix, "multi"); err != nil {
			return err
		}
		if err := w.Attr("pindices", joinUint(entry.PIndices)); err != nil {
			return err
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

func formatU32(u uint32) string { return strconv.FormatUint(uint64(u), 10) }

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

func joinFloat(v []float64) string {
	var b strings.Builder
	for i, n := range v {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.FormatFloat(n, 'g', -1, 64))
	}
	return b.String()
}
