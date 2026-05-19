# 3mf

Parse and construct [3MF](https://3mf.io) (3D Manufacturing Format) files in Go.

This module implements the 3MF Core Specification plus the major 3MF
extensions (Materials, Production, Beam Lattice, Slice, Secure Content,
Volumetric). XML I/O is built on
[github.com/lestrrat-go/helium](https://github.com/lestrrat-go/helium); the
public API uses functional options via
[github.com/lestrrat-go/option/v3](https://github.com/lestrrat-go/option).

## Status

Early. The package builds and round-trips the core spec plus every extension
listed above through write → ZIP → re-read. Many sharper edges of the spec
(secure-content key wrap policies, exotic volumetric node taxonomies, etc.)
are surfaced as escape hatches rather than fully typed APIs; see the
per-package docs.

## Install

```text
go get github.com/lestrrat-go/3mf
```

## Read a package

```go
pkg, err := tmf.Open("model.3mf")
if err != nil {
    log.Fatal(err)
}
model := pkg.Model()
fmt.Println("unit:", model.Unit())
for _, obj := range model.Resources().Objects() {
    fmt.Printf("object %d: %d vertices, %d triangles\n",
        obj.ID(),
        len(obj.Mesh().Vertices()),
        len(obj.Mesh().Triangles()))
}
```

## Construct a package

```go
mesh := tmf.NewMesh(
    tmf.WithVertices([]tmf.Vertex{
        {X: 0, Y: 0, Z: 0},
        {X: 10, Y: 0, Z: 0},
        {X: 0, Y: 10, Z: 0},
        {X: 0, Y: 0, Z: 10},
    }),
    tmf.WithTriangles([]tmf.Triangle{
        {V1: 0, V2: 1, V3: 2},
        {V1: 0, V2: 1, V3: 3},
        {V1: 0, V2: 2, V3: 3},
        {V1: 1, V2: 2, V3: 3},
    }),
)
obj := tmf.NewObject(
    tmf.WithObjectID(1),
    tmf.WithObjectName("tetrahedron"),
    tmf.WithMesh(mesh),
)
model := tmf.NewModel(
    tmf.WithUnit(tmf.UnitMillimeter),
    tmf.WithObject(obj),
    tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
)
pkg := tmf.NewPackage(tmf.WithModel(model))
if err := pkg.Save("out.3mf"); err != nil {
    log.Fatal(err)
}
```

## Extensions

Extensions live in sub-packages and register themselves on import. Blank-
import the ones you need:

```go
import (
    _ "github.com/lestrrat-go/3mf/production"
    _ "github.com/lestrrat-go/3mf/materials"
    _ "github.com/lestrrat-go/3mf/beamlattice"
    _ "github.com/lestrrat-go/3mf/slice"
    _ "github.com/lestrrat-go/3mf/securecontent"
    _ "github.com/lestrrat-go/3mf/volumetric"
)
```

| Sub-package    | Namespace prefix | Notes |
|----------------|------------------|-------|
| `production`   | `p`              | Per-element UUIDs, cross-part references |
| `materials`    | `m`              | Color groups, textures, composites, multi-properties |
| `beamlattice`  | `b`              | Beam lattices and ball sets on meshes |
| `slice`        | `s`              | Slice stacks for printer-ready 2D contours |
| `securecontent`| `sc`             | Encryption metadata + AES-GCM helpers; user supplies key resolver |
| `volumetric`   | `v`              | Implicit / volumetric functions (preserved as opaque trees) |

To mark an extension as required by your model so its namespace appears in
`requiredextensions` and a prefix is declared on the root `<model>`:

```go
import "github.com/lestrrat-go/3mf/production"

production.Require(model)
model.Resources().Objects()[0].SetUUID("00000000-0000-0000-0000-000000000001")
```

## License

MIT — see `LICENSE`.
