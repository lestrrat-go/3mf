package materials

import tmf "github.com/lestrrat-go/3mf"

// Namespace is the URI of the 3MF Materials extension.
const Namespace = tmf.NSMaterials

// Prefix is the conventional prefix used for the Materials extension.
const Prefix = tmf.PrefixMaterials

// Resources is the materials-extension payload attached to tmf.Resources
// (accessed via Resources.ExtensionResources(Namespace)).
type Resources struct {
	ColorGroups        []*ColorGroup
	Texture2Ds         []*Texture2D
	Texture2DGroups    []*Texture2DGroup
	CompositeMaterials []*CompositeMaterials
	MultiProperties    []*MultiProperties
}

// ColorGroup is a list of colors that triangles or objects reference by
// (PID, PIndex). One ColorGroup typically holds a small palette.
type ColorGroup struct {
	ID     uint32
	Colors []tmf.Color
}

// Texture2D is an image resource that lives elsewhere in the OPC package,
// referenced by Path. The image bytes themselves are accessed via the
// owning tmf.Package.Part(Path).
type Texture2D struct {
	ID          uint32
	Path        string
	ContentType string
	TileStyleU  string // "wrap" | "mirror" | "clamp" | "none"
	TileStyleV  string
	Filter      string // "auto" | "linear" | "nearest"
	BoxMin      *Vec2  // optional sub-rectangle within the texture
	BoxMax      *Vec2
}

// Texture2DGroup combines per-vertex UV coordinates with a Texture2D
// reference. Triangles use Texture2DGroup.ID via the standard PID attribute,
// and the per-triangle p1/p2/p3 indices select rows from Coords.
type Texture2DGroup struct {
	ID         uint32
	TextureID  uint32
	Coords     []TextureCoord
}

// TextureCoord is a single (u, v) entry inside a Texture2DGroup.
type TextureCoord struct {
	U, V float64
}

// CompositeMaterials defines a list of weighted blends of base materials.
// Useful for printers that mix per-voxel materials.
type CompositeMaterials struct {
	ID             uint32
	MatID          uint32   // BaseMaterials id this composite references
	MatIndices     []uint32 // indices into the referenced base material list
	Composites     []Composite
}

// Composite is one composite entry; Values has the same length as the
// surrounding CompositeMaterials.MatIndices slice.
type Composite struct {
	Values []float64
}

// MultiProperties combines multiple property layers (e.g., color + finish)
// into a single PID that triangles can reference.
type MultiProperties struct {
	ID         uint32
	PIDs       []uint32 // ordered list of property-group ids
	BlendMethods []string
	Multis     []MultiEntry
}

// MultiEntry is one row of a MultiProperties: PIndices[i] selects an index
// inside MultiProperties.PIDs[i].
type MultiEntry struct {
	PIndices []uint32
}

// Vec2 is a 2D coordinate used by Texture2D box attributes.
type Vec2 struct{ U, V float64 }
