package opc

import (
	"path"
	"strings"
)

// NormalizePartName returns the canonical part-name form expected by OPC:
// rooted (begins with "/"), forward slashes, and percent-decoded. OPC part
// names look like "/3D/3dmodel.model" — a leading slash, no trailing slash
// (unless the part is the root itself), and no "." or ".." segments.
func NormalizePartName(p string) string {
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	cleaned := path.Clean(p)
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

// ResolveRelative joins a relationship target (which may be relative) to the
// base part name and returns the absolute part name. base is expected to be
// the part that owns the relationship; target is taken verbatim from the
// Target attribute of <Relationship>.
func ResolveRelative(base, target string) string {
	if strings.HasPrefix(target, "/") {
		return NormalizePartName(target)
	}
	dir := path.Dir(NormalizePartName(base))
	return NormalizePartName(path.Join(dir, target))
}

// RelsName returns the path of the .rels part that holds the relationships
// for the given part. For "/3D/3dmodel.model" it returns "/3D/_rels/3dmodel.model.rels".
// For the package itself ("/") it returns "/_rels/.rels".
func RelsName(part string) string {
	p := NormalizePartName(part)
	if p == "/" {
		return "/_rels/.rels"
	}
	dir, base := path.Split(p)
	return dir + "_rels/" + base + ".rels"
}

// IsRelsPart reports whether name is the path of an OPC relationships part.
func IsRelsPart(name string) bool {
	p := NormalizePartName(name)
	return p == "/_rels/.rels" || strings.HasSuffix(p, ".rels") && strings.Contains(p, "/_rels/")
}

// IsContentTypesPart reports whether name is the [Content_Types].xml part.
func IsContentTypesPart(name string) bool {
	return NormalizePartName(name) == "/[Content_Types].xml"
}
