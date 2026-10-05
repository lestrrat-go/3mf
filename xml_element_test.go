package tmf_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmf "github.com/lestrrat-go/3mf"
)

const extensionElementTestNamespace = "urn:3mf:element-api-test"

type elementCaptureReader struct {
	tmf.BaseExtensionReader

	element *tmf.Element
}

func (*elementCaptureReader) Namespace() string { return extensionElementTestNamespace }

func (r *elementCaptureReader) ReadResourceElement(_ *tmf.Resources, elem *tmf.Element) error {
	r.element = elem
	return nil
}

func TestExtensionElementAPI(t *testing.T) {
	reader := &elementCaptureReader{}
	tmf.RegisterExtensionReader(reader)
	const modelXML = `<model xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02" ` +
		`xmlns:x="urn:3mf:element-api-test"><resources><x:custom x:id="42">` +
		`<x:note><![CDATA[A < B]]></x:note></x:custom></resources><build/></model>`
	_, err := tmf.ReadModel(t.Context(), []byte(modelXML))
	require.NoError(t, err)
	require.NotNil(t, reader.element)
	require.Equal(t, "custom", reader.element.LocalName())
	require.Equal(t, extensionElementTestNamespace, reader.element.NamespaceURI())
	require.Equal(t, "42", reader.element.AttrNS("id", extensionElementTestNamespace))
	require.Equal(t, "x:id", reader.element.Attributes()[0].Name)

	var notes []*tmf.Element
	for child := range reader.element.ChildElements("note") {
		notes = append(notes, child)
	}
	require.Len(t, notes, 1)
	require.Equal(t, "A < B", notes[0].TextContent())
}
