package tmf

import (
	"maps"
	"math"

	"github.com/lestrrat-go/option/v3"
)

// Vertex is a 3D coordinate in the model's local unit system.
type Vertex struct {
	X, Y, Z float64
}

// Triangle is a face of a mesh referencing three Vertices by index. The
// per-vertex property indices PIndices and the per-triangle property
// identifier PID encode 3MF material/property references; both default to
// zero, meaning "inherit from the object".
type Triangle struct {
	V1, V2, V3 uint32

	// PID is the resource id of a property group (BaseMaterials, ColorGroup,
	// etc.) that this triangle references. A value of zero means no override.
	PID uint32

	// P1, P2, P3 are per-vertex indices into the property group named by PID.
	// When PID is zero these are ignored.
	P1, P2, P3 uint32

	// HasPID is true when an explicit PID was set on this triangle (so a
	// zero PID can still be distinguished from the default).
	HasPID bool

	// HasPIndices is true when the per-vertex P1/P2/P3 indices were set.
	// Some triangles set only PID, in which case all three vertices share
	// the same P1 value.
	HasPIndices bool
}

// Matrix is a 4x3 affine transform stored in row-major order matching the
// 3MF wire format ("m00 m01 m02 m10 m11 m12 m20 m21 m22 m30 m31 m32"). The
// implicit fourth column is (0 0 0 1)^T.
type Matrix [12]float64

// IdentityMatrix returns the identity transform.
func IdentityMatrix() Matrix {
	return Matrix{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
}

// IsIdentity reports whether m equals the identity transform within a
// tight tolerance.
func (m Matrix) IsIdentity() bool {
	id := IdentityMatrix()
	const eps = 1e-12
	for i := range m {
		if math.Abs(m[i]-id[i]) > eps {
			return false
		}
	}
	return true
}

// Mesh is the geometry attached to an Object.
type Mesh struct {
	vertices  []Vertex
	triangles []Triangle

	// extension payloads keyed by namespace URI
	extensions map[string]any
}

// NewMesh constructs a Mesh from functional options.
func NewMesh(opts ...Option) *Mesh {
	m := &Mesh{}
	for _, o := range opts {
		switch o.Ident() {
		case identVertices{}:
			m.vertices = append(m.vertices, option.MustGet[[]Vertex](o)...)
		case identTriangles{}:
			m.triangles = append(m.triangles, option.MustGet[[]Triangle](o)...)
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

// Vertices returns the mesh vertices. The returned slice is owned by the mesh
// and must not be mutated by the caller.
func (m *Mesh) Vertices() []Vertex { return m.vertices }

// Triangles returns the mesh triangles. The returned slice is owned by the
// mesh and must not be mutated by the caller.
func (m *Mesh) Triangles() []Triangle { return m.triangles }

// SetVertices replaces the vertex list. The provided slice is retained
// without copying.
func (m *Mesh) SetVertices(v []Vertex) { m.vertices = v }

// SetTriangles replaces the triangle list. The provided slice is retained
// without copying.
func (m *Mesh) SetTriangles(t []Triangle) { m.triangles = t }

// AppendVertex appends a single vertex and returns its zero-based index.
func (m *Mesh) AppendVertex(v Vertex) uint32 {
	m.vertices = append(m.vertices, v)
	return uint32(len(m.vertices) - 1)
}

// AppendTriangle appends a single triangle.
func (m *Mesh) AppendTriangle(t Triangle) { m.triangles = append(m.triangles, t) }

// Extension returns the parsed extension payload that was attached to the mesh
// for namespace ns, or nil if none was attached. The returned value is the
// extension-specific Go type — for instance, a *beamlattice.BeamLattice when
// ns is beamlattice.Namespace.
func (m *Mesh) Extension(ns string) any {
	if m.extensions == nil {
		return nil
	}
	return m.extensions[ns]
}

// SetExtension attaches an extension payload for namespace ns to the mesh.
// Passing nil removes any existing payload for that namespace.
func (m *Mesh) SetExtension(ns string, v any) {
	if v == nil {
		delete(m.extensions, ns)
		return
	}
	if m.extensions == nil {
		m.extensions = make(map[string]any)
	}
	m.extensions[ns] = v
}

// Extensions returns a copy of the extension namespaces present on the mesh.
func (m *Mesh) Extensions() map[string]any {
	if len(m.extensions) == 0 {
		return nil
	}
	out := make(map[string]any, len(m.extensions))
	maps.Copy(out, m.extensions)
	return out
}

// WithVertices supplies a slice of vertices to NewMesh.
func WithVertices(v []Vertex) Option { return option.New(identVertices{}, v) }

// WithTriangles supplies a slice of triangles to NewMesh.
func WithTriangles(t []Triangle) Option { return option.New(identTriangles{}, t) }
