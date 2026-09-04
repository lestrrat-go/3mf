package tmf

import "github.com/lestrrat-go/option/v3"

// Option is the type passed to the New* constructors in this package and to
// the extension sub-packages. It is a type alias for
// github.com/lestrrat-go/option/v3.Interface.
type Option = option.Interface

// Identifier types for options consumed by core constructors. Each identifier
// is an unexported zero-size struct so that only this package may construct
// the corresponding Option values.
type (
	identUnit                  struct{}
	identLanguage              struct{}
	identThumbnail             struct{}
	identModelMetadata         struct{}
	identResources             struct{}
	identObjectResource        struct{}
	identBaseMaterialsResource struct{}
	identBuildItem             struct{}
	identExtensionPayload      struct{}
	identRequiredExtension     struct{}

	identObjectID         struct{}
	identObjectName       struct{}
	identObjectType       struct{}
	identObjectUUID       struct{}
	identObjectPartNumber struct{}
	identObjectThumbnail  struct{}
	identObjectPID        struct{}
	identObjectPIndex     struct{}
	identMesh             struct{}
	identComponent        struct{}
	identComponents       struct{}

	identComponentObjectRef struct{}
	identComponentTransform struct{}
	identComponentUUID      struct{}
	identComponentPath      struct{}

	identVertices  struct{}
	identTriangles struct{}

	identMetadataName     struct{}
	identMetadataValue    struct{}
	identMetadataType     struct{}
	identMetadataPreserve struct{}

	identBuildItemObjectRef  struct{}
	identBuildItemObjectID   struct{}
	identBuildItemPath       struct{}
	identBuildItemTransform  struct{}
	identBuildItemPartNumber struct{}
	identBuildItemUUID       struct{}
	identBuildItemMetadata   struct{}
)

// extensionPayload is the value carried by WithExtension; it associates an
// arbitrary value with the namespace URI of the extension that produced it.
type extensionPayload struct {
	namespace string
	value     any
}

// requiredExtension is the value carried by WithRequiredExtension; it tells
// the writer to declare a prefix for the given namespace and to list it in
// the model's requiredextensions attribute.
type requiredExtension struct {
	namespace string
	prefix    string
}

// WithExtension attaches an extension payload identified by its namespace URI
// to the surrounding Mesh, Object, Build, or Model value. Extension packages
// (materials, beamlattice, etc.) typically wrap this option with a more
// strongly typed helper.
func WithExtension(namespace string, value any) Option {
	return option.New(identExtensionPayload{}, extensionPayload{namespace: namespace, value: value})
}

// WithRequiredExtension instructs the writer to declare a namespace prefix
// for the given extension namespace URI on the root model element and to
// list it in the requiredextensions attribute.
func WithRequiredExtension(namespace, prefix string) Option {
	return option.New(identRequiredExtension{}, requiredExtension{namespace: namespace, prefix: prefix})
}
