package xmltree_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lestrrat-go/3mf/internal/xmltree"
)

func TestParseNamespacesAndText(t *testing.T) {
	const input = `<model xmlns="urn:core" xmlns:p="urn:production" xmlns:q="urn:production" ` +
		`p:UUID="123" xml:lang="en"><q:item q:ref="part"><![CDATA[A < B]]> &amp; C</q:item></model>`
	doc, err := xmltree.Parse(t.Context(), []byte(input))
	require.NoError(t, err)
	root := doc.DocumentElement()
	require.Equal(t, "model", root.LocalName())
	require.Equal(t, "urn:core", root.Namespace().URI())
	require.Len(t, root.Namespaces(), 3)
	require.Equal(t, "p", root.Namespaces()[1].Prefix())

	uuid, ok := root.FindAttribute(xmltree.NSPredicate{Local: "UUID", NamespaceURI: "urn:production"})
	require.True(t, ok)
	require.Equal(t, "p:UUID", uuid.Name())
	require.Equal(t, "123", uuid.Value())
	lang, ok := root.FindAttribute(xmltree.NSPredicate{
		Local: "lang", NamespaceURI: "http://www.w3.org/XML/1998/namespace",
	})
	require.True(t, ok)
	require.Equal(t, "en", lang.Value())

	var items []*xmltree.Element
	for child := range xmltree.Children(root) {
		if elem, ok := child.(*xmltree.Element); ok {
			items = append(items, elem)
		}
	}
	require.Len(t, items, 1)
	require.Equal(t, "urn:production", items[0].Namespace().URI())
	require.Equal(t, "q", items[0].Namespace().Prefix())
	ref, ok := items[0].FindAttribute(xmltree.NSPredicate{Local: "ref", NamespaceURI: "urn:production"})
	require.True(t, ok)
	require.Equal(t, "q:ref", ref.Name())
	var value strings.Builder
	for child := range xmltree.Children(items[0]) {
		if text, ok := child.(*xmltree.Text); ok {
			value.Write(text.Content())
		}
	}
	require.Equal(t, "A < B & C", value.String())
}

func TestParseRejectsMalformedXMLAndCancellation(t *testing.T) {
	_, err := xmltree.Parse(t.Context(), []byte(`<model><item></model>`))
	require.Error(t, err)
	_, err = xmltree.Parse(t.Context(), []byte(`<model/><other/>`))
	require.Error(t, err)
	_, err = xmltree.Parse(t.Context(), []byte(`<model><p:item/></model>`))
	require.Error(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = xmltree.Parse(ctx, []byte(`<model/>`))
	require.ErrorIs(t, err, context.Canceled)
}
