package production

import (
	tmf "github.com/lestrrat-go/3mf"
)

// Namespace is the URI of the 3MF Production Extension.
const Namespace = tmf.NSProduction

// Prefix is the conventional prefix used for the Production extension.
const Prefix = tmf.PrefixProduction

type reader struct{ tmf.BaseExtensionReader }

func (reader) Namespace() string { return Namespace }

type writer struct{ tmf.BaseExtensionWriter }

func (writer) Namespace() string { return Namespace }

func init() {
	tmf.RegisterExtensionReader(reader{})
	tmf.RegisterExtensionWriter(writer{})
}

// Require declares the Production extension on m using the conventional
// prefix. After calling Require, the model will declare xmlns:p on its
// root element and add "p" to its requiredextensions attribute.
func Require(m *tmf.Model) { m.RequireExtension(Namespace, Prefix) }
