// Package tmf implements parsing and construction of 3MF (3D Manufacturing
// Format) files.
//
// 3MF is an XML-based packaging format for additive manufacturing data,
// distributed as an OPC (Open Packaging Conventions) ZIP container. This
// package implements the 3MF core specification together with the major
// 3MF extensions (Materials, Production, Beam Lattice, Slice, Secure Content,
// and Volumetric Extensions).
//
// # Reading
//
//	pkg, err := tmf.Open("model.3mf")
//	if err != nil { ... }
//	defer pkg.Close()
//	model := pkg.Model()
//	for _, obj := range model.Resources().Objects() {
//	    // ...
//	}
//
// # Writing
//
//	mesh := tmf.NewMesh(
//	    tmf.WithVertices(verts),
//	    tmf.WithTriangles(tris),
//	)
//	obj := tmf.NewObject(
//	    tmf.WithObjectID(1),
//	    tmf.WithObjectType(tmf.ObjectTypeModel),
//	    tmf.WithMesh(mesh),
//	)
//	model := tmf.NewModel(
//	    tmf.WithUnit(tmf.UnitMillimeter),
//	    tmf.WithObjects(obj),
//	    tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
//	)
//	pkg := tmf.NewPackage(tmf.WithModel(model))
//	if err := pkg.Save("output.3mf"); err != nil { ... }
//
// # Extensions
//
// Extension data is exposed both through the typed sub-packages
// (github.com/lestrrat-go/3mf/materials, .../production, .../beamlattice,
// .../slice, .../securecontent, .../volumetric) and through a generic
// attachment mechanism on Object, Mesh, Build, and Model values so that
// unrecognized extension data is preserved on round-trip.
package tmf
