package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestRequiredIfValidator(t *testing.T) {
	t.Parallel()

	t.Run("required_if should fail if called with Var validation", func(t *testing.T) {
		t.Parallel()
		err := validation.Var("testValue", "required_if=Status active")
		assert.ErrorPart(t, err, "required_if can only be used on struct fields")
	})

	t.Run("when the condition field matches and the required field is set it should pass", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(&TestStruct{
			Status: "active",
			Field:  "some value",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field matches and the required field is zero it should fail", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when the condition field does not match and the required field is zero it should pass", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "inactive",
			Field:  "",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field does not match and the required field is set it should pass", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "inactive",
			Field:  "some value",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field is missing it should return an error", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Field string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Field: "some value",
		})
		assert.ErrorPart(t, err, "field Status does not exist in the struct")
	})

	t.Run("when the validator has invalid parameters it should return an error", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if=Status"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  "",
		})
		assert.ErrorPart(t, err, "required_if requires a field name and a value to compare")
	})

	t.Run("when the condition field is a non-string and matches the value it should enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Count int
			Field string `validate:"required_if=Count 10"`
		}
		err := validation.Struct(TestStruct{
			Count: 10,
			Field: "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when the condition field is a non-string and does not match the value it should not enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Count int
			Field string `validate:"required_if=Count 10"`
		}
		err := validation.Struct(TestStruct{
			Count: 5,
			Field: "",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field is a pointer and matches the value it should enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status *string
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: ptr.Of("active"),
			Field:  "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when the condition field is any and nil it should fail to check it", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status any
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: nil,
			Field:  "",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field is nil it should not enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status *string
			Field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: nil,
			Field:  "",
		})
		assert.NoError(t, err)
	})

	t.Run("when the field under validation is a pointer and required it should enforce required", func(t *testing.T) {
		type TestStruct struct {
			Status string
			Field  *string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  nil,
		})
		assert.ErrorPart(t, err, "found nil while dereferencing")
	})

	t.Run("when the field under validation is a pointer and set it should pass", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string  `validate:"required"`
			Field  *string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  ptr.Of("some value"),
		})
		assert.NoError(t, err)
	})

	t.Run("when the validator is used with parameters missing it should return an error", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"required_if="`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  "",
		})
		assert.ErrorPart(t, err, "required_if requires a field name and a value to compare")
	})

	t.Run("when the condition field value is numeric and matches it should enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Code  int    `validate:"required"`
			Field string `validate:"required_if=Code 200"`
		}
		err := validation.Struct(TestStruct{
			Code:  200,
			Field: "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when the condition field value is boolean and matches it should enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Flag  bool   `validate:"required"`
			Field string `validate:"required_if=Flag true"`
		}
		err := validation.Struct(TestStruct{
			Flag:  true,
			Field: "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when the condition field value is boolean and does not match it should not enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Flag  bool
			Field string `validate:"required_if=Flag true"`
		}
		err := validation.Struct(TestStruct{
			Flag:  false,
			Field: "",
		})
		assert.NoError(t, err)
	})

	t.Run("when the condition field is an unexported field it should succeed", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			status string
			Field  string `validate:"required_if=status active"`
		}
		err := validation.Struct(TestStruct{
			status: "active",
			Field:  "not-empty",
		})
		assert.NoError(t, err)
	})

	t.Run("when the field under validation is unexported it should perform the check", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			field  string `validate:"required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			field:  "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when using 'omitempty' with 'required_if' and the omitempty matches it should ignore the required_if", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Field  string `validate:"omitempty,required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  "",
		})
		assert.NoError(t, err)
	})

	t.Run("when multiple conditions are specified and any match it should enforce required", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string `validate:"required"`
			Type   string `validate:"required"`
			Field  string `validate:"required_if=Type admin,required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Type:   "user",
			Field:  "",
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when required_if is after a dive it should fail if it's the zero value", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string   `validate:"required"`
			Field  []string `validate:"dive,required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  []string{""},
		})
		assert.ErrorPart(t, err, "the value is the zero-value")
	})

	t.Run("when required_if is after a dive it should succeed if the value is not empty", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string   `validate:"required"`
			Field  []string `validate:"dive,required_if=Status active"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  []string{"value"},
		})
		assert.NoError(t, err)
	})

	t.Run("when required_if is poorly formatted after a dive it should return an error", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct {
			Status string   `validate:"required"`
			Field  []string `validate:"dive,required_if=Status"`
		}
		err := validation.Struct(TestStruct{
			Status: "active",
			Field:  []string{""},
		})
		assert.ErrorPart(t, err, "required_if requires a field name and a value to compare")
	})
}
