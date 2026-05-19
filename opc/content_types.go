package opc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/lestrrat-go/helium"
	"github.com/lestrrat-go/helium/stream"
)

// ContentTypeNS is the XML namespace for [Content_Types].xml.
const ContentTypeNS = "http://schemas.openxmlformats.org/package/2006/content-types"

// ContentTypes is the in-memory representation of an OPC [Content_Types].xml
// part. It maps part extensions and part names to MIME types so that
// consumers can look up the type of any part in the package.
type ContentTypes struct {
	// Defaults maps lower-cased file extensions (without leading dot) to
	// the MIME type used for parts that have no Override entry.
	Defaults map[string]string

	// Overrides maps absolute part names to MIME types. Override entries
	// take precedence over Defaults.
	Overrides map[string]string
}

// NewContentTypes returns an empty ContentTypes with the maps initialized.
func NewContentTypes() *ContentTypes {
	return &ContentTypes{
		Defaults:  map[string]string{},
		Overrides: map[string]string{},
	}
}

// AddDefault registers a Default <ext, content-type> entry.
func (c *ContentTypes) AddDefault(ext, contentType string) {
	if c.Defaults == nil {
		c.Defaults = map[string]string{}
	}
	c.Defaults[strings.ToLower(strings.TrimPrefix(ext, "."))] = contentType
}

// AddOverride registers an Override <partName, content-type> entry. partName
// is normalized to an absolute OPC part name.
func (c *ContentTypes) AddOverride(partName, contentType string) {
	if c.Overrides == nil {
		c.Overrides = map[string]string{}
	}
	c.Overrides[NormalizePartName(partName)] = contentType
}

// Lookup returns the MIME type registered for partName, consulting Overrides
// first and then Defaults. An empty string is returned when no match is found.
func (c *ContentTypes) Lookup(partName string) string {
	p := NormalizePartName(partName)
	if t, ok := c.Overrides[p]; ok {
		return t
	}
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(p)), ".")
	return c.Defaults[ext]
}

// ParseContentTypes reads a [Content_Types].xml payload and returns the
// parsed ContentTypes value.
func ParseContentTypes(data []byte) (*ContentTypes, error) {
	doc, err := helium.NewParser().Parse(context.Background(), data)
	if err != nil {
		return nil, fmt.Errorf("opc: parse content types: %w", err)
	}
	root := doc.DocumentElement()
	if root == nil || root.LocalName() != "Types" {
		return nil, fmt.Errorf("opc: content types root must be <Types>")
	}
	ct := NewContentTypes()
	for child := range helium.Children(root) {
		elem, ok := child.(*helium.Element)
		if !ok {
			continue
		}
		switch elem.LocalName() {
		case "Default":
			ext := attr(elem, "Extension")
			typ := attr(elem, "ContentType")
			if ext != "" && typ != "" {
				ct.AddDefault(ext, typ)
			}
		case "Override":
			part := attr(elem, "PartName")
			typ := attr(elem, "ContentType")
			if part != "" && typ != "" {
				ct.AddOverride(part, typ)
			}
		}
	}
	return ct, nil
}

// WriteTo serializes c as a [Content_Types].xml payload.
func (c *ContentTypes) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}
	sw := stream.NewWriter(cw)
	if err := sw.StartDocument("1.0", "UTF-8", "yes"); err != nil {
		return cw.n, err
	}
	if err := sw.StartElement("Types"); err != nil {
		return cw.n, err
	}
	if err := sw.WriteAttribute("xmlns", ContentTypeNS); err != nil {
		return cw.n, err
	}
	for ext, typ := range sortedKV(c.Defaults) {
		if err := sw.StartElement("Default"); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("Extension", ext); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("ContentType", typ); err != nil {
			return cw.n, err
		}
		if err := sw.EndElement(); err != nil {
			return cw.n, err
		}
	}
	for part, typ := range sortedKV(c.Overrides) {
		if err := sw.StartElement("Override"); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("PartName", part); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("ContentType", typ); err != nil {
			return cw.n, err
		}
		if err := sw.EndElement(); err != nil {
			return cw.n, err
		}
	}
	if err := sw.EndElement(); err != nil {
		return cw.n, err
	}
	if err := sw.EndDocument(); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// Bytes returns the serialized [Content_Types].xml payload.
func (c *ContentTypes) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// attr returns the first attribute on elem with the given local name,
// regardless of namespace.
func attr(elem *helium.Element, local string) string {
	a, ok := elem.FindAttribute(helium.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return a.Value()
}

// countingWriter is a thin io.Writer that tracks how many bytes were written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
