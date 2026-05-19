package production_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmf "github.com/lestrrat-go/3mf"
	"github.com/lestrrat-go/3mf/production"
)

func TestProductionUUIDRoundTrip(t *testing.T) {
	obj := tmf.NewObject(
		tmf.WithObjectID(1),
		tmf.WithObjectUUID("00000000-0000-0000-0000-000000000001"),
		tmf.WithMesh(tmf.NewMesh(
			tmf.WithVertices([]tmf.Vertex{{X: 0}, {X: 1}, {X: 0, Y: 1}}),
			tmf.WithTriangles([]tmf.Triangle{{V1: 0, V2: 1, V3: 2}}),
		)),
	)
	item := tmf.NewBuildItem(
		tmf.WithObjectRef(obj),
		tmf.WithItemUUID("00000000-0000-0000-0000-000000000010"),
	)
	model := tmf.NewModel(tmf.WithObject(obj), tmf.WithBuildItem(item))
	production.Require(model)
	model.Build().UUID = "00000000-0000-0000-0000-000000000020"

	data, err := tmf.MarshalModel(model)
	require.NoError(t, err)
	s := string(data)
	require.Contains(t, s, `xmlns:p="`+production.Namespace+`"`)
	require.Contains(t, s, `p:UUID="00000000-0000-0000-0000-000000000001"`)
	require.Contains(t, s, `p:UUID="00000000-0000-0000-0000-000000000010"`)
	require.Contains(t, s, `p:UUID="00000000-0000-0000-0000-000000000020"`)

	got, err := tmf.ReadModel(t.Context(), data)
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", got.Resources().Objects()[0].UUID())
	require.Equal(t, "00000000-0000-0000-0000-000000000010", got.Build().Items[0].UUID)
	require.Equal(t, "00000000-0000-0000-0000-000000000020", got.Build().UUID)
}
