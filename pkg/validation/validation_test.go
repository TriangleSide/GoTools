package validation

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	t.Run("when a struct with validation rule is set and the set value is valid it should succeed", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, Struct(struct {
			Value int `validate:"gte=0"`
		}{Value: 0}))
	})

	t.Run("when a pointer is passed to the struct validation is should succeed if the value is valid", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, Struct(&struct {
			Value int `validate:"gte=0"`
		}{Value: 0}))
	})

	t.Run("when a validation rule is set and the set value is invalid it should fail", func(t *testing.T) {
		t.Parallel()
		assert.ErrorPart(t, Struct(struct {
			Value int `validate:"gte=0"`
		}{Value: -1}),
			"validation failed on field 'Value' with validator 'gte' and parameter(s) '0'")
	})

	t.Run("when a pointer is passed to the struct validation is should fail if the value is invalid", func(t *testing.T) {
		t.Parallel()
		assert.ErrorPart(t, Struct(&struct {
			Value int `validate:"gte=0"`
		}{Value: -1}),
			"validation failed on field 'Value' with validator 'gte' and parameter(s) '0'")
	})

	t.Run("when validating a struct with non-required field it should succeed", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, Struct(struct{ Value int }{Value: 0}))
	})

	t.Run("when calling the struct validation a non-struct value it should fail", func(t *testing.T) {
		t.Parallel()
		assert.PanicExact(t, func() {
			_ = Struct(0)
		}, "Type must be a struct or a pointer to a struct.")
	})

	t.Run("when validating nil it should fail", func(t *testing.T) {
		t.Parallel()
		assert.ErrorPart(t, Struct[*struct{}](nil), "struct validation on nil value")
	})

	t.Run("when using custom validator it should adhere to its logic", func(t *testing.T) {
		const errMsg = "custom validation failed"
		RegisterValidation("custom", func(fl validator.FieldLevel) bool {
			return fl.Field().Len() > 3
		}, func(err validator.FieldError) string {
			return errMsg
		})
		type testStruct struct {
			Name string `validate:"custom"`
		}

		t.Run("when custom validation rule is passed it should succeed", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, Struct(testStruct{Name: "abcd"}))
		})

		t.Run("when custom validation rule is violated it should fail", func(t *testing.T) {
			t.Parallel()
			assert.ErrorPart(t, Struct(testStruct{Name: "abc"}), errMsg)
		})
	})

	t.Run("when many validations fail it should list all errors", func(t *testing.T) {
		t.Parallel()
		err := Struct(struct {
			IntValue int    `validate:"gte=0"`
			StrValue string `validate:"required"`
		}{
			IntValue: -1,
			StrValue: "",
		})
		assert.ErrorPart(t, err, "validation failed on field 'IntValue' with validator 'gte'")
		assert.ErrorPart(t, err, "validation failed on field 'StrValue' with validator 'required'")
	})

	t.Run("when a variable satisfies the required tag it should succeed", func(t *testing.T) {
		t.Parallel()
		myInt := 1
		assert.NoError(t, Var(&myInt, "required,gt=0"))
	})

	t.Run("when a variable violates the validation tag it should fail", func(t *testing.T) {
		t.Parallel()
		myInt := 0
		assert.ErrorPart(t, Var(&myInt, "required,gt=0"),
			"validation failed with validator 'gt' and parameter(s) '0'")
	})

	t.Run("when registering the same custom validation twice it should panic", func(t *testing.T) {
		assert.Panic(t, func() {
			for i := 0; i < 2; i++ {
				RegisterValidation("multipleRegistrationTest",
					func(fl validator.FieldLevel) bool { return true },
					func(err validator.FieldError) string { return "" })
			}
		})
	})

	t.Run("when registering a validation with nil function it should panic", func(t *testing.T) {
		assert.Panic(t, func() {
			RegisterValidation("nil_func", nil, func(err validator.FieldError) string { return "" })
		})
		assert.Panic(t, func() {
			RegisterValidation("nil_msg_func", func(fl validator.FieldLevel) bool { return true }, nil)
		})
	})

	t.Run("when the error formatter is passed an error it doesn't recognize it should simply return the error", func(t *testing.T) {
		t.Parallel()
		assert.ErrorExact(t, formatErrorMessage(errors.New("test error")), "test error")
	})
}
