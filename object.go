package tmf

import (
	"maps"

	"github.com/lestrrat-go/option/v3"
)

// Object is a 3MF resource that holds either a mesh or a list of components
// (or, in the case of certain extensions, alternative geometry such as a
// beam lattice or implicit function reference).
type Object struct {
	id         uint32
	name       string
	objType    ObjectType
	uuid       string
	partNumber string
	thumbnail  string

	// pid / pIndex form an optional reference to a property group that
	// applies to the whole object unless overridden per triangle.
	pid    uint32
	pIndex uint32
	hasPID bool

	mesh       *Mesh
	components []*Component

	extensions map[string]any
}

// NewObject constructs an Object from functional options.
func NewObject(opts ...Option) *Object {
	obj := &Object{}
	for _, o := range opts {
		switch o.Ident() {
		case identObjectID{}:
			obj.id = option.MustGet[uint32](o)
		case identObjectName{}:
			obj.name = option.MustGet[string](o)
		case identObjectType{}:
			obj.objType = option.MustGet[ObjectType](o)
		case identObjectUUID{}:
			obj.uuid = option.MustGet[string](o)
		case identObjectPartNumber{}:
			obj.partNumber = option.MustGet[string](o)
		case identObjectThumbnail{}:
			obj.thumbnail = option.MustGet[string](o)
		case identObjectPID{}:
			obj.pid = option.MustGet[uint32](o)
			obj.hasPID = true
		case identObjectPIndex{}:
			obj.pIndex = option.MustGet[uint32](o)
			obj.hasPID = true
		case identMesh{}:
			obj.mesh = option.MustGet[*Mesh](o)
		case identComponents{}:
			obj.components = append(obj.components, option.MustGet[[]*Component](o)...)
		case identComponent{}:
			obj.components = append(obj.components, option.MustGet[*Component](o))
		case identExtensionPayload{}:
			p := option.MustGet[extensionPayload](o)
			if obj.extensions == nil {
				obj.extensions = make(map[string]any)
			}
			obj.extensions[p.namespace] = p.value
		}
	}
	return obj
}

// ID returns the object's id attribute, unique within its model part.
func (o *Object) ID() uint32 { return o.id }

// SetID overrides the object id.
func (o *Object) SetID(id uint32) { o.id = id }

// Name returns the human-readable name attribute, or "".
func (o *Object) Name() string { return o.name }

// SetName sets the human-readable name attribute.
func (o *Object) SetName(s string) { o.name = s }

// Type returns the ObjectType (defaults to ObjectTypeModel).
func (o *Object) Type() ObjectType { return o.objType }

// SetType sets the ObjectType.
func (o *Object) SetType(t ObjectType) { o.objType = t }

// UUID returns the Production-extension UUID, or "".
func (o *Object) UUID() string { return o.uuid }

// SetUUID sets the Production-extension UUID.
func (o *Object) SetUUID(s string) { o.uuid = s }

// PartNumber returns the part-number attribute, or "".
func (o *Object) PartNumber() string { return o.partNumber }

// SetPartNumber sets the part-number attribute.
func (o *Object) SetPartNumber(s string) { o.partNumber = s }

// Thumbnail returns the in-package thumbnail path, or "".
func (o *Object) Thumbnail() string { return o.thumbnail }

// SetThumbnail sets the in-package thumbnail path.
func (o *Object) SetThumbnail(s string) { o.thumbnail = s }

// PID returns the object-level property-group id and whether one was set.
func (o *Object) PID() (uint32, bool) { return o.pid, o.hasPID }

// PIndex returns the index into the property group named by PID. Only
// meaningful when PID is set.
func (o *Object) PIndex() uint32 { return o.pIndex }

// SetPID sets the object-level property-group id.
func (o *Object) SetPID(pid uint32) {
	o.pid = pid
	o.hasPID = true
}

// SetPIndex sets the per-object property index.
func (o *Object) SetPIndex(idx uint32) {
	o.pIndex = idx
	o.hasPID = true
}

// Mesh returns the mesh geometry attached to this object, or nil if the
// object is a components-only or extension-defined object.
func (o *Object) Mesh() *Mesh { return o.mesh }

// SetMesh attaches a mesh to the object, replacing any existing components.
func (o *Object) SetMesh(m *Mesh) {
	o.mesh = m
	o.components = nil
}

// Components returns the components participating in this object. The
// returned slice is owned by the object.
func (o *Object) Components() []*Component { return o.components }

// SetComponents replaces the component list.
func (o *Object) SetComponents(cs []*Component) {
	o.components = cs
	o.mesh = nil
}

// AppendComponent appends a single component.
func (o *Object) AppendComponent(c *Component) { o.components = append(o.components, c) }

// Extension returns an extension payload attached to this object, keyed by
// namespace URI.
func (o *Object) Extension(ns string) any {
	if o.extensions == nil {
		return nil
	}
	return o.extensions[ns]
}

// SetExtension attaches an extension payload. Passing nil removes it.
func (o *Object) SetExtension(ns string, v any) {
	if v == nil {
		delete(o.extensions, ns)
		return
	}
	if o.extensions == nil {
		o.extensions = make(map[string]any)
	}
	o.extensions[ns] = v
}

// Extensions returns a copy of the extension namespaces attached.
func (o *Object) Extensions() map[string]any {
	if len(o.extensions) == 0 {
		return nil
	}
	out := make(map[string]any, len(o.extensions))
	maps.Copy(out, o.extensions)
	return out
}

// Option constructors for Object.

// WithObjectID sets the id attribute on a new Object.
func WithObjectID(id uint32) Option { return option.New(identObjectID{}, id) }

// WithObjectName sets the human-readable name attribute on a new Object.
func WithObjectName(s string) Option { return option.New(identObjectName{}, s) }

// WithObjectType sets the type attribute on a new Object.
func WithObjectType(t ObjectType) Option { return option.New(identObjectType{}, t) }

// WithObjectUUID sets the Production-extension UUID on a new Object.
func WithObjectUUID(s string) Option { return option.New(identObjectUUID{}, s) }

// WithObjectPartNumber sets the part-number attribute on a new Object.
func WithObjectPartNumber(s string) Option { return option.New(identObjectPartNumber{}, s) }

// WithObjectThumbnail sets the thumbnail path on a new Object.
func WithObjectThumbnail(s string) Option { return option.New(identObjectThumbnail{}, s) }

// WithObjectPID sets the object-level property-group id.
func WithObjectPID(pid uint32) Option { return option.New(identObjectPID{}, pid) }

// WithObjectPIndex sets the object-level property index.
func WithObjectPIndex(idx uint32) Option { return option.New(identObjectPIndex{}, idx) }

// WithMesh attaches a Mesh to a new Object.
func WithMesh(m *Mesh) Option { return option.New(identMesh{}, m) }

// WithComponents attaches a slice of Components to a new Object.
func WithComponents(cs []*Component) Option { return option.New(identComponents{}, cs) }

// WithComponent appends a single Component to a new Object.
func WithComponent(c *Component) Option { return option.New(identComponent{}, c) }
