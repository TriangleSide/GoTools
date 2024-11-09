package logger

import (
	"context"
	"maps"
)

// contextKeyType is its own type to avoid collisions in the context.
type contextKeyType string

const (
	// contextKey is used to access the fields in the context.
	contextKey contextKeyType = "__loggerFields"
)

// AddField adds a field to the context for the logger.
func AddField(ctx *context.Context, key string, value any) Logger {
	fieldsNotCast := (*ctx).Value(contextKey)
	var newFields map[string]any
	if fieldsNotCast == nil {
		newFields = make(map[string]any, 1)
	} else {
		fields, fieldsCastOk := fieldsNotCast.(map[string]any)
		if !fieldsCastOk {
			panic("The entry context fields are not the correct type.")
		}
		newFields = make(map[string]any, len(fields)+1)
		maps.Copy(newFields, fields)
	}
	newFields[key] = value
	*ctx = context.WithValue(*ctx, contextKey, newFields)
	return &entry{
		fields: newFields,
	}
}

// AddFields adds many fields to the context for the logger.
func AddFields(ctx *context.Context, fieldsToAdd map[string]any) Logger {
	fieldsNotCast := (*ctx).Value(contextKey)
	var newFields map[string]any
	if fieldsNotCast == nil {
		newFields = make(map[string]any, len(fieldsToAdd))
	} else {
		fields, fieldsCastOk := fieldsNotCast.(map[string]any)
		if !fieldsCastOk {
			panic("The entry context fields are not the correct type.")
		}
		newFields = make(map[string]any, len(fields)+len(fieldsToAdd))
		maps.Copy(newFields, fields)
	}
	for k, v := range fieldsToAdd {
		newFields[k] = v
	}
	*ctx = context.WithValue(*ctx, contextKey, newFields)
	return &entry{
		fields: newFields,
	}
}

// FromCtx returns a Logger from the context.
// This should be used in conjunction with AddField and AddFields.
func FromCtx(ctx context.Context) Logger {
	fieldsNotCast := ctx.Value(contextKey)
	if fieldsNotCast == nil {
		return &entry{
			fields: nil,
		}
	}
	fields, fieldsCastOk := fieldsNotCast.(map[string]any)
	if !fieldsCastOk {
		panic("The entry context fields are not the correct type.")
	}
	return &entry{
		fields: fields,
	}
}
