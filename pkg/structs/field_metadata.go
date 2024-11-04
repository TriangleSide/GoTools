package structs

import (
	"reflect"

	"github.com/TriangleSide/GoTools/pkg/datastructures/readonly"
)

// FieldMetadata is the metadata extracted a struct fields.
type FieldMetadata struct {
	reflectType reflect.Type
	tags        *readonly.Map[string, string]
	anonymous   *readonly.Slice[string]
}

// Type returns the fields type.
func (f *FieldMetadata) Type() reflect.Type {
	return f.reflectType
}

// Tags returns a map of tag names to its respective tag content.
func (f *FieldMetadata) Tags() *readonly.Map[string, string] {
	return f.tags
}

// Anonymous returns the anonymous struct name chain before getting to a field.
//
//	type DeepExample struct {
//	  DeepField string
//	}
//
//	type Example struct {
//	  DeepExample
//	  Field string
//	}
//
// If calling Anonymous on the struct Example and the field DeepField, ["DeepExample"] would be returned.
func (f *FieldMetadata) Anonymous() *readonly.Slice[string] {
	return f.anonymous
}
