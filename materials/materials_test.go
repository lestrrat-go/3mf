package materials_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	tmf "github.com/lestrrat-go/3mf"
	"github.com/lestrrat-go/3mf/materials"
)

func TestColorGroupRoundTrip(t *testing.T) {
	mesh := tmf.NewMesh(
		tmf.WithVertices([]tmf.Vertex{{X: 0}, {X: 1}, {X: 0, Y: 1}}),
		tmf.WithTriangles([]tmf.Triangle{
			{V1: 0, V2: 1, V3: 2, PID: 100, P1: 0, P2: 1, P3: 0, HasPID: true, HasPIndices: true},
		}),
	)
	obj := tmf.NewObject(tmf.WithObjectID(1), tmf.WithMesh(mesh))
	model := tmf.NewModel(
		tmf.WithUnit(tmf.UnitMillimeter),
		tmf.WithObject(obj),
		tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
	)
	materials.Require(model)

	mr := materials.Of(model.Resources())
	mr.ColorGroups = append(mr.ColorGroups, &materials.ColorGroup{
		ID:     100,
		Colors: []tmf.Color{tmf.NewColor(0xff, 0, 0), tmf.NewColor(0, 0xff, 0)},
	})

	data, err := tmf.MarshalModel(model)
	require.NoError(t, err)
	require.Contains(t, string(data), `xmlns:m="`+materials.Namespace+`"`)
	require.Contains(t, string(data), `<m:colorgroup id="100"`)
	require.Contains(t, string(data), `pid="100"`)

	// Round-trip the model XML and verify the color group is recovered.
	got, err := tmf.ReadModel(t.Context(), data)
	require.NoError(t, err)
	gmr := materials.Of(got.Resources())
	require.NotNil(t, gmr)
	require.Len(t, gmr.ColorGroups, 1)
	require.Equal(t, uint32(100), gmr.ColorGroups[0].ID)
	require.Len(t, gmr.ColorGroups[0].Colors, 2)
}

func TestPackageRoundTripWithMaterials(t *testing.T) {
	mesh := tmf.NewMesh(
		tmf.WithVertices([]tmf.Vertex{{X: 0}, {X: 1}, {X: 0, Y: 1}, {X: 0, Y: 0, Z: 1}}),
		tmf.WithTriangles([]tmf.Triangle{
			{V1: 0, V2: 1, V3: 2},
			{V1: 0, V2: 1, V3: 3},
			{V1: 0, V2: 2, V3: 3},
			{V1: 1, V2: 2, V3: 3},
		}),
	)
	obj := tmf.NewObject(tmf.WithObjectID(1), tmf.WithMesh(mesh))
	model := tmf.NewModel(
		tmf.WithObject(obj),
		tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
	)
	materials.Require(model)
	mr := materials.Of(model.Resources())
	mr.ColorGroups = append(mr.ColorGroups, &materials.ColorGroup{
		ID:     5,
		Colors: []tmf.Color{tmf.NewColor(255, 128, 64)},
	})

	pkg := tmf.NewPackage(tmf.WithModel(model))
	var buf bytes.Buffer
	_, err := pkg.WriteTo(&buf)
	require.NoError(t, err)

	got, err := tmf.ReadPackage(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	gmr := materials.Of(got.Model().Resources())
	require.NotNil(t, gmr)
	require.Len(t, gmr.ColorGroups, 1)
	require.Equal(t, uint32(5), gmr.ColorGroups[0].ID)
}
