// Package opc implements just enough of the Open Packaging Conventions (ECMA-376
// Part 2) to support reading and writing 3MF (.3mf) packages.
//
// 3MF is distributed as an OPC package: a ZIP archive that contains
//   - a [Content_Types].xml file at the root listing the MIME types of the
//     parts inside the archive,
//   - a _rels/.rels file at the root listing the top-level relationships
//     that bootstrap the package (in particular, the relationship that
//     points to the primary 3D model part), and
//   - any number of "parts" (regular files inside the ZIP, with paths that
//     always start with "/"), each of which may have its own
//     <partdir>/_rels/<partname>.rels relationship file.
//
// This package is intentionally minimal: it implements just the pieces of
// OPC that 3MF actually uses. In particular it does not implement digital
// signatures, interleaved parts, or piece notation. Producers that rely on
// any of those features will not round-trip cleanly.
package opc
