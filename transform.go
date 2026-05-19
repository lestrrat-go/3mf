package tmf

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMatrix parses a 3MF transform string ("m00 m01 m02 m10 m11 m12 m20 m21 m22 m30 m31 m32",
// 12 whitespace-separated floats) and returns the corresponding Matrix.
// An empty string yields IdentityMatrix().
func ParseMatrix(s string) (Matrix, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return IdentityMatrix(), nil
	}
	fields := strings.Fields(s)
	if len(fields) != 12 {
		return Matrix{}, fmt.Errorf("tmf: transform must have 12 components, got %d", len(fields))
	}
	var m Matrix
	for i, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return Matrix{}, fmt.Errorf("tmf: transform component %d: %w", i, err)
		}
		m[i] = v
	}
	return m, nil
}

// FormatMatrix serializes m as a 12-component, space-separated transform
// string suitable for use as a transform attribute. The shortest decimal
// representation that round-trips is used.
func FormatMatrix(m Matrix) string {
	var b strings.Builder
	for i, v := range m {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
	}
	return b.String()
}
