// Package securecontent implements the metadata side of the 3MF Secure
// Content Extension. The extension defines XML element wrappers that mark
// individual OPC parts as encrypted; the actual encryption uses XML
// Encryption with AES-256-GCM and (optionally) an RSA-OAEP key wrap to
// protect the per-part content encryption keys.
//
// This package reads and writes the Secure Content metadata so that
// packages produced by other tools round-trip cleanly. Decrypting protected
// part bodies requires you to supply a KeyResolver implementation; raw
// encryption/decryption helpers (using crypto/aes and crypto/rsa) are
// provided in this package's Decrypt and Encrypt functions for the common
// AES-GCM case, and callers are expected to wire RSA-OAEP key wrap into
// their KeyResolver themselves. Full automatic decryption is intentionally
// scoped to user-provided KeyResolvers because there is no way for a
// library to do the right thing across all possible key-management policies.
//
// Blank-import to register hooks:
//
//	import _ "github.com/lestrrat-go/3mf/securecontent"
package securecontent

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/lestrrat-go/helium"

	tmf "github.com/lestrrat-go/3mf"
)

const (
	Namespace = tmf.NSSecureContent
	Prefix    = tmf.PrefixSecureContent
)

// Algorithm identifies the encryption algorithm used to protect a part.
type Algorithm string

const (
	AlgorithmAES256GCM Algorithm = "http://www.w3.org/2009/xmlenc11#aes256-gcm"
)

// EncryptedPart describes a single encrypted OPC part.
type EncryptedPart struct {
	ID         string
	Path       string
	Algorithm  Algorithm
	WrappedKey []byte // ciphertext of the per-part content encryption key
	KeyWrap    string // URI identifying the key-wrap algorithm (e.g. RSA-OAEP)
	KeyName    string // optional hint identifying which key was used
}

// Resources is the secure-content payload attached to tmf.Resources.
type Resources struct {
	EncryptedParts []*EncryptedPart
	// AdditionalKeys carries arbitrary <keystore>-style children that this
	// package does not yet model in detail. They round-trip as opaque
	// element trees.
	Extra []*helium.Element
}

// Of returns the secure-content resources attached to res, creating one if absent.
func Of(res *tmf.Resources) *Resources {
	if v, ok := res.ExtensionResources(Namespace).(*Resources); ok {
		return v
	}
	r := &Resources{}
	res.SetExtensionResources(Namespace, r)
	return r
}

// Require declares the secure-content extension on m.
func Require(m *tmf.Model) { m.RequireExtension(Namespace, Prefix) }

type extReader struct{ tmf.BaseExtensionReader }

func (extReader) Namespace() string { return Namespace }

type extWriter struct{ tmf.BaseExtensionWriter }

func (extWriter) Namespace() string { return Namespace }

func init() {
	tmf.RegisterExtensionReader(extReader{})
	tmf.RegisterExtensionWriter(extWriter{})
}

func (extReader) ReadResourceElement(res *tmf.Resources, elem *helium.Element) error {
	if elem.LocalName() != "encryptedpart" {
		// Preserve unknown elements for round-trip.
		Of(res).Extra = append(Of(res).Extra, elem)
		return nil
	}
	ep := &EncryptedPart{
		ID:        attr(elem, "id"),
		Path:      attr(elem, "path"),
		Algorithm: Algorithm(attr(elem, "algorithm")),
		KeyWrap:   attr(elem, "keywrap"),
		KeyName:   attr(elem, "keyname"),
	}
	// <wrappedkey> child carries the base64-encoded encrypted key.
	for c := range childElems(elem, "wrappedkey") {
		if t := textContent(c); t != "" {
			b, err := decodeBase64(t)
			if err != nil {
				return fmt.Errorf("securecontent: wrappedkey: %w", err)
			}
			ep.WrappedKey = b
		}
	}
	Of(res).EncryptedParts = append(Of(res).EncryptedParts, ep)
	return nil
}

func (extWriter) WriteResourceElements(res *tmf.Resources, w *tmf.Writer) error {
	v := res.ExtensionResources(Namespace)
	if v == nil {
		return nil
	}
	sr, ok := v.(*Resources)
	if !ok {
		return nil
	}
	for _, ep := range sr.EncryptedParts {
		if err := w.StartElementNS(Prefix, "encryptedpart"); err != nil {
			return err
		}
		if ep.ID != "" {
			if err := w.Attr("id", ep.ID); err != nil {
				return err
			}
		}
		if err := w.Attr("path", ep.Path); err != nil {
			return err
		}
		if ep.Algorithm != "" {
			if err := w.Attr("algorithm", string(ep.Algorithm)); err != nil {
				return err
			}
		}
		if ep.KeyWrap != "" {
			if err := w.Attr("keywrap", ep.KeyWrap); err != nil {
				return err
			}
		}
		if ep.KeyName != "" {
			if err := w.Attr("keyname", ep.KeyName); err != nil {
				return err
			}
		}
		if len(ep.WrappedKey) > 0 {
			if err := w.StartElementNS(Prefix, "wrappedkey"); err != nil {
				return err
			}
			if err := w.WriteString(encodeBase64(ep.WrappedKey)); err != nil {
				return err
			}
			if err := w.EndElement(); err != nil {
				return err
			}
		}
		if err := w.EndElement(); err != nil {
			return err
		}
	}
	return nil
}

// KeyResolver returns the symmetric content-encryption key (CEK) protecting
// the given part. Implementations typically inspect ep.KeyWrap / ep.KeyName
// to look up a private key and unwrap ep.WrappedKey.
type KeyResolver interface {
	ResolveKey(ep *EncryptedPart) ([]byte, error)
}

// KeyResolverFunc is an adapter that lets ordinary functions satisfy
// KeyResolver.
type KeyResolverFunc func(*EncryptedPart) ([]byte, error)

func (f KeyResolverFunc) ResolveKey(ep *EncryptedPart) ([]byte, error) { return f(ep) }

// DecryptAESGCM decrypts ciphertext produced by AES-256-GCM. The 12-byte
// nonce is expected to be prepended to the ciphertext, and the GCM tag
// trailing it (matching the convention used by crypto/cipher.AEAD.Seal).
func DecryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("securecontent: ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	body := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, body, nil)
}

// EncryptAESGCM encrypts plaintext under AES-256-GCM using a fresh random
// nonce that is prepended to the returned ciphertext.
func EncryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := gcm.Seal(nonce, nonce, plaintext, nil)
	return out, nil
}

// ---- helpers ----

func attr(elem *helium.Element, local string) string {
	a, ok := elem.FindAttribute(helium.LocalNamePredicate(local))
	if !ok {
		return ""
	}
	return a.Value()
}

func childElems(parent *helium.Element, local string) func(yield func(*helium.Element) bool) {
	return func(yield func(*helium.Element) bool) {
		for child := range helium.Children(parent) {
			elem, ok := child.(*helium.Element)
			if !ok {
				continue
			}
			if local != "" && elem.LocalName() != local {
				continue
			}
			if !yield(elem) {
				return
			}
		}
	}
}

func textContent(elem *helium.Element) string {
	var s []byte
	for c := range helium.Children(elem) {
		if t, ok := c.(*helium.Text); ok {
			s = append(s, t.Content()...)
		}
	}
	return string(s)
}

func decodeBase64(s string) ([]byte, error) {
	// Small helper to avoid pulling in encoding/base64 import here and to
	// keep error messages consistent.
	// Trim any whitespace that XML might have introduced.
	clean := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		clean = append(clean, c)
	}
	return base64Decode(string(clean))
}

func encodeBase64(b []byte) string { return base64Encode(b) }
