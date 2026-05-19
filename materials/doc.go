// Package materials implements the 3MF Materials and Properties Extension.
//
// The Materials extension adds:
//   - additional property-group resource elements: ColorGroup, Texture2DGroup,
//     CompositeMaterials, MultiProperties.
//   - the Texture2D resource (image references).
//   - per-object metallicdisplayproperties / specularproperties used by some
//     producers, modeled here as PBSpecular and PBMetallic resource groups.
//
// Triangle- and object-level pid/pindex attributes are already understood by
// the core tmf package; the property-group resources defined here are the
// resources those pids reference.
//
// Blank-import this package to register the extension hooks:
//
//	import _ "github.com/lestrrat-go/3mf/materials"
package materials
