package gen

import (
	"testing"
)

func fieldNames(fields []fieldDef) []string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}

func TestMappableFields(t *testing.T) {
	longField := func(name, raw string) fieldDef {
		return fieldDef{Name: name, RawName: raw, Type: "int64", Func: "Long"}
	}
	bytesField := func(name, raw string) fieldDef {
		return fieldDef{Name: name, RawName: raw, Type: "byte", Func: "Bytes", Slice: true}
	}
	stringField := func(name, raw string) fieldDef {
		return fieldDef{Name: name, RawName: raw, Type: "string", Func: "String"}
	}

	photo := structDef{
		Name:    "Photo",
		RawName: "photo",
		Fields: []fieldDef{
			longField("ID", "id"),
			longField("AccessHash", "access_hash"),
			bytesField("FileReference", "file_reference"),
		},
	}
	inputPhotoFileLocation := structDef{
		Name:    "InputPhotoFileLocation",
		RawName: "inputPhotoFileLocation",
		Fields: []fieldDef{
			longField("ID", "id"),
			longField("AccessHash", "access_hash"),
			bytesField("FileReference", "file_reference"),
			stringField("ThumbSize", "thumb_size"),
		},
	}

	t.Run("ThumbSizeBecomesParameter", func(t *testing.T) {
		mapping, ok := mappableFields(photo, inputPhotoFileLocation)
		if !ok {
			t.Fatal("expected mapping to be generated")
		}
		if got := fieldNames(mapping.Params); len(got) != 1 || got[0] != "ThumbSize" {
			t.Fatalf("expected ThumbSize parameter, got %v", got)
		}
		// id/access_hash/file_reference must still be mapped from the source.
		if len(mapping.Fields) != 3 {
			t.Fatalf("expected 3 mapped fields, got %d", len(mapping.Fields))
		}
	})

	t.Run("UnmappableFieldRejected", func(t *testing.T) {
		// A required field that is neither mappable nor a known parameter must
		// still cause the mapping to be skipped.
		withExtra := inputPhotoFileLocation
		withExtra.Fields = append(append([]fieldDef{}, inputPhotoFileLocation.Fields...),
			longField("Secret", "secret"))
		if _, ok := mappableFields(photo, withExtra); ok {
			t.Fatal("expected mapping with unmappable required field to be skipped")
		}
	})
}

func TestParameterField(t *testing.T) {
	for _, tt := range []struct {
		name  string
		field fieldDef
		want  bool
	}{
		{"ThumbSize", fieldDef{RawName: "thumb_size", Type: "string"}, true},
		{"Conditional", fieldDef{RawName: "thumb_size", Type: "string", Conditional: true}, false},
		{"WrongType", fieldDef{RawName: "thumb_size", Type: "int64"}, false},
		{"Other", fieldDef{RawName: "access_hash", Type: "string"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := parameterField(structDef{}, tt.field); got != tt.want {
				t.Fatalf("parameterField() = %v, want %v", got, tt.want)
			}
		})
	}
}
