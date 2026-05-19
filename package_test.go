package tmf_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	tmf "github.com/lestrrat-go/3mf"
)

func TestRoundTripMinimalModel(t *testing.T) {
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
		tmf.WithObjectName("tet"),
		tmf.WithMesh(mesh),
	)
	item := tmf.NewBuildItem(tmf.WithObjectRef(obj))
	model := tmf.NewModel(
		tmf.WithUnit(tmf.UnitMillimeter),
		tmf.WithObject(obj),
		tmf.WithBuildItem(item),
	)
	model.AppendMetadata(&tmf.Metadata{Name: "Title", Value: "Tetrahedron"})

	pkg := tmf.NewPackage(tmf.WithModel(model))

	var buf bytes.Buffer
	_, err := pkg.WriteTo(&buf)
	require.NoError(t, err, "write package")

	// Round-trip back through ReadPackage.
	got, err := tmf.ReadPackage(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err, "read package back")

	gm := got.Model()
	require.Equal(t, tmf.UnitMillimeter, gm.Unit())
	require.Len(t, gm.Resources().Objects(), 1)
	gotObj := gm.Resources().Objects()[0]
	require.Equal(t, uint32(1), gotObj.ID())
	require.Equal(t, "tet", gotObj.Name())
	require.NotNil(t, gotObj.Mesh())
	require.Len(t, gotObj.Mesh().Vertices(), 4)
	require.Len(t, gotObj.Mesh().Triangles(), 4)
	require.Len(t, gm.Build().Items, 1)
	require.Equal(t, uint32(1), gm.Build().Items[0].ObjectID)

	md := gm.Metadata()
	require.Len(t, md, 1)
	require.Equal(t, "Title", md[0].Name)
	require.Equal(t, "Tetrahedron", md[0].Value)
}

func TestMarshalModelXMLShape(t *testing.T) {
	mesh := tmf.NewMesh(
		tmf.WithVertices([]tmf.Vertex{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}}),
		tmf.WithTriangles([]tmf.Triangle{{V1: 0, V2: 1, V3: 2}}),
	)
	obj := tmf.NewObject(tmf.WithObjectID(1), tmf.WithMesh(mesh))
	model := tmf.NewModel(
		tmf.WithUnit(tmf.UnitMillimeter),
		tmf.WithObject(obj),
		tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
	)
	data, err := tmf.MarshalModel(model)
	require.NoError(t, err)
	got := string(data)
	require.Contains(t, got, `xmlns="`+tmf.NSCore+`"`)
	require.Contains(t, got, `unit="millimeter"`)
	require.Contains(t, got, `<object id="1"`)
	require.Contains(t, got, `<triangle v1="0" v2="1" v3="2"`)
	require.Contains(t, got, `<item objectid="1"`)
}
