package tmf

import (
	"maps"

	"github.com/lestrrat-go/option/v3"
)

// Model is the in-memory representation of a single 3MF model part
// (3dmodel.model). A Package may hold multiple Models when the Production
// extension is in use; the primary Model is reached via Package.Model().
type Model struct {
	unit      Unit
	language  string
	thumbnail string
	metadata  []*Metadata

	resources *Resources
	build     *Build

	// requiredExtensions is the set of extension namespaces this model has
	// declared as required, keyed by URI -> conventional prefix.
	requiredExtensions map[string]string

	// extensions holds arbitrary extension payloads attached at the model
	// level (typically used for cross-cutting extension state).
	extensions map[string]any
}

// NewModel constructs a Model from functional options. The model's Resources
// and Build are always non-nil after construction so callers can append to
// them directly.
func NewModel(opts ...Option) *Model {
	m := &Model{
		unit:      UnitMillimeter,
		resources: &Resources{},
		build:     &Build{},
	}
	for _, o := range opts {
		switch o.Ident() {
		case identUnit{}:
			m.unit = option.MustGet[Unit](o)
		case identLanguage{}:
			m.language = option.MustGet[string](o)
		case identThumbnail{}:
			m.thumbnail = option.MustGet[string](o)
		case identModelMetadata{}:
			m.metadata = append(m.metadata, option.MustGet[*Metadata](o))
		case identResources{}:
			m.resources = option.MustGet[*Resources](o)
		case identObjectResource{}:
			m.resources.AppendObject(option.MustGet[*Object](o))
		case identBaseMaterialsResource{}:
			m.resources.AppendBaseMaterials(option.MustGet[*BaseMaterials](o))
		case identBuildItem{}:
			m.build.AppendItem(option.MustGet[*BuildItem](o))
		case identRequiredExtension{}:
			re := option.MustGet[requiredExtension](o)
			if m.requiredExtensions == nil {
				m.requiredExtensions = make(map[string]string)
			}
			m.requiredExtensions[re.namespace] = re.prefix
		case identExtensionPayload{}:
			p := option.MustGet[extensionPayload](o)
			if m.extensions == nil {
				m.extensions = make(map[string]any)
			}
			m.extensions[p.namespace] = p.value
		}
	}
	return m
}

// Unit returns the unit attribute on <model>.
func (m *Model) Unit() Unit { return m.unit }

// SetUnit sets the unit attribute on <model>.
func (m *Model) SetUnit(u Unit) { m.unit = u }

// Language returns the xml:lang attribute, or "".
func (m *Model) Language() string { return m.language }

// SetLanguage sets the xml:lang attribute.
func (m *Model) SetLanguage(s string) { m.language = s }

// Thumbnail returns the in-package thumbnail path, or "".
func (m *Model) Thumbnail() string { return m.thumbnail }

// SetThumbnail sets the in-package thumbnail path.
func (m *Model) SetThumbnail(s string) { m.thumbnail = s }

// Metadata returns the model's metadata entries.
func (m *Model) Metadata() []*Metadata { return m.metadata }

// AppendMetadata appends a metadata entry.
func (m *Model) AppendMetadata(md *Metadata) { m.metadata = append(m.metadata, md) }

// Resources returns the resource list (never nil).
func (m *Model) Resources() *Resources { return m.resources }

// Build returns the build list (never nil).
func (m *Model) Build() *Build { return m.build }

// RequiredExtensions returns a copy of the map URI -> declared prefix.
func (m *Model) RequiredExtensions() map[string]string {
	if len(m.requiredExtensions) == 0 {
		return nil
	}
	out := make(map[string]string, len(m.requiredExtensions))
	maps.Copy(out, m.requiredExtensions)
	return out
}

// RequireExtension declares that namespace ns is required by this model and
// should be prefixed with prefix when serialized.
func (m *Model) RequireExtension(ns, prefix string) {
	if m.requiredExtensions == nil {
		m.requiredExtensions = make(map[string]string)
	}
	m.requiredExtensions[ns] = prefix
}

// Extension returns an arbitrary extension payload attached at the model
// level, or nil.
func (m *Model) Extension(ns string) any {
	if m.extensions == nil {
		return nil
	}
	return m.extensions[ns]
}

// SetExtension attaches an extension payload. Passing nil removes it.
func (m *Model) SetExtension(ns string, v any) {
	if v == nil {
		delete(m.extensions, ns)
		return
	}
	if m.extensions == nil {
		m.extensions = make(map[string]any)
	}
	m.extensions[ns] = v
}

// Extensions returns a copy of the attached extension payloads.
func (m *Model) Extensions() map[string]any {
	if len(m.extensions) == 0 {
		return nil
	}
	out := make(map[string]any, len(m.extensions))
	maps.Copy(out, m.extensions)
	return out
}

// Option constructors for Model.

// WithUnit sets the unit attribute on a new Model.
func WithUnit(u Unit) Option { return option.New(identUnit{}, u) }

// WithLanguage sets the xml:lang attribute on a new Model.
func WithLanguage(s string) Option { return option.New(identLanguage{}, s) }

// WithThumbnail sets the in-package thumbnail path on a new Model.
func WithThumbnail(s string) Option { return option.New(identThumbnail{}, s) }

// WithModelMetadata appends a Metadata entry to a new Model.
func WithModelMetadata(md *Metadata) Option { return option.New(identModelMetadata{}, md) }

// WithResources replaces the Resources of a new Model.
func WithResources(r *Resources) Option { return option.New(identResources{}, r) }

// WithObject appends an Object resource to a new Model.
func WithObject(o *Object) Option { return option.New(identObjectResource{}, o) }

// WithObjects appends multiple Object resources to a new Model.
func WithObjects(objs ...*Object) []Option {
	out := make([]Option, len(objs))
	for i, o := range objs {
		out[i] = option.New(identObjectResource{}, o)
	}
	return out
}

// WithBaseMaterials appends a BaseMaterials resource to a new Model.
func WithBaseMaterials(b *BaseMaterials) Option {
	return option.New(identBaseMaterialsResource{}, b)
}

// WithBuildItem appends a BuildItem to a new Model.
func WithBuildItem(b *BuildItem) Option { return option.New(identBuildItem{}, b) }
