package mcp

import (
	"fmt"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
)

// uuidStringSchema is the correct JSON schema for uuid.UUID: a plain string in
// UUID format.  Without this override the jsonschema-go library infers the
// wrong schema — [16]byte becomes {type:array,items:{type:integer},minItems:16,
// maxItems:16} — because it operates on the underlying Go type and is unaware
// of uuid.UUID's custom JSON marshaller, which always encodes as a string.
var uuidStringSchema = &jsonschema.Schema{Type: "string", Format: "uuid"}

// uuidTypeSchemas is passed to jsonschema.ForOptions.TypeSchemas so that every
// uuid.UUID field (and *uuid.UUID field, after pointer dereferencing) is
// rendered as {type:"string",format:"uuid"} in the generated schema.
var uuidTypeSchemas = map[reflect.Type]*jsonschema.Schema{
	reflect.TypeFor[uuid.UUID](): uuidStringSchema,
}

// MustSchema generates a JSON schema for T with uuid.UUID fields correctly
// represented as {type:"string",format:"uuid"} rather than byte arrays.
// It panics if schema inference fails, which indicates a programming error
// (e.g. an unsupported field type) that should be caught at startup.
func MustSchema[T any]() *jsonschema.Schema {
	s, err := jsonschema.For[T](&jsonschema.ForOptions{
		TypeSchemas: uuidTypeSchemas,
	})
	if err != nil {
		var zero T
		panic(fmt.Sprintf("mcp.MustSchema[%T]: %v", zero, err))
	}
	return s
}
