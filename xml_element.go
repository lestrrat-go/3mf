package tmf

import (
	"iter"
	"strings"

	"github.com/lestrrat-go/3mf/internal/xmltree"
)

// Element is the parsed XML subtree passed to an ExtensionReader hook.
type Element struct{ raw *xmltree.Element }

func wrapElement(raw *xmltree.Element) *Element { return &Element{raw: raw} }

// LocalName returns the element name without its namespace prefix.
func (e *Element) LocalName() string { return e.raw.LocalName() }

// NamespaceURI returns the namespace URI, or an empty string for an unqualified element.
func (e *Element) NamespaceURI() string {
	if namespace := e.raw.Namespace(); namespace != nil {
		return namespace.URI()
	}
	return ""
}

// Attr returns the first attribute with the given local name, regardless of namespace.
func (e *Element) Attr(local string) string {
	attr, ok := e.raw.FindAttribute(xmltree.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return attr.Value()
}

// AttrNS returns the attribute with the given local name and namespace URI.
func (e *Element) AttrNS(local, namespaceURI string) string {
	attr, ok := e.raw.FindAttribute(xmltree.NSPredicate{Local: local, NamespaceURI: namespaceURI})
	if !ok {
		return ""
	}
	return attr.Value()
}

// ElementAttribute is an attribute on a parsed extension element.
type ElementAttribute struct {
	Name         string
	NamespaceURI string
	Value        string
}

// Attributes returns the element's attributes in source order.
func (e *Element) Attributes() []ElementAttribute {
	attrs := e.raw.Attributes()
	out := make([]ElementAttribute, 0, len(attrs))
	for _, attr := range attrs {
		out = append(out, ElementAttribute{
			Name: attr.Name(), NamespaceURI: attr.NamespaceURI(), Value: attr.Value(),
		})
	}
	return out
}

// ChildElements iterates over direct child elements, optionally filtered by local name.
// An empty local name includes every element child.
func (e *Element) ChildElements(local string) iter.Seq[*Element] {
	return func(yield func(*Element) bool) {
		for child := range xmltree.Children(e.raw) {
			elem, ok := child.(*xmltree.Element)
			if !ok || (local != "" && elem.LocalName() != local) {
				continue
			}
			if !yield(wrapElement(elem)) {
				return
			}
		}
	}
}

// TextContent concatenates direct text children, including CDATA content.
func (e *Element) TextContent() string {
	var value strings.Builder
	for child := range xmltree.Children(e.raw) {
		if text, ok := child.(*xmltree.Text); ok {
			value.Write(text.Content())
		}
	}
	return value.String()
}
