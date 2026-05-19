package tmf

import (
	"io"

	"github.com/lestrrat-go/helium/stream"
)

// Writer is the streaming XML writer surface that the core serializer and the
// extension sub-packages share. It wraps a helium *stream.Writer and adds a
// shared map of namespace prefixes that the core writer has pre-declared on
// the root <model> element so that extension writers can emit
// prefix-qualified attributes and elements without triggering a duplicate
// xmlns declaration.
//
// Extension implementations must use Attr/Element/StartElement on this type
// instead of calling the underlying stream.Writer's *NS methods. Direct
// access to the underlying writer is available via Raw for output that
// genuinely does not interact with extension namespaces (e.g. plain text
// content).
type Writer struct {
	w         stream.Writer
	prefixURI map[string]string // prefix -> URI declared on root
	uriPrefix map[string]string // URI -> prefix declared on root
}

// NewWriter wraps an io.Writer in a stream.Writer and returns a *Writer ready
// for use. Callers typically prefer the higher-level WriteModel function.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:         stream.NewWriter(w),
		prefixURI: map[string]string{},
		uriPrefix: map[string]string{},
	}
}

// Raw returns the underlying stream.Writer. Use only when none of the
// helpers on Writer suffice.
func (w *Writer) Raw() *stream.Writer { return &w.w }

// DeclarePrefix records that prefix is bound to uri on the root element.
// Subsequent calls to Attr / Element with this prefix will emit a
// prefix-qualified name without re-declaring the namespace. DeclarePrefix
// does not itself write any output — the caller must emit the xmlns:prefix
// attribute, typically via the helpers used while writing the root element.
func (w *Writer) DeclarePrefix(prefix, uri string) {
	w.prefixURI[prefix] = uri
	w.uriPrefix[uri] = prefix
}

// PrefixFor returns the prefix declared for uri, or "" when none has been
// declared.
func (w *Writer) PrefixFor(uri string) string { return w.uriPrefix[uri] }

// Attr writes a plain attribute on the currently open element.
func (w *Writer) Attr(name, value string) error {
	return w.w.WriteAttribute(name, value)
}

// AttrNS writes an attribute whose qualified name is prefix:localName. The
// prefix is expected to have been declared via DeclarePrefix; this method
// does not emit any xmlns declaration.
func (w *Writer) AttrNS(prefix, localName, value string) error {
	if prefix == "" {
		return w.w.WriteAttribute(localName, value)
	}
	return w.w.WriteAttribute(prefix+":"+localName, value)
}

// StartElement opens an element with the given local name (no namespace).
func (w *Writer) StartElement(name string) error { return w.w.StartElement(name) }

// StartElementNS opens an element with the given prefix and local name.
// The prefix must have been declared via DeclarePrefix.
func (w *Writer) StartElementNS(prefix, localName string) error {
	if prefix == "" {
		return w.w.StartElement(localName)
	}
	return w.w.StartElement(prefix + ":" + localName)
}

// EndElement closes the currently open element.
func (w *Writer) EndElement() error { return w.w.EndElement() }

// WriteString writes escaped text content.
func (w *Writer) WriteString(s string) error { return w.w.WriteString(s) }
