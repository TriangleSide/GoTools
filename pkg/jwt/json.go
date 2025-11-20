package jwt

import (
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/structs"
)

const (
	// extraForQuotesColonAndComma counts the extra characters added per field in the JSON output.
	extraForQuotesColonAndComma = 6
	// extraForBraces counts the extra characters added for the opening and closing braces in the JSON output.
	extraForBraces = 2
)

// marshalToStableJSON takes a struct and marshals it to a JSON string with stable field ordering.
// Since this is only for JWTs, we can be assured that:
//   - the json tag is always present.
//   - the field is either string or int64.
//
// So we can skip error handling for those cases.
func marshalToStableJSON(v any) string {
	value := reflection.Dereference(reflect.ValueOf(v))
	metadata := structs.MetadataFromType(value.Type())

	strCount := 0
	fieldNameTJSONStringValue := make(map[string]string)
	for fieldName, fieldMetadata := range metadata.All() {
		jsonFieldName := fieldMetadata.Tags().Get("json")
		structValue, _ := structs.ValueFromName(value.Interface(), fieldName)
		if structValue.IsZero() {
			continue
		}
		var valueStr string
		if structValue.Kind() == reflect.String {
			valueStr = strconv.Quote(structValue.String())
		} else {
			valueStr = strconv.FormatInt(structValue.Int(), 10)
		}
		fieldNameTJSONStringValue[jsonFieldName] = valueStr
		strCount += len(jsonFieldName) + len(valueStr) + extraForQuotesColonAndComma
	}

	sortedFields := make([]string, 0, len(fieldNameTJSONStringValue))
	for fieldName := range fieldNameTJSONStringValue {
		sortedFields = append(sortedFields, fieldName)
	}
	slices.Sort(sortedFields)

	strCount += extraForBraces
	var sb strings.Builder
	sb.Grow(strCount)
	sb.WriteByte('{')
	for i, fieldName := range sortedFields {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('"')
		sb.WriteString(fieldName)
		sb.WriteString(`":`)
		sb.WriteString(fieldNameTJSONStringValue[fieldName])
	}
	sb.WriteByte('}')

	return sb.String()
}
