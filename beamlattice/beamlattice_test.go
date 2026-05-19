package beamlattice_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmf "github.com/lestrrat-go/3mf"
	"github.com/lestrrat-go/3mf/beamlattice"
)

func TestBeamLatticeRoundTrip(t *testing.T) {
	mesh := tmf.NewMesh(
		tmf.WithVertices([]tmf.Vertex{{X: 0}, {X: 1}, {X: 0, Y: 1}}),
	)
	mesh.SetExtension(beamlattice.Namespace, &beamlattice.BeamLattice{
		Radius:    0.5,
		MinLength: 0.01,
		Cap:       beamlattice.CapModeButt,
		Beams: []beamlattice.Beam{
			{V1: 0, V2: 1},
			{V1: 1, V2: 2, R1: 0.3, R2: 0.4, Cap1: beamlattice.CapModeHemisphere, HasCap1: true},
		},
		BallSets: []beamlattice.BallSet{
			{Identifier: "joints", Balls: []beamlattice.Ball{{V: 0, R: 0.6}}},
		},
	})
	obj := tmf.NewObject(tmf.WithObjectID(1), tmf.WithMesh(mesh))
	model := tmf.NewModel(
		tmf.WithObject(obj),
		tmf.WithBuildItem(tmf.NewBuildItem(tmf.WithObjectRef(obj))),
	)
	beamlattice.Require(model)

	data, err := tmf.MarshalModel(model)
	require.NoError(t, err)
	require.Contains(t, string(data), `xmlns:b="`+beamlattice.Namespace+`"`)
	require.Contains(t, string(data), `<b:beamlattice`)

	got, err := tmf.ReadModel(t.Context(), data)
	require.NoError(t, err)
	gObj := got.Resources().Objects()[0]
	bl := beamlattice.Of(gObj.Mesh())
	require.NotNil(t, bl)
	require.Equal(t, 0.5, bl.Radius)
	require.Len(t, bl.Beams, 2)
	require.Equal(t, beamlattice.CapModeButt, bl.Cap)
	require.Len(t, bl.BallSets, 1)
	require.Equal(t, "joints", bl.BallSets[0].Identifier)
}
