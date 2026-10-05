package xmltree_test

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lestrrat-go/3mf/internal/xmltree"
)

func TestWriterProducesNamespacedEscapedXML(t *testing.T) {
	var output bytes.Buffer
	w := xmltree.NewWriter(&output)
	require.NoError(t, w.StartDocument("1.0", "UTF-8", "yes"))
	require.NoError(t, w.StartElement("model"))
	require.NoError(t, w.WriteAttribute("xmlns", "urn:core"))
	require.NoError(t, w.WriteAttribute("xmlns:p", "urn:production"))
	require.NoError(t, w.StartElement("p:item"))
	require.NoError(t, w.WriteAttribute("p:UUID", `A " B`))
	require.NoError(t, w.WriteString("A < B & C"))
	require.NoError(t, w.EndElement())
	require.NoError(t, w.EndElement())
	require.NoError(t, w.EndDocument())
	require.Contains(t, output.String(), `p:UUID="A &#34; B"`)
	require.Contains(t, output.String(), `A &lt; B &amp; C`)

	var decoded struct {
		XMLName xml.Name `xml:"urn:core model"`
		Item    string   `xml:"urn:production item"`
	}
	require.NoError(t, xml.Unmarshal(output.Bytes(), &decoded))
	require.Equal(t, "A < B & C", decoded.Item)
}

func TestWriterRequiresOpenStartForAttributes(t *testing.T) {
	var output bytes.Buffer
	w := xmltree.NewWriter(&output)
	require.Error(t, w.WriteAttribute("name", "value"))
	require.NoError(t, w.StartElement("item"))
	require.NoError(t, w.WriteString("text"))
	require.Error(t, w.WriteAttribute("name", "value"))
	require.NoError(t, w.EndElement())
	require.NoError(t, w.EndDocument())
}
