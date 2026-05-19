package tmf

import "github.com/lestrrat-go/option/v3"

// Component is a reference to another Object that participates in the
// containing Object's geometry, optionally with an affine transform applied.
type Component struct {
	// ObjectID is the id attribute of the referenced object in the same
	// model part. When Path is non-empty (Production extension) the reference
	// is resolved against that part instead.
	ObjectID uint32

	// Path, when non-empty, identifies the model part that holds the
	// referenced object (Production extension, p:path attribute).
	Path string

	// UUID is the Production-extension UUID of this component instance.
	UUID string

	// Transform is the affine transform applied to the referenced object.
	// IdentityMatrix is treated as "no transform" and is not serialized.
	Transform Matrix
}

// NewComponent constructs a Component from functional options.
func NewComponent(opts ...Option) *Component {
	c := &Component{Transform: IdentityMatrix()}
	for _, o := range opts {
		switch o.Ident() {
		case identComponentObjectRef{}:
			c.ObjectID = option.MustGet[*Object](o).ID()
		case identComponentTransform{}:
			c.Transform = option.MustGet[Matrix](o)
		case identComponentUUID{}:
			c.UUID = option.MustGet[string](o)
		case identComponentPath{}:
			c.Path = option.MustGet[string](o)
		}
	}
	return c
}

// WithComponentObjectRef sets the referenced Object on a Component. The
// component's ObjectID is taken from the object's id.
func WithComponentObjectRef(o *Object) Option {
	return option.New(identComponentObjectRef{}, o)
}

// WithComponentTransform sets the affine transform on a Component.
func WithComponentTransform(m Matrix) Option {
	return option.New(identComponentTransform{}, m)
}

// WithComponentUUID sets the Production-extension UUID on a Component.
func WithComponentUUID(u string) Option { return option.New(identComponentUUID{}, u) }

// WithComponentPath sets the Production-extension path attribute on a Component.
func WithComponentPath(p string) Option { return option.New(identComponentPath{}, p) }
