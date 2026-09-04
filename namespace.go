package tmf

// Namespace URIs used in 3MF documents and OPC packages.
const (
	// NSCore is the 3MF core specification namespace.
	NSCore = "http://schemas.microsoft.com/3dmanufacturing/core/2015/02"

	// NSMaterials is the 3MF Materials and Properties Extension namespace.
	NSMaterials = "http://schemas.microsoft.com/3dmanufacturing/material/2015/02"

	// NSProduction is the 3MF Production Extension namespace.
	NSProduction = "http://schemas.microsoft.com/3dmanufacturing/production/2015/06"

	// NSBeamLattice is the 3MF Beam Lattice Extension namespace.
	NSBeamLattice = "http://schemas.microsoft.com/3dmanufacturing/beamlattice/2017/02"

	// NSSlice is the 3MF Slice Extension namespace.
	NSSlice = "http://schemas.microsoft.com/3dmanufacturing/slice/2015/07"

	// NSSecureContent is the 3MF Secure Content Extension namespace.
	NSSecureContent = "http://schemas.microsoft.com/3dmanufacturing/securecontent/2019/04"

	// NSVolumetric is the 3MF Volumetric Extension namespace.
	NSVolumetric = "http://schemas.microsoft.com/3dmanufacturing/volumetric/2022/01"

	// NSXML is the XML namespace itself (used for xml:lang etc.).
	NSXML = "http://www.w3.org/XML/1998/namespace"

	// NSOPCContentTypes is the OPC Content Types namespace.
	NSOPCContentTypes = "http://schemas.openxmlformats.org/package/2006/content-types"

	// NSOPCRelationships is the OPC Relationships namespace.
	NSOPCRelationships = "http://schemas.openxmlformats.org/package/2006/relationships"

	// NSOPCMetadataCoreProperties is the OPC Core Properties namespace.
	NSOPCMetadataCoreProperties = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
)

// Conventional prefixes used when writing 3MF XML. These are the same prefixes
// used by Microsoft's reference implementation and are widely recognized by
// 3MF consumers.
const (
	PrefixMaterials     = "m"
	PrefixProduction    = "p"
	PrefixBeamLattice   = "b"
	PrefixSlice         = "s"
	PrefixSecureContent = "sc"
	PrefixVolumetric    = "v"
)

// Relationship type URIs.
const (
	// RelTypeStartPart is the OPC relationship type for the primary 3MF model part.
	RelTypeStartPart = "http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"

	// RelTypeThumbnail is the OPC relationship type for a package-level thumbnail.
	RelTypeThumbnail = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"

	// RelTypeTexture is the relationship type for a texture referenced from a model part.
	RelTypeTexture = "http://schemas.microsoft.com/3dmanufacturing/2013/01/3dtexture"

	// RelTypePrintTicket is the relationship type for a print ticket.
	RelTypePrintTicket = "http://schemas.microsoft.com/3dmanufacturing/2013/01/printticket"

	// RelTypeCoreProperties is the relationship type for OPC core metadata.
	RelTypeCoreProperties = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"

	// RelTypeMustPreserve is the relationship type used by the Production
	// extension to flag non-root model parts that must be preserved.
	RelTypeMustPreserve = "http://schemas.openxmlformats.org/package/2006/relationships/mustpreserve"
)

// Content types for parts contained in a 3MF package.
const (
	ContentType3DModel        = "application/vnd.ms-package.3dmanufacturing-3dmodel+xml"
	ContentTypePrintTicket    = "application/vnd.ms-printing.printticket+xml"
	ContentTypeCoreProperties = "application/vnd.openxmlformats-package.core-properties+xml"
	ContentTypeRelationships  = "application/vnd.openxmlformats-package.relationships+xml"
	ContentTypePNG            = "image/png"
	ContentTypeJPEG           = "image/jpeg"
	ContentTypeTexture        = "application/vnd.ms-package.3dmanufacturing-3dmodeltexture"
)

// DefaultModelPath is the conventional path for the root 3MF model part.
const DefaultModelPath = "/3D/3dmodel.model"

// modelElementName is the local name of the root element of a 3MF model part.
const modelElementName = "model"
