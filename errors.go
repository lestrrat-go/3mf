package tmf

import "errors"

var (
	// ErrInvalidPackage is returned when an OPC package cannot be read because
	// it is malformed or missing required parts.
	ErrInvalidPackage = errors.New("tmf: invalid package")

	// ErrNoModelPart is returned when an OPC package has no relationship that
	// points to a 3MF model part.
	ErrNoModelPart = errors.New("tmf: no 3D model part")

	// ErrUnknownExtension is returned when a parser encounters a namespace
	// declared in the requiredextensions list that this library does not
	// implement.
	ErrUnknownExtension = errors.New("tmf: unknown required extension")

	// ErrMalformedModel is returned when the 3dmodel.model XML does not
	// match the expected schema.
	ErrMalformedModel = errors.New("tmf: malformed model XML")
)
