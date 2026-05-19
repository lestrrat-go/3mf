// Package xmlutil contains small helpers shared by the core reader/writer
// and the extension sub-packages. It is intentionally unexported.
package xmlutil

import (
	"strconv"
	"strings"

	"github.com/lestrrat-go/helium"
)

// Attr returns the value of the first attribute on elem whose local name
// equals local, regardless of namespace. An empty string is returned when
// no such attribute exists.
func Attr(elem *helium.Element, local string) string {
	a, ok := elem.FindAttribute(helium.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return a.Value()
}

// AttrNS returns the value of the first attribute on elem with the given
// local name and namespace URI.
func AttrNS(elem *helium.Element, local, ns string) string {
	a, ok := elem.FindAttribute(helium.NSPredicate{Local: local, NamespaceURI: ns})
	if !ok {
		return ""
	}
	return a.Value()
}

// AttrUint32 parses a uint32 attribute, returning 0 when absent or malformed.
// The "ok" return distinguishes "absent" from "zero".
func AttrUint32(elem *helium.Element, local string) (uint32, bool) {
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
func AttrFloat64(elem *helium.Element, local string) (float64, bool) {
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
func ChildElements(parent *helium.Element, local string) func(yield func(*helium.Element) bool) {
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

// TextContent returns the concatenation of all direct Text/CDATA children of
// elem. Comments, processing instructions, and element children are ignored.
func TextContent(elem *helium.Element) string {
	var b strings.Builder
	for child := range helium.Children(elem) {
		switch v := child.(type) {
		case *helium.Text:
			b.Write(v.Content())
		case *helium.CDATASection:
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
