package tmf

import (
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"
)

// WriteModel serializes m as a 3dmodel.model XML document to w.
func WriteModel(w io.Writer, m *Model) error {
	xw := NewWriter(w)
	return writeModel(xw, m)
}

// MarshalModel serializes m as a 3dmodel.model XML document.
func MarshalModel(m *Model) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteModel(&buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeModel(w *Writer, m *Model) error {
	sw := w.Raw()
	if err := sw.StartDocument("1.0", "UTF-8", "yes"); err != nil {
		return err
	}
	if err := sw.StartElement("model"); err != nil {
		return err
	}

	// xmlns + extension namespaces. We pre-declare on root so that prefix-
	// qualified attributes / elements written deeper in the document do not
	// re-declare them.
	if err := w.Attr("xmlns", NSCore); err != nil {
		return err
	}
	w.DeclarePrefix("", NSCore)
	// Stable order for required extensions: sort by prefix.
	extPrefixes := make([]string, 0, len(m.requiredExtensions))
	uriByPrefix := map[string]string{}
	for uri, prefix := range m.requiredExtensions {
		if prefix == "" {
			continue
		}
		extPrefixes = append(extPrefixes, prefix)
		uriByPrefix[prefix] = uri
	}
	sort.Strings(extPrefixes)
	for _, prefix := range extPrefixes {
		uri := uriByPrefix[prefix]
		if err := w.Attr("xmlns:"+prefix, uri); err != nil {
			return err
		}
		w.DeclarePrefix(prefix, uri)
	}

	// unit, xml:lang, requiredextensions, thumbnail
	if m.unit != UnitUnknown {
		if err := w.Attr("unit", m.unit.String()); err != nil {
			return err
		}
	}
	if m.language != "" {
		if err := w.Attr("xml:lang", m.language); err != nil {
			return err
		}
	}
	if len(extPrefixes) > 0 {
		if err := w.Attr("requiredextensions", strings.Join(extPrefixes, " ")); err != nil {
			return err
		}
	}
	if m.thumbnail != "" {
		if err := w.Attr("thumbnail", m.thumbnail); err != nil {
			return err
		}
	}

	// metadata
	for _, md := range m.metadata {
		if err := writeMetadata(w, md); err != nil {
			return err
		}
	}

	// resources
	if err := w.StartElement("resources"); err != nil {
		return err
	}
	for _, b := range m.resources.baseMaterials {
		if err := writeBaseMaterials(w, b); err != nil {
			return err
		}
	}
	for _, o := range m.resources.objects {
		if err := writeObject(w, o); err != nil {
			return err
		}
	}
	// Extension-supplied resource elements
	for _, prefix := range extPrefixes {
		uri := uriByPrefix[prefix]
		if ew := LookupExtensionWriter(uri); ew != nil {
			if err := ew.WriteResourceElements(m.resources, w); err != nil {
				return err
			}
		}
	}
	if err := w.EndElement(); err != nil {
		return err
	}

	// build
	if err := w.StartElement("build"); err != nil {
		return err
	}
	if m.build.UUID != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "UUID", m.build.UUID); err != nil {
				return err
			}
		}
	}
	for _, item := range m.build.Items {
		if err := writeBuildItem(w, item); err != nil {
			return err
		}
	}
	for _, prefix := range extPrefixes {
		uri := uriByPrefix[prefix]
		if ew := LookupExtensionWriter(uri); ew != nil {
			if err := ew.WriteBuildElements(m.build, w); err != nil {
				return err
			}
		}
	}
	if err := w.EndElement(); err != nil {
		return err
	}

	if err := w.EndElement(); err != nil { // </model>
		return err
	}
	return sw.EndDocument()
}

func writeMetadata(w *Writer, md *Metadata) error {
	if err := w.StartElement("metadata"); err != nil {
		return err
	}
	if err := w.Attr("name", md.Name); err != nil {
		return err
	}
	if md.Type != "" {
		if err := w.Attr("type", md.Type); err != nil {
			return err
		}
	}
	if md.Preserve {
		if err := w.Attr("preserve", "true"); err != nil {
			return err
		}
	}
	if md.Value != "" {
		if err := w.WriteString(md.Value); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeBaseMaterials(w *Writer, b *BaseMaterials) error {
	if err := w.StartElement("basematerials"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(b.id)); err != nil {
		return err
	}
	for _, mat := range b.materials {
		if err := w.StartElement("base"); err != nil {
			return err
		}
		if err := w.Attr("name", mat.Name); err != nil {
			return err
		}
		if mat.DisplayColor.Set {
			if err := w.Attr("displaycolor", mat.DisplayColor.String()); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

func writeObject(w *Writer, o *Object) error {
	if err := w.StartElement("object"); err != nil {
		return err
	}
	if err := w.Attr("id", formatU32(o.id)); err != nil {
		return err
	}
	if o.objType != ObjectTypeModel {
		if err := w.Attr("type", o.objType.String()); err != nil {
			return err
		}
	}
	if o.name != "" {
		if err := w.Attr("name", o.name); err != nil {
			return err
		}
	}
	if o.partNumber != "" {
		if err := w.Attr("partnumber", o.partNumber); err != nil {
			return err
		}
	}
	if o.thumbnail != "" {
		if err := w.Attr("thumbnail", o.thumbnail); err != nil {
			return err
		}
	}
	if o.hasPID {
		if err := w.Attr("pid", formatU32(o.pid)); err != nil {
			return err
		}
		if err := w.Attr("pindex", formatU32(o.pIndex)); err != nil {
			return err
		}
	}
	if o.uuid != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "UUID", o.uuid); err != nil {
				return err
			}
		}
	}

	if o.mesh != nil {
		if err := writeMesh(w, o.mesh); err != nil {
			return err
		}
	}
	if len(o.components) > 0 {
		if err := w.StartElement("components"); err != nil {
			return err
		}
		for _, c := range o.components {
			if err := writeComponent(w, c); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}

	// Extension element hooks (per-object).
	for _, ns := range sortedExtensionNamespaces(w) {
		if ew := LookupExtensionWriter(ns); ew != nil {
			if err := ew.WriteObjectElements(o, w); err != nil {
				return err
			}
		}
	}
	return w.EndElement()
}

func writeMesh(w *Writer, m *Mesh) error {
	if err := w.StartElement("mesh"); err != nil {
		return err
	}
	if err := w.StartElement("vertices"); err != nil {
		return err
	}
	for _, v := range m.vertices {
		if err := w.StartElement("vertex"); err != nil {
			return err
		}
		if err := w.Attr("x", formatFloat(v.X)); err != nil {
			return err
		}
		if err := w.Attr("y", formatFloat(v.Y)); err != nil {
			return err
		}
		if err := w.Attr("z", formatFloat(v.Z)); err != nil {
			return err
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	if err := w.EndElement(); err != nil {
		return err
	}
	if err := w.StartElement("triangles"); err != nil {
		return err
	}
	for _, t := range m.triangles {
		if err := w.StartElement("triangle"); err != nil {
			return err
		}
		if err := w.Attr("v1", formatU32(t.V1)); err != nil {
			return err
		}
		if err := w.Attr("v2", formatU32(t.V2)); err != nil {
			return err
		}
		if err := w.Attr("v3", formatU32(t.V3)); err != nil {
			return err
		}
		if t.HasPID {
			if err := w.Attr("pid", formatU32(t.PID)); err != nil {
				return err
			}
		}
		if t.HasPIndices {
			if err := w.Attr("p1", formatU32(t.P1)); err != nil {
				return err
			}
			if err := w.Attr("p2", formatU32(t.P2)); err != nil {
				return err
			}
			if err := w.Attr("p3", formatU32(t.P3)); err != nil {
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
	// Mesh-level extension hooks
	for _, ns := range sortedExtensionNamespaces(w) {
		if ew := LookupExtensionWriter(ns); ew != nil {
			if err := ew.WriteMeshElements(m, w); err != nil {
				return err
			}
		}
	}
	return w.EndElement()
}

func writeComponent(w *Writer, c *Component) error {
	if err := w.StartElement("component"); err != nil {
		return err
	}
	if err := w.Attr("objectid", formatU32(c.ObjectID)); err != nil {
		return err
	}
	if !c.Transform.IsIdentity() {
		if err := w.Attr("transform", FormatMatrix(c.Transform)); err != nil {
			return err
		}
	}
	if c.Path != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "path", c.Path); err != nil {
				return err
			}
		}
	}
	if c.UUID != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "UUID", c.UUID); err != nil {
				return err
			}
		}
	}
	return w.EndElement()
}

func writeBuildItem(w *Writer, b *BuildItem) error {
	if err := w.StartElement("item"); err != nil {
		return err
	}
	if err := w.Attr("objectid", formatU32(b.ObjectID)); err != nil {
		return err
	}
	if !b.Transform.IsIdentity() {
		if err := w.Attr("transform", FormatMatrix(b.Transform)); err != nil {
			return err
		}
	}
	if b.PartNumber != "" {
		if err := w.Attr("partnumber", b.PartNumber); err != nil {
			return err
		}
	}
	if b.Path != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "path", b.Path); err != nil {
				return err
			}
		}
	}
	if b.UUID != "" {
		if prefix := w.PrefixFor(NSProduction); prefix != "" {
			if err := w.AttrNS(prefix, "UUID", b.UUID); err != nil {
				return err
			}
		}
	}
	if len(b.Metadata) > 0 {
		if err := w.StartElement("metadatagroup"); err != nil {
			return err
		}
		for _, md := range b.Metadata {
			if err := writeMetadata(w, md); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return w.EndElement()
}

// sortedExtensionNamespaces returns the namespace URIs declared on w, in
// stable order, for deterministic extension-hook invocation.
func sortedExtensionNamespaces(w *Writer) []string {
	out := make([]string, 0, len(w.prefixURI))
	for prefix, uri := range w.prefixURI {
		if prefix == "" {
			continue
		}
		out = append(out, uri)
	}
	sort.Strings(out)
	return out
}

// formatU32 formats a uint32 in decimal with no allocations beyond the
// returned string.
func formatU32(u uint32) string { return strconv.FormatUint(uint64(u), 10) }

// formatFloat formats a float64 using the shortest decimal representation
// that round-trips to the same value.
func formatFloat(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }
