package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Writer wraps an io.Writer with optional field filtering.
// It implements workflow.ResultWriter.
type Writer struct {
	w      io.Writer
	fields string
}

// New returns a Writer that writes to w, filtered to fields (JSON array string).
// Pass an empty fields string to write all fields.
func New(w io.Writer, fields string) *Writer {
	return &Writer{w: w, fields: fields}
}

// WriteResult marshals v to indented JSON and writes it, applying any field filter.
func (wr *Writer) WriteResult(v any) error {
	return Write(wr.w, v, wr.fields)
}

// WriteError writes err as a JSON error object to w.
func WriteError(w io.Writer, err error) error {
	return Write(w, map[string]string{"error": err.Error()}, "")
}

// Write marshals v to indented JSON and writes it to w.
// If fieldsJSON is non-empty it must be a JSON array of field name strings.
// Only those fields are included in the output; an unknown field name is an error.
// Pass an empty fieldsJSON to bypass filtering (e.g. for --schema output).
func Write(w io.Writer, v any, fieldsJSON string) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}

	if fieldsJSON == "" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return fmt.Errorf("formatting output: %w", err)
		}
		fmt.Fprintln(w, buf.String())
		return nil
	}

	var fields []string
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return fmt.Errorf("--fields must be a JSON array of strings (e.g. '[\"step\",\"instruction\"]'): %w", err)
	}

	if len(fields) == 0 {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return fmt.Errorf("formatting output: %w", err)
		}
		fmt.Fprintln(w, buf.String())
		return nil
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("filtering output: value is not a JSON object: %w", err)
	}

	for _, f := range fields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("unknown field %q", f)
		}
	}

	filtered := make(map[string]json.RawMessage, len(fields))
	for _, f := range fields {
		filtered[f] = m[f]
	}

	out, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling filtered output: %w", err)
	}
	fmt.Fprintln(w, string(out))
	return nil
}
