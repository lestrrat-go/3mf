package tmf

import "github.com/lestrrat-go/option/v3"

// BuildItem is one entry in the model's build list. Each item references an
// Object resource (by id, or by id plus path when the Production extension
// is in use) and may apply an affine transform when the object is placed on
// the print bed.
type BuildItem struct {
	// ObjectID is the id of the referenced object in the model part named by
	// Path (or in the current part when Path is empty).
	ObjectID uint32

	// Path, when non-empty, identifies the model part holding the referenced
	// object (Production extension, p:path attribute).
	Path string

	// Transform is the affine transform applied to the referenced object
	// during build. IdentityMatrix means no transform is serialized.
	Transform Matrix

	// PartNumber is an optional human-readable part identifier.
	PartNumber string

	// UUID is the Production-extension UUID for this build item.
	UUID string

	// Metadata is the optional list of metadata entries attached to this
	// build item. The 3MF spec allows metadata on build items via the
	// production extension.
	Metadata []*Metadata
}

// NewBuildItem constructs a BuildItem from functional options.
func NewBuildItem(opts ...Option) *BuildItem {
	b := &BuildItem{Transform: IdentityMatrix()}
	for _, o := range opts {
		switch o.Ident() {
		case identBuildItemObjectRef{}:
			b.ObjectID = option.MustGet[*Object](o).ID()
		case identBuildItemObjectID{}:
			b.ObjectID = option.MustGet[uint32](o)
		case identBuildItemPath{}:
			b.Path = option.MustGet[string](o)
		case identBuildItemTransform{}:
			b.Transform = option.MustGet[Matrix](o)
		case identBuildItemPartNumber{}:
			b.PartNumber = option.MustGet[string](o)
		case identBuildItemUUID{}:
			b.UUID = option.MustGet[string](o)
		case identBuildItemMetadata{}:
			b.Metadata = append(b.Metadata, option.MustGet[*Metadata](o))
		}
	}
	return b
}

// WithObjectRef wires a BuildItem to an existing Object by reference; the
// item's ObjectID is taken from the object's id at construction time.
func WithObjectRef(o *Object) Option { return option.New(identBuildItemObjectRef{}, o) }

// WithObjectID sets a BuildItem's ObjectID directly (useful when the
// target object lives in a different model part referenced by WithItemPath).
func WithBuildObjectID(id uint32) Option { return option.New(identBuildItemObjectID{}, id) }

// WithItemPath sets the Production-extension path attribute on a BuildItem.
func WithItemPath(p string) Option { return option.New(identBuildItemPath{}, p) }

// WithItemTransform sets the affine transform on a BuildItem.
func WithItemTransform(m Matrix) Option { return option.New(identBuildItemTransform{}, m) }

// WithItemPartNumber sets the optional part-number on a BuildItem.
func WithItemPartNumber(s string) Option { return option.New(identBuildItemPartNumber{}, s) }

// WithItemUUID sets the Production-extension UUID on a BuildItem.
func WithItemUUID(s string) Option { return option.New(identBuildItemUUID{}, s) }

// WithItemMetadata appends a Metadata entry to a BuildItem.
func WithItemMetadata(m *Metadata) Option { return option.New(identBuildItemMetadata{}, m) }

// Build represents the model's <build> element: the ordered list of items
// to be manufactured, plus an optional Production-extension UUID.
type Build struct {
	UUID  string
	Items []*BuildItem
}

// AppendItem appends an item to the build list.
func (b *Build) AppendItem(item *BuildItem) { b.Items = append(b.Items, item) }
