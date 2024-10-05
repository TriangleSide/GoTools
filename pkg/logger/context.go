package logger

import (
	"context"
	"maps"
)

type contextKeyType string

const (
	contextKey contextKeyType = "__logCtx"
)

func WithField(ctx context.Context, key string, value any) context.Context {
	fieldNotCast := ctx.Value(contextKey)
	var newFields map[string]any
	if fieldNotCast == nil {
		newFields = make(map[string]any, 1)
	} else {
		fields := fieldNotCast.(map[string]any)
		newFields = make(map[string]any, len(fields)+1)
		maps.Copy(newFields, fields)
	}
	newFields[key] = value
	return context.WithValue(ctx, contextKey, newFields)
}

func WithFields(ctx context.Context, fieldsToAdd map[string]any) context.Context {
	fieldNotCast := ctx.Value(contextKey)
	var newFields map[string]any
	if fieldNotCast == nil {
		newFields = make(map[string]any, len(fieldsToAdd))
	} else {
		fields, fieldsCastOk := fieldNotCast.(map[string]any)
		if !fieldsCastOk {
			panic("The logger context fields are not the correct type.")
		}
		newFields = make(map[string]any, len(fields)+len(fieldsToAdd))
		maps.Copy(newFields, fields)
	}
	for k, v := range fieldsToAdd {
		newFields[k] = v
	}
	return context.WithValue(ctx, contextKey, newFields)
}
