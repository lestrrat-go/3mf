package tmf

import (
	"sync"

	"github.com/lestrrat-go/helium"
)

// ExtensionReader is implemented by extension sub-packages and registered via
// RegisterExtensionReader. The core 3MF reader invokes the appropriate hook
// when it encounters an element belonging to the extension's namespace at a
// position where the extension is allowed to contribute.
//
// All hooks have a default no-op implementation; extensions need only
// implement the ones that apply to them. ExtensionReader is intentionally
// minimal — it covers the four contexts where 3MF extensions inject new
// elements: top-level <resources>, individual <object>, individual <mesh>,
// and the model's <build>.
type ExtensionReader interface {
	// Namespace returns the URI of the namespace this reader handles.
	Namespace() string

	// ReadResourceElement is invoked for each direct child of <resources>
	// whose namespace matches Namespace().
	ReadResourceElement(res *Resources, elem *helium.Element) error

	// ReadObjectElement is invoked for each direct child of <object> whose
	// namespace matches Namespace().
	ReadObjectElement(obj *Object, elem *helium.Element) error

	// ReadMeshElement is invoked for each direct child of <mesh> whose
	// namespace matches Namespace().
	ReadMeshElement(mesh *Mesh, elem *helium.Element) error

	// ReadBuildElement is invoked for each direct child of <build> whose
	// namespace matches Namespace().
	ReadBuildElement(b *Build, elem *helium.Element) error
}

// ExtensionWriter is the symmetric write-side hook for an extension.
// Each method is called by the core writer at the corresponding position in
// the XML stream so that the extension may emit additional elements.
//
// Implementations must use the supplied *Writer's StartElementNS/AttrNS
// helpers with the extension's pre-declared prefix; they must not close
// the elements opened by the core writer.
type ExtensionWriter interface {
	// Namespace returns the URI of the namespace this writer handles.
	Namespace() string

	// WriteResourceElements is invoked once per model, after all core
	// resource elements have been written but while the <resources> element
	// is still open.
	WriteResourceElements(res *Resources, w *Writer) error

	// WriteObjectElements is invoked once per object, after the core
	// <mesh>/<components> child has been written but while the <object>
	// element is still open.
	WriteObjectElements(obj *Object, w *Writer) error

	// WriteMeshElements is invoked once per mesh, after the core
	// <vertices>/<triangles> children have been written but while the
	// <mesh> element is still open.
	WriteMeshElements(mesh *Mesh, w *Writer) error

	// WriteBuildElements is invoked once per build, after the core
	// <item> children have been written but while the <build> element
	// is still open.
	WriteBuildElements(b *Build, w *Writer) error
}

var (
	extMu      sync.RWMutex
	extReaders = map[string]ExtensionReader{}
	extWriters = map[string]ExtensionWriter{}
)

// RegisterExtensionReader installs r as the reader for r.Namespace(). It is
// safe to call from package init functions. Re-registering a namespace
// replaces the previous handler.
func RegisterExtensionReader(r ExtensionReader) {
	extMu.Lock()
	extReaders[r.Namespace()] = r
	extMu.Unlock()
}

// RegisterExtensionWriter installs w as the writer for w.Namespace().
func RegisterExtensionWriter(w ExtensionWriter) {
	extMu.Lock()
	extWriters[w.Namespace()] = w
	extMu.Unlock()
}

// LookupExtensionReader returns the registered reader for ns, or nil.
func LookupExtensionReader(ns string) ExtensionReader {
	extMu.RLock()
	defer extMu.RUnlock()
	return extReaders[ns]
}

// LookupExtensionWriter returns the registered writer for ns, or nil.
func LookupExtensionWriter(ns string) ExtensionWriter {
	extMu.RLock()
	defer extMu.RUnlock()
	return extWriters[ns]
}

// BaseExtensionReader is a convenience embedding that provides no-op
// implementations of every ExtensionReader hook. Extension implementations
// embed it and override only the hooks they care about.
type BaseExtensionReader struct{}

func (BaseExtensionReader) ReadResourceElement(*Resources, *helium.Element) error { return nil }
func (BaseExtensionReader) ReadObjectElement(*Object, *helium.Element) error      { return nil }
func (BaseExtensionReader) ReadMeshElement(*Mesh, *helium.Element) error          { return nil }
func (BaseExtensionReader) ReadBuildElement(*Build, *helium.Element) error        { return nil }

// BaseExtensionWriter is the write-side analog of BaseExtensionReader.
type BaseExtensionWriter struct{}

func (BaseExtensionWriter) WriteResourceElements(*Resources, *Writer) error { return nil }
func (BaseExtensionWriter) WriteObjectElements(*Object, *Writer) error      { return nil }
func (BaseExtensionWriter) WriteMeshElements(*Mesh, *Writer) error          { return nil }
func (BaseExtensionWriter) WriteBuildElements(*Build, *Writer) error        { return nil }
