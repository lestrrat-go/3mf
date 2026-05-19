package opc

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/lestrrat-go/helium"
	"github.com/lestrrat-go/helium/stream"
)

// RelationshipNS is the XML namespace for OPC relationships parts.
const RelationshipNS = "http://schemas.openxmlformats.org/package/2006/relationships"

// TargetMode controls whether a Relationship's Target is resolved within the
// package (Internal) or as an arbitrary URL (External).
type TargetMode int

const (
	TargetModeInternal TargetMode = iota
	TargetModeExternal
)

func (m TargetMode) String() string {
	if m == TargetModeExternal {
		return "External"
	}
	return "Internal"
}

// Relationship is a single <Relationship> entry inside an OPC .rels part.
type Relationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode TargetMode
}

// Relationships holds all relationships originating from a single source
// part. The empty source part name "/" denotes the package itself, whose
// relationships live in /_rels/.rels.
type Relationships struct {
	// Source is the part that owns these relationships. Use "/" for the
	// package itself.
	Source string

	// Entries is the ordered list of <Relationship> elements.
	Entries []Relationship
}

// NewRelationships returns an empty Relationships rooted at source.
func NewRelationships(source string) *Relationships {
	return &Relationships{Source: NormalizePartName(source)}
}

// Add appends a new internal relationship to entries.
func (r *Relationships) Add(id, relType, target string) {
	r.Entries = append(r.Entries, Relationship{
		ID:     id,
		Type:   relType,
		Target: target,
	})
}

// AddExternal appends a new external relationship.
func (r *Relationships) AddExternal(id, relType, target string) {
	r.Entries = append(r.Entries, Relationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: TargetModeExternal,
	})
}

// FindByType returns the first relationship whose Type equals relType, or
// nil if none matches.
func (r *Relationships) FindByType(relType string) *Relationship {
	for i := range r.Entries {
		if r.Entries[i].Type == relType {
			return &r.Entries[i]
		}
	}
	return nil
}

// FindAllByType returns every relationship whose Type equals relType.
func (r *Relationships) FindAllByType(relType string) []*Relationship {
	var out []*Relationship
	for i := range r.Entries {
		if r.Entries[i].Type == relType {
			out = append(out, &r.Entries[i])
		}
	}
	return out
}

// ResolveTarget returns the absolute part name that entry.Target resolves to
// when interpreted relative to r.Source. For external relationships the raw
// Target is returned unchanged.
func (r *Relationships) ResolveTarget(entry *Relationship) string {
	if entry.TargetMode == TargetModeExternal {
		return entry.Target
	}
	return ResolveRelative(r.Source, entry.Target)
}

// ParseRelationships reads a .rels payload originating from source.
func ParseRelationships(source string, data []byte) (*Relationships, error) {
	doc, err := helium.NewParser().Parse(context.Background(), data)
	if err != nil {
		return nil, fmt.Errorf("opc: parse relationships: %w", err)
	}
	root := doc.DocumentElement()
	if root == nil || root.LocalName() != "Relationships" {
		return nil, fmt.Errorf("opc: relationships root must be <Relationships>")
	}
	out := NewRelationships(source)
	for child := range helium.Children(root) {
		elem, ok := child.(*helium.Element)
		if !ok || elem.LocalName() != "Relationship" {
			continue
		}
		mode := TargetModeInternal
		if attr(elem, "TargetMode") == "External" {
			mode = TargetModeExternal
		}
		out.Entries = append(out.Entries, Relationship{
			ID:         attr(elem, "Id"),
			Type:       attr(elem, "Type"),
			Target:     attr(elem, "Target"),
			TargetMode: mode,
		})
	}
	return out, nil
}

// WriteTo serializes r as a .rels payload.
func (r *Relationships) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}
	sw := stream.NewWriter(cw)
	if err := sw.StartDocument("1.0", "UTF-8", "yes"); err != nil {
		return cw.n, err
	}
	if err := sw.StartElement("Relationships"); err != nil {
		return cw.n, err
	}
	if err := sw.WriteAttribute("xmlns", RelationshipNS); err != nil {
		return cw.n, err
	}
	for _, e := range r.Entries {
		if err := sw.StartElement("Relationship"); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("Id", e.ID); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("Type", e.Type); err != nil {
			return cw.n, err
		}
		if err := sw.WriteAttribute("Target", e.Target); err != nil {
			return cw.n, err
		}
		if e.TargetMode == TargetModeExternal {
			if err := sw.WriteAttribute("TargetMode", "External"); err != nil {
				return cw.n, err
			}
		}
		if err := sw.EndElement(); err != nil {
			return cw.n, err
		}
	}
	if err := sw.EndElement(); err != nil {
		return cw.n, err
	}
	if err := sw.EndDocument(); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// Bytes returns the serialized .rels payload.
func (r *Relationships) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := r.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
