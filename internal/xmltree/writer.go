package xmltree

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Writer writes XML while allowing attributes after StartElement.
type Writer struct {
	encoder *xml.Encoder
	pending *xml.StartElement
	stack   []xml.Name
}

// NewWriter returns a writer for XML output.
func NewWriter(output io.Writer) Writer {
	return Writer{encoder: xml.NewEncoder(output)}
}

// StartDocument writes the XML declaration.
func (w *Writer) StartDocument(version, encoding, standalone string) error {
	if version == "" {
		version = "1.0"
	}
	inst := []byte(`version="` + version + `"`)
	if encoding != "" {
		inst = append(inst, []byte(` encoding="`+encoding+`"`)...)
	}
	if standalone != "" {
		inst = append(inst, []byte(` standalone="`+standalone+`"`)...)
	}
	return w.encoder.EncodeToken(xml.ProcInst{Target: "xml", Inst: inst})
}

// EndDocument flushes the document and reports unclosed elements.
func (w *Writer) EndDocument() error {
	if len(w.stack) != 0 {
		return fmt.Errorf("xmltree: %d unclosed elements", len(w.stack))
	}
	return w.encoder.Close()
}

// StartElement starts an element. Attributes may follow before content.
func (w *Writer) StartElement(name string) error {
	if name == "" {
		return fmt.Errorf("xmltree: empty element name")
	}
	if err := w.flushPending(); err != nil {
		return err
	}
	xmlName := xml.Name{Local: name}
	w.pending = &xml.StartElement{Name: xmlName}
	w.stack = append(w.stack, xmlName)
	return nil
}

// WriteAttribute adds an attribute to the pending start element.
func (w *Writer) WriteAttribute(name, value string) error {
	if w.pending == nil {
		return fmt.Errorf("xmltree: attribute outside start element")
	}
	w.pending.Attr = append(w.pending.Attr, xml.Attr{Name: xml.Name{Local: name}, Value: value})
	return nil
}

// WriteString writes escaped character data.
func (w *Writer) WriteString(value string) error {
	if err := w.flushPending(); err != nil {
		return err
	}
	return w.encoder.EncodeToken(xml.CharData(value))
}

// EndElement closes the current element.
func (w *Writer) EndElement() error {
	if len(w.stack) == 0 {
		return fmt.Errorf("xmltree: end element without start element")
	}
	if err := w.flushPending(); err != nil {
		return err
	}
	name := w.stack[len(w.stack)-1]
	w.stack = w.stack[:len(w.stack)-1]
	return w.encoder.EncodeToken(xml.EndElement{Name: name})
}

func (w *Writer) flushPending() error {
	if w.pending == nil {
		return nil
	}
	start := *w.pending
	w.pending = nil
	return w.encoder.EncodeToken(start)
}
