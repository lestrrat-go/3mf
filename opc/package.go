package opc

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

// Package is an in-memory view of an OPC ZIP package. Parts are stored as
// byte slices, alongside a single ContentTypes value and a set of
// Relationships keyed by source part name. The package itself owns the
// "/_rels/.rels" relationship file (keyed under "/" in Relationships).
type Package struct {
	ContentTypes  *ContentTypes
	Relationships map[string]*Relationships
	Parts         map[string][]byte
}

// NewPackage returns an empty Package with all maps and the ContentTypes
// initialized.
func NewPackage() *Package {
	return &Package{
		ContentTypes:  NewContentTypes(),
		Relationships: map[string]*Relationships{},
		Parts:         map[string][]byte{},
	}
}

// AddPart registers a part with its content type. partName is normalized
// to an absolute OPC name. The content type is recorded as an Override
// unless one of the package's Default entries already covers the part's
// file extension.
func (p *Package) AddPart(partName, contentType string, data []byte) {
	name := NormalizePartName(partName)
	p.Parts[name] = data
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(name)), ".")
	if ext != "" && p.ContentTypes.Defaults[ext] == contentType {
		return
	}
	if ext != "" && p.ContentTypes.Defaults[ext] == "" {
		p.ContentTypes.AddDefault(ext, contentType)
		return
	}
	p.ContentTypes.AddOverride(name, contentType)
}

// AddRelationship appends a relationship rooted at source. Source "/" is
// the package-level relationship store. The relationship's Target is taken
// verbatim from the caller (callers typically pass a relative path).
func (p *Package) AddRelationship(source, id, relType, target string) {
	src := NormalizePartName(source)
	rels := p.Relationships[src]
	if rels == nil {
		rels = NewRelationships(src)
		p.Relationships[src] = rels
	}
	rels.Add(id, relType, target)
}

// PackageRelationships returns the relationships rooted at "/", creating
// an empty container if none yet exists.
func (p *Package) PackageRelationships() *Relationships {
	rels := p.Relationships["/"]
	if rels == nil {
		rels = NewRelationships("/")
		p.Relationships["/"] = rels
	}
	return rels
}

// PartRelationships returns the relationships rooted at partName, or nil
// if none exist.
func (p *Package) PartRelationships(partName string) *Relationships {
	return p.Relationships[NormalizePartName(partName)]
}

// ReadFrom parses an OPC package from r (which must implement io.ReaderAt
// because archive/zip requires random access). size is the total size of
// the archive in bytes.
func ReadFrom(r io.ReaderAt, size int64) (*Package, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("opc: open zip: %w", err)
	}
	return readZip(zr)
}

// ReadBytes parses an OPC package from a byte slice.
func ReadBytes(data []byte) (*Package, error) {
	return ReadFrom(bytes.NewReader(data), int64(len(data)))
}

func readZip(zr *zip.Reader) (*Package, error) {
	pkg := NewPackage()
	var ctData []byte
	relsParts := map[string][]byte{}
	for _, f := range zr.File {
		name := NormalizePartName(f.Name)
		data, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("opc: read %s: %w", f.Name, err)
		}
		switch {
		case IsContentTypesPart(name):
			ctData = data
		case IsRelsPart(name):
			relsParts[name] = data
		default:
			pkg.Parts[name] = data
		}
	}
	if ctData == nil {
		return nil, errors.New("opc: missing [Content_Types].xml")
	}
	ct, err := ParseContentTypes(ctData)
	if err != nil {
		return nil, err
	}
	pkg.ContentTypes = ct
	for relsName, data := range relsParts {
		src := sourceForRels(relsName)
		rels, err := ParseRelationships(src, data)
		if err != nil {
			return nil, fmt.Errorf("opc: parse %s: %w", relsName, err)
		}
		pkg.Relationships[src] = rels
	}
	return pkg, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// sourceForRels returns the part name that owns the relationship part at
// relsName. For "/_rels/.rels" it returns "/" (the package). For
// "/3D/_rels/3dmodel.model.rels" it returns "/3D/3dmodel.model".
func sourceForRels(relsName string) string {
	p := NormalizePartName(relsName)
	if p == rootRelsName {
		return "/"
	}
	dir, base := path.Split(p)
	dir = strings.TrimSuffix(dir, "_rels/")
	base = strings.TrimSuffix(base, ".rels")
	return NormalizePartName(dir + base)
}

// WriteTo serializes pkg as a ZIP archive to w.
func (p *Package) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}
	zw := zip.NewWriter(cw)
	// Write [Content_Types].xml first by convention (some viewers expect it).
	ctBytes, err := p.ContentTypes.Bytes()
	if err != nil {
		return cw.n, err
	}
	if err := writeZipEntry(zw, "[Content_Types].xml", ctBytes); err != nil {
		return cw.n, err
	}
	// Then relationship parts, in deterministic order.
	relsNames := make([]string, 0, len(p.Relationships))
	for k := range p.Relationships {
		relsNames = append(relsNames, k)
	}
	sortStrings(relsNames)
	for _, src := range relsNames {
		rels := p.Relationships[src]
		if rels == nil || len(rels.Entries) == 0 {
			continue
		}
		data, err := rels.Bytes()
		if err != nil {
			return cw.n, err
		}
		name := strings.TrimPrefix(RelsName(src), "/")
		if err := writeZipEntry(zw, name, data); err != nil {
			return cw.n, err
		}
	}
	// Then ordinary parts in deterministic order.
	partNames := make([]string, 0, len(p.Parts))
	for k := range p.Parts {
		partNames = append(partNames, k)
	}
	sortStrings(partNames)
	for _, name := range partNames {
		stripped := strings.TrimPrefix(name, "/")
		if err := writeZipEntry(zw, stripped, p.Parts[name]); err != nil {
			return cw.n, err
		}
	}
	if err := zw.Close(); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

func writeZipEntry(zw *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	fw, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

func sortStrings(s []string) {
	// Small in-place insertion sort to avoid importing "sort" twice and to
	// keep the dependency surface minimal.
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
