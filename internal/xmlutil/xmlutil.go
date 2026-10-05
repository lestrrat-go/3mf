// Package xmlutil contains small helpers shared by the core reader/writer
// and the extension sub-packages. It is intentionally unexported.
package xmlutil

import (
	"strconv"
	"strings"

	"github.com/lestrrat-go/3mf/internal/xmltree"
)

// Attr returns the value of the first attribute on elem whose local name
// equals local, regardless of namespace. An empty string is returned when
// no such attribute exists.
func Attr(elem *xmltree.Element, local string) string {
	a, ok := elem.FindAttribute(xmltree.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return a.Value()
}

// AttrNS returns the value of the first attribute on elem with the given
// local name and namespace URI.
func AttrNS(elem *xmltree.Element, local, ns string) string {
	a, ok := elem.FindAttribute(xmltree.NSPredicate{Local: local, NamespaceURI: ns})
	if !ok {
		return ""
	}
	return a.Value()
}

// AttrUint32 parses a uint32 attribute, returning 0 when absent or malformed.
// The "ok" return distinguishes "absent" from "zero".
func AttrUint32(elem *xmltree.Element, local string) (uint32, bool) {
	s := Attr(elem, local)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

// AttrFloat64 parses a float64 attribute, returning 0 when absent or malformed.
func AttrFloat64(elem *xmltree.Element, local string) (float64, bool) {
	s := Attr(elem, local)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// ChildElements iterates over child elements of parent, optionally filtered
// by local name. Pass "" to get every element child.
func ChildElements(parent *xmltree.Element, local string) func(yield func(*xmltree.Element) bool) {
	return func(yield func(*xmltree.Element) bool) {
		for child := range xmltree.Children(parent) {
			elem, ok := child.(*xmltree.Element)
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

// TextContent returns the concatenation of direct text children of elem.
// The decoder also returns CDATA as text. Element children are ignored.
func TextContent(elem *xmltree.Element) string {
	var b strings.Builder
	for child := range xmltree.Children(elem) {
		if v, ok := child.(*xmltree.Text); ok {
			b.Write(v.Content())
		}
	}
	return b.String()
}

// ParseFloats parses a whitespace-separated list of decimal floats from s.
// Returned slice has the same length as the number of fields in s.
func ParseFloats(s string) ([]float64, error) {
	fields := strings.Fields(s)
	out := make([]float64, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
