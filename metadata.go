package tmf

import "github.com/lestrrat-go/option/v3"

// Metadata is a key/value annotation attached to a Model, Object, or
// BuildItem. Names use the form "namespace:localname"; the 3MF specification
// reserves a fixed set of namespace-free names (Title, Designer, etc.) and
// allows arbitrary additional names when the namespace is declared on the
// owning element.
type Metadata struct {
	Name     string
	Value    string
	Type     string
	Preserve bool
}

// NewMetadata constructs a Metadata from functional options.
func NewMetadata(opts ...Option) *Metadata {
	m := &Metadata{}
	for _, o := range opts {
		switch o.Ident() {
		case identMetadataName{}:
			m.Name = option.MustGet[string](o)
		case identMetadataValue{}:
			m.Value = option.MustGet[string](o)
		case identMetadataType{}:
			m.Type = option.MustGet[string](o)
		case identMetadataPreserve{}:
			m.Preserve = option.MustGet[bool](o)
		}
	}
	return m
}

// WithMetadataName sets the name attribute on a new Metadata.
func WithMetadataName(s string) Option { return option.New(identMetadataName{}, s) }

// WithMetadataValue sets the text body of a new Metadata.
func WithMetadataValue(s string) Option { return option.New(identMetadataValue{}, s) }

// WithMetadataType sets the type attribute on a new Metadata.
func WithMetadataType(s string) Option { return option.New(identMetadataType{}, s) }

// WithMetadataPreserve sets the preserve attribute on a new Metadata.
func WithMetadataPreserve(b bool) Option { return option.New(identMetadataPreserve{}, b) }

// ReservedMetadataNames is the set of metadata names reserved by the 3MF
// core specification. Producers should not use these names with custom
// semantics.
var ReservedMetadataNames = map[string]struct{}{
	"Title":            {},
	"Designer":         {},
	"Description":      {},
	"Copyright":        {},
	"LicenseTerms":     {},
	"Rating":           {},
	"CreationDate":     {},
	"ModificationDate": {},
	"Application":      {},
}
