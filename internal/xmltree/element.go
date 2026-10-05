// Package xmltree provides the small XML tree used by 3MF extension readers.
package xmltree

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
	"strings"
)

// Node is an element or text child in an XML tree.
type Node interface{ isNode() }

// Element is an XML element with its attributes and children in source order.
type Element struct {
	name       xml.Name
	namespace  *Namespace
	namespaces []*Namespace
	attributes []*Attribute
	children   []Node
}

func (*Element) isNode() {}

// LocalName returns the element name without its namespace prefix.
func (e *Element) LocalName() string { return e.name.Local }

// Namespace returns the element's namespace, or nil for an unqualified name.
func (e *Element) Namespace() *Namespace { return e.namespace }

// Namespaces returns namespace declarations made on this element.
func (e *Element) Namespaces() []*Namespace { return e.namespaces }

// Attributes returns the element's attributes in source order.
func (e *Element) Attributes() []*Attribute { return e.attributes }

// FindAttribute returns the first attribute matching predicate.
func (e *Element) FindAttribute(predicate AttributePredicate) (*Attribute, bool) {
	for _, a := range e.attributes {
		if predicate.Match(a) {
			return a, true
		}
	}
	return nil, false
}

// Namespace records a prefix and the URI to which it is bound.
type Namespace struct {
	prefix string
	uri    string
}

// Prefix returns the declared namespace prefix.
func (n *Namespace) Prefix() string { return n.prefix }

// URI returns the namespace URI.
func (n *Namespace) URI() string { return n.uri }

// Attribute is an XML attribute with a resolved namespace URI.
type Attribute struct {
	name  string
	local string
	uri   string
	value string
}

// Name returns the attribute's qualified name.
func (a *Attribute) Name() string { return a.name }

// Value returns the attribute value.
func (a *Attribute) Value() string { return a.value }

// NamespaceURI returns the resolved attribute namespace URI.
func (a *Attribute) NamespaceURI() string { return a.uri }

// AttributePredicate matches an attribute by name or namespace.
type AttributePredicate interface{ Match(*Attribute) bool }

type localNamePredicate string

func (p localNamePredicate) Match(a *Attribute) bool { return a.local == string(p) }

// LocalNamePredicate matches attributes with the given local name.
func LocalNamePredicate(local string) AttributePredicate { return localNamePredicate(local) }

// NSPredicate matches an attribute by local name and namespace URI.
type NSPredicate struct {
	Local        string
	NamespaceURI string
}

// Match reports whether a has the predicate's name and namespace URI.
func (p NSPredicate) Match(a *Attribute) bool {
	return a.local == p.Local && a.uri == p.NamespaceURI
}

// Text is a text child of an element.
type Text struct{ content []byte }

func (*Text) isNode() {}

// Content returns the text bytes.
func (t *Text) Content() []byte { return t.content }

// Children iterates over an element's direct children in source order.
func Children(parent *Element) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for _, child := range parent.children {
			if !yield(child) {
				return
			}
		}
	}
}

// Document is the parsed XML document.
type Document struct{ root *Element }

// DocumentElement returns the document's root element.
func (d *Document) DocumentElement() *Element { return d.root }

// Parse reads XML data into an element tree and checks ctx between tokens.
func Parse(ctx context.Context, data []byte) (*Document, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	doc := &Document{}
	var stack []*Element
	var rawNames []xml.Name
	scopes := []map[string]string{{"xml": "http://www.w3.org/XML/1998/namespace"}}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		token, err := decoder.RawToken()
		if err == io.EOF {
			if len(stack) != 0 {
				return nil, fmt.Errorf("xmltree: unclosed element")
			}
			return doc, nil
		}
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case xml.StartElement:
			scope := scopes[len(scopes)-1]
			elem := &Element{}
			for _, attr := range token.Attr {
				prefix, ok := declaredPrefix(attr.Name)
				if !ok {
					continue
				}
				if len(elem.namespaces) == 0 {
					copyScope := make(map[string]string, len(scope)+1)
					for inheritedPrefix, uri := range scope {
						copyScope[inheritedPrefix] = uri
					}
					scope = copyScope
				}
				scope[prefix] = attr.Value
				elem.namespaces = append(elem.namespaces, &Namespace{prefix: prefix, uri: attr.Value})
			}
			uri := scope[""]
			if token.Name.Space != "" {
				var ok bool
				uri, ok = scope[token.Name.Space]
				if !ok || uri == "" {
					return nil, fmt.Errorf("xmltree: undeclared element prefix %q", token.Name.Space)
				}
			}
			elem.name = xml.Name{Space: uri, Local: token.Name.Local}
			if uri != "" {
				elem.namespace = &Namespace{prefix: token.Name.Space, uri: uri}
			}
			for _, attr := range token.Attr {
				if _, ok := declaredPrefix(attr.Name); ok {
					continue
				}
				name := attr.Name.Local
				uri := ""
				if attr.Name.Space != "" {
					var ok bool
					uri, ok = scope[attr.Name.Space]
					if !ok || uri == "" {
						return nil, fmt.Errorf("xmltree: undeclared attribute prefix %q", attr.Name.Space)
					}
					name = attr.Name.Space + ":" + name
				}
				elem.attributes = append(elem.attributes, &Attribute{
					name: name, local: attr.Name.Local, uri: uri, value: attr.Value,
				})
			}
			if len(stack) == 0 {
				if doc.root != nil {
					return nil, fmt.Errorf("xmltree: multiple root elements")
				}
				doc.root = elem
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, elem)
			}
			stack = append(stack, elem)
			rawNames = append(rawNames, token.Name)
			scopes = append(scopes, scope)
		case xml.EndElement:
			if len(rawNames) == 0 || rawNames[len(rawNames)-1] != token.Name {
				return nil, fmt.Errorf("xmltree: mismatched end element %q", token.Name.Local)
			}
			stack = stack[:len(stack)-1]
			rawNames = rawNames[:len(rawNames)-1]
			scopes = scopes[:len(scopes)-1]
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(token)) != "" {
					return nil, fmt.Errorf("xmltree: text outside root element")
				}
				continue
			}
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, &Text{content: bytes.Clone(token)})
		}
	}
}

func declaredPrefix(name xml.Name) (string, bool) {
	if name.Space == "xmlns" {
		return name.Local, true
	}
	if name.Space == "" && name.Local == "xmlns" {
		return "", true
	}
	return "", false
}
