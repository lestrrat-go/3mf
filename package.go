package tmf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/lestrrat-go/option/v3"

	"github.com/lestrrat-go/3mf/opc"
)

// Package is a 3MF package: an OPC container plus a primary in-memory Model.
// When the Production extension is in use, additional model parts are kept
// around in OPC.Parts so that AdditionalModel(path) can decode them on
// demand and writes can faithfully round-trip the package.
type Package struct {
	opc       *opc.Package
	model     *Model
	modelPath string

	// modelCache caches Model objects parsed from non-primary parts so that
	// repeated AdditionalModel calls are cheap.
	modelCache map[string]*Model
}

// NewPackage constructs an empty Package with the given primary Model and
// optional package-level metadata (thumbnail, attachments, etc.).
func NewPackage(opts ...Option) *Package {
	p := &Package{
		opc:        opc.NewPackage(),
		modelPath:  DefaultModelPath,
		modelCache: map[string]*Model{},
	}
	for _, o := range opts {
		switch o.Ident() {
		case identPackageModel{}:
			p.model = option.MustGet[*Model](o)
		case identPackageModelPath{}:
			p.modelPath = option.MustGet[string](o)
		case identPackageThumbnail{}:
			t := option.MustGet[Thumbnail](o)
			p.opc.AddPart(t.Path, t.ContentType, t.Data)
			rels := p.opc.PackageRelationships()
			rels.Add(nextRelID(rels), RelTypeThumbnail, t.Path)
		case identPackageAttachment{}:
			a := option.MustGet[Attachment](o)
			p.opc.AddPart(a.Path, a.ContentType, a.Data)
			if a.RelType != "" {
				rels := p.opc.PackageRelationships()
				rels.Add(nextRelID(rels), a.RelType, a.Path)
			}
		}
	}
	if p.model == nil {
		p.model = NewModel()
	}
	return p
}

// identifiers for package-level options.
type (
	identPackageModel      struct{}
	identPackageModelPath  struct{}
	identPackageThumbnail  struct{}
	identPackageAttachment struct{}
)

// Thumbnail attaches a package-level thumbnail image (referenced from the
// package's _rels/.rels).
type Thumbnail struct {
	Path        string // typically "/Metadata/thumbnail.png"
	ContentType string // typically ContentTypePNG
	Data        []byte
}

// Attachment is an extra OPC part written into the package, optionally with
// a package-level relationship of the given type.
type Attachment struct {
	Path        string
	ContentType string
	RelType     string // when non-empty, a package relationship is created
	Data        []byte
}

// WithModel attaches a primary Model to a new Package.
func WithModel(m *Model) Option { return option.New(identPackageModel{}, m) }

// WithModelPath overrides the in-package path of the primary 3MF model part.
// Defaults to DefaultModelPath ("/3D/3dmodel.model").
func WithModelPath(p string) Option { return option.New(identPackageModelPath{}, p) }

// WithPackageThumbnail attaches a package-level thumbnail.
func WithPackageThumbnail(t Thumbnail) Option { return option.New(identPackageThumbnail{}, t) }

// WithAttachment attaches an extra OPC part.
func WithAttachment(a Attachment) Option { return option.New(identPackageAttachment{}, a) }

// Model returns the primary Model of the package (never nil).
func (p *Package) Model() *Model { return p.model }

// SetModel replaces the primary Model.
func (p *Package) SetModel(m *Model) { p.model = m }

// ModelPath returns the in-package path of the primary 3MF model part.
func (p *Package) ModelPath() string { return p.modelPath }

// OPC returns the underlying OPC package. Modifying it directly is fine but
// callers should ensure consistency with the primary Model.
func (p *Package) OPC() *opc.Package { return p.opc }

// Part returns the raw bytes of the part at the given absolute name, or nil
// when no such part exists.
func (p *Package) Part(name string) []byte {
	return p.opc.Parts[opc.NormalizePartName(name)]
}

// AddPart writes raw bytes into the package at the given absolute part name,
// with the given content type.
func (p *Package) AddPart(name, contentType string, data []byte) {
	p.opc.AddPart(name, contentType, data)
}

// AddAttachment is a convenience that registers an attachment plus an
// optional package-level relationship.
func (p *Package) AddAttachment(a Attachment) {
	p.opc.AddPart(a.Path, a.ContentType, a.Data)
	if a.RelType != "" {
		rels := p.opc.PackageRelationships()
		rels.Add(nextRelID(rels), a.RelType, a.Path)
	}
}

// AdditionalModel decodes the secondary 3MF model part at the given absolute
// part name (Production extension). Results are cached.
func (p *Package) AdditionalModel(name string) (*Model, error) {
	abs := opc.NormalizePartName(name)
	if m, ok := p.modelCache[abs]; ok {
		return m, nil
	}
	data := p.Part(abs)
	if data == nil {
		return nil, fmt.Errorf("tmf: no part %s", abs)
	}
	m, err := ReadModel(context.Background(), data)
	if err != nil {
		return nil, err
	}
	p.modelCache[abs] = m
	return m, nil
}

// Open reads a 3MF package from the file at path.
func Open(filePath string) (*Package, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return ReadPackage(bytes.NewReader(data), int64(len(data)))
}

// ReadPackage parses a 3MF package from r (which must implement io.ReaderAt
// because the ZIP format requires random access). size is the total size of
// the package in bytes.
func ReadPackage(r io.ReaderAt, size int64) (*Package, error) {
	op, err := opc.ReadFrom(r, size)
	if err != nil {
		return nil, err
	}
	return packageFromOPC(op)
}

func packageFromOPC(op *opc.Package) (*Package, error) {
	rels := op.PackageRelationships()
	startRel := rels.FindByType(RelTypeStartPart)
	if startRel == nil {
		return nil, ErrNoModelPart
	}
	modelPath := rels.ResolveTarget(startRel)
	data := op.Parts[modelPath]
	if data == nil {
		return nil, fmt.Errorf("%w: start part %s not found in package", ErrInvalidPackage, modelPath)
	}
	m, err := ReadModel(context.Background(), data)
	if err != nil {
		return nil, err
	}
	return &Package{
		opc:        op,
		model:      m,
		modelPath:  modelPath,
		modelCache: map[string]*Model{modelPath: m},
	}, nil
}

// Save writes p to the file at filePath, creating the file if necessary.
func (p *Package) Save(filePath string) error {
	f, err := os.Create(filePath) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = p.WriteTo(f)
	return err
}

// WriteTo serializes the package as a ZIP archive to w.
func (p *Package) WriteTo(w io.Writer) (int64, error) {
	// Serialize the primary model and stash it in the OPC parts before
	// emitting the ZIP. Also ensure the [Content_Types] entry and the
	// start-part relationship are in place.
	if err := p.syncOPC(); err != nil {
		return 0, err
	}
	return p.opc.WriteTo(w)
}

// syncOPC re-serializes the primary Model into the OPC parts table and
// ensures [Content_Types].xml and the package relationship to the start
// part are present.
func (p *Package) syncOPC() error {
	if p.model == nil {
		return fmt.Errorf("tmf: package has no model")
	}
	data, err := MarshalModel(p.model)
	if err != nil {
		return err
	}
	p.opc.Parts[opc.NormalizePartName(p.modelPath)] = data

	// Register the model part's content type.
	p.opc.ContentTypes.AddOverride(p.modelPath, ContentType3DModel)
	p.opc.ContentTypes.AddDefault("rels", ContentTypeRelationships)

	// Ensure the start-part relationship exists.
	rels := p.opc.PackageRelationships()
	if rels.FindByType(RelTypeStartPart) == nil {
		rels.Add(nextRelID(rels), RelTypeStartPart, p.modelPath)
	}

	// Pull in textures referenced from materials extension and other
	// per-part rels remain as-is — they were carried in from the OPC
	// reader (round-trip) or added by the caller via AddPart.
	_ = path.Ext // keep the import alive when only used in test-only code
	return nil
}

// nextRelID returns a fresh "rIdN" identifier that doesn't collide with the
// existing entries in r.
func nextRelID(r *opc.Relationships) string {
	highest := 0
	for _, e := range r.Entries {
		if len(e.ID) > 3 && e.ID[:3] == "rId" {
			n := 0
			for _, c := range e.ID[3:] {
				if c < '0' || c > '9' {
					n = 0
					break
				}
				n = n*10 + int(c-'0')
			}
			if n > highest {
				highest = n
			}
		}
	}
	return fmt.Sprintf("rId%d", highest+1)
}
