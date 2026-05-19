// Package production implements the 3MF Production Extension.
//
// The Production Extension carries per-element UUIDs (on <model>, <build>,
// <item>, <object>, and <component>) and the path attribute that lets a
// build item or component reference an object that lives in a different
// model part. Those attributes are read and written directly by the core
// tmf package — production.Namespace simply exposes the namespace URI for
// callers that want to mark their model as requiring the extension via
// tmf.Model.RequireExtension. Importing this package for its side effects
// registers an empty hook so that LookupExtensionReader/Writer succeed for
// the namespace and so that extra production-specific child elements (none
// are currently defined by the spec) round-trip cleanly.
//
// Blank-import this package to "enable" the production extension:
//
//	import _ "github.com/lestrrat-go/3mf/production"
package production
