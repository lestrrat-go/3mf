package tmf

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lestrrat-go/helium"

	"github.com/lestrrat-go/3mf/internal/xmlutil"
)

// ReadModel parses a 3dmodel.model XML payload into a *Model.
func ReadModel(ctx context.Context, data []byte) (*Model, error) {
	doc, err := helium.NewParser().Parse(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("tmf: parse model XML: %w", err)
	}
	root := doc.DocumentElement()
	if root == nil || root.LocalName() != "model" {
		return nil, fmt.Errorf("%w: root element must be <model>", ErrMalformedModel)
	}
	return readModel(root)
}

func readModel(root *helium.Element) (*Model, error) {
	m := NewModel()
	// Unit
	if s := xmlutil.Attr(root, "unit"); s != "" {
		u, err := ParseUnit(s)
		if err != nil {
			return nil, err
		}
		m.unit = u
	} else {
		m.unit = UnitMillimeter
	}
	m.language = xmlutil.AttrNS(root, "lang", NSXML)
	m.thumbnail = xmlutil.Attr(root, "thumbnail")

	// requiredextensions + namespace declarations: capture prefix -> URI
	// for every namespace declared on <model>, and resolve the
	// requiredextensions attribute (a list of prefixes) into URIs.
	prefixToURI := map[string]string{}
	for _, ns := range root.Namespaces() {
		if ns.URI() != NSCore && ns.URI() != "" {
			prefixToURI[ns.Prefix()] = ns.URI()
		}
	}
	if req := xmlutil.Attr(root, "requiredextensions"); req != "" {
		for _, prefix := range fieldsCommaSpace(req) {
			if uri, ok := prefixToURI[prefix]; ok {
				m.RequireExtension(uri, prefix)
			}
		}
	}

	// metadata, resources, build (in that order per spec)
	for elem := range xmlutil.ChildElements(root, "") {
		switch elem.LocalName() {
		case "metadata":
			m.metadata = append(m.metadata, readMetadata(elem))
		case "resources":
			if err := readResources(elem, m.resources); err != nil {
				return nil, err
			}
		case "build":
			if err := readBuild(elem, m.build); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

func readMetadata(elem *helium.Element) *Metadata {
	md := &Metadata{
		Name: xmlutil.Attr(elem, "name"),
		Type: xmlutil.Attr(elem, "type"),
	}
	if p := xmlutil.Attr(elem, "preserve"); p == "true" || p == "1" {
		md.Preserve = true
	}
	if t := xmlutil.TextContent(elem); t != "" {
		md.Value = t
	}
	return md
}

func readResources(elem *helium.Element, res *Resources) error {
	for child := range xmlutil.ChildElements(elem, "") {
		ns := ""
		if n := child.Namespace(); n != nil {
			ns = n.URI()
		}
		if ns != "" && ns != NSCore {
			if r := LookupExtensionReader(ns); r != nil {
				if err := r.ReadResourceElement(res, child); err != nil {
					return err
				}
			}
			continue
		}
		switch child.LocalName() {
		case "object":
			obj, err := readObject(child)
			if err != nil {
				return err
			}
			res.AppendObject(obj)
		case "basematerials":
			b, err := readBaseMaterials(child)
			if err != nil {
				return err
			}
			res.AppendBaseMaterials(b)
		}
	}
	return nil
}

func readBaseMaterials(elem *helium.Element) (*BaseMaterials, error) {
	id, _ := xmlutil.AttrUint32(elem, "id")
	bm := &BaseMaterials{id: id}
	for child := range xmlutil.ChildElements(elem, "base") {
		mat := BaseMaterial{Name: xmlutil.Attr(child, "name")}
		c, err := ParseColor(xmlutil.Attr(child, "displaycolor"))
		if err != nil {
			return nil, err
		}
		mat.DisplayColor = c
		bm.materials = append(bm.materials, mat)
	}
	return bm, nil
}

func readObject(elem *helium.Element) (*Object, error) {
	obj := &Object{}
	if v, ok := xmlutil.AttrUint32(elem, "id"); ok {
		obj.id = v
	}
	obj.name = xmlutil.Attr(elem, "name")
	t, err := ParseObjectType(xmlutil.Attr(elem, "type"))
	if err != nil {
		return nil, err
	}
	obj.objType = t
	obj.uuid = xmlutil.AttrNS(elem, "UUID", NSProduction)
	obj.partNumber = xmlutil.Attr(elem, "partnumber")
	obj.thumbnail = xmlutil.Attr(elem, "thumbnail")
	if v, ok := xmlutil.AttrUint32(elem, "pid"); ok {
		obj.pid = v
		obj.hasPID = true
	}
	if v, ok := xmlutil.AttrUint32(elem, "pindex"); ok {
		obj.pIndex = v
		obj.hasPID = true
	}
	for child := range xmlutil.ChildElements(elem, "") {
		ns := ""
		if n := child.Namespace(); n != nil {
			ns = n.URI()
		}
		if ns != "" && ns != NSCore {
			if r := LookupExtensionReader(ns); r != nil {
				if err := r.ReadObjectElement(obj, child); err != nil {
					return nil, err
				}
			}
			continue
		}
		switch child.LocalName() {
		case "mesh":
			mesh, err := readMesh(child)
			if err != nil {
				return nil, err
			}
			obj.mesh = mesh
		case "components":
			for c := range xmlutil.ChildElements(child, "component") {
				comp, err := readComponent(c)
				if err != nil {
					return nil, err
				}
				obj.components = append(obj.components, comp)
			}
		}
	}
	return obj, nil
}

func readMesh(elem *helium.Element) (*Mesh, error) {
	mesh := &Mesh{}
	for child := range xmlutil.ChildElements(elem, "") {
		ns := ""
		if n := child.Namespace(); n != nil {
			ns = n.URI()
		}
		if ns != "" && ns != NSCore {
			if r := LookupExtensionReader(ns); r != nil {
				if err := r.ReadMeshElement(mesh, child); err != nil {
					return nil, err
				}
			}
			continue
		}
		switch child.LocalName() {
		case "vertices":
			for v := range xmlutil.ChildElements(child, "vertex") {
				x, _ := xmlutil.AttrFloat64(v, "x")
				y, _ := xmlutil.AttrFloat64(v, "y")
				z, _ := xmlutil.AttrFloat64(v, "z")
				mesh.vertices = append(mesh.vertices, Vertex{X: x, Y: y, Z: z})
			}
		case "triangles":
			for t := range xmlutil.ChildElements(child, "triangle") {
				tri, err := readTriangle(t)
				if err != nil {
					return nil, err
				}
				mesh.triangles = append(mesh.triangles, tri)
			}
		}
	}
	return mesh, nil
}

func readTriangle(elem *helium.Element) (Triangle, error) {
	v1, _ := xmlutil.AttrUint32(elem, "v1")
	v2, _ := xmlutil.AttrUint32(elem, "v2")
	v3, _ := xmlutil.AttrUint32(elem, "v3")
	t := Triangle{V1: v1, V2: v2, V3: v3}
	if pid, ok := xmlutil.AttrUint32(elem, "pid"); ok {
		t.PID = pid
		t.HasPID = true
	}
	p1, hasP1 := xmlutil.AttrUint32(elem, "p1")
	p2, hasP2 := xmlutil.AttrUint32(elem, "p2")
	p3, hasP3 := xmlutil.AttrUint32(elem, "p3")
	if hasP1 || hasP2 || hasP3 {
		t.P1 = p1
		if hasP2 {
			t.P2 = p2
		} else {
			t.P2 = p1
		}
		if hasP3 {
			t.P3 = p3
		} else {
			t.P3 = p1
		}
		t.HasPIndices = true
	}
	return t, nil
}

func readComponent(elem *helium.Element) (*Component, error) {
	c := &Component{Transform: IdentityMatrix()}
	id, _ := xmlutil.AttrUint32(elem, "objectid")
	c.ObjectID = id
	c.UUID = xmlutil.AttrNS(elem, "UUID", NSProduction)
	c.Path = xmlutil.AttrNS(elem, "path", NSProduction)
	if s := xmlutil.Attr(elem, "transform"); s != "" {
		m, err := ParseMatrix(s)
		if err != nil {
			return nil, err
		}
		c.Transform = m
	}
	return c, nil
}

func readBuild(elem *helium.Element, build *Build) error {
	build.UUID = xmlutil.AttrNS(elem, "UUID", NSProduction)
	for child := range xmlutil.ChildElements(elem, "") {
		ns := ""
		if n := child.Namespace(); n != nil {
			ns = n.URI()
		}
		if ns != "" && ns != NSCore {
			if r := LookupExtensionReader(ns); r != nil {
				if err := r.ReadBuildElement(build, child); err != nil {
					return err
				}
			}
			continue
		}
		if child.LocalName() != "item" {
			continue
		}
		item, err := readBuildItem(child)
		if err != nil {
			return err
		}
		build.Items = append(build.Items, item)
	}
	return nil
}

func readBuildItem(elem *helium.Element) (*BuildItem, error) {
	id, _ := xmlutil.AttrUint32(elem, "objectid")
	item := &BuildItem{
		ObjectID:   id,
		Transform:  IdentityMatrix(),
		PartNumber: xmlutil.Attr(elem, "partnumber"),
		Path:       xmlutil.AttrNS(elem, "path", NSProduction),
		UUID:       xmlutil.AttrNS(elem, "UUID", NSProduction),
	}
	if s := xmlutil.Attr(elem, "transform"); s != "" {
		m, err := ParseMatrix(s)
		if err != nil {
			return nil, err
		}
		item.Transform = m
	}
	for md := range xmlutil.ChildElements(elem, "metadatagroup") {
		for mElem := range xmlutil.ChildElements(md, "metadata") {
			item.Metadata = append(item.Metadata, readMetadata(mElem))
		}
	}
	return item, nil
}

// fieldsCommaSpace splits s on ASCII whitespace or commas; empty fields
// are discarded.
func fieldsCommaSpace(s string) []string {
	var out []string
	start := -1
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			if start >= 0 {
				out = append(out, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

// formatUint renders u in decimal without heap allocation in hot paths.
//
//nolint:unused
func formatUint(u uint32) string { return strconv.FormatUint(uint64(u), 10) }
