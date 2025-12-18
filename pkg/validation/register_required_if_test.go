package validation_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestRequiredIfValidator_UsedWithVar_ReturnsError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		value             any
		rule              string
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:              "required_if used with Var",
			value:             "testValue",
			rule:              "required_if=Status active",
			expectedErrorPart: "required_if can only be used on struct fields",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var(testCase.value, testCase.rule)
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestRequiredIfValidator_ConditionMatches_EnforcesRequired(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		validate          func() error
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name: "string condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "string condition matches with extra whitespace and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status   active"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "non-string condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Count int
					Field string `validate:"required_if=Count 10"`
				}
				return validation.Struct(TestStruct{
					Count: 10,
					Field: "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "non-string pointer condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Count *int
					Field string `validate:"required_if=Count 10"`
				}
				return validation.Struct(TestStruct{
					Count: ptr.Of(10),
					Field: "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "pointer condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status *string
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: ptr.Of("active"),
					Field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "any condition holds string that matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status any
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: any("active"),
					Field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "any condition holds int that matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status any
					Field  string `validate:"required_if=Status 10"`
				}
				return validation.Struct(TestStruct{
					Status: any(10),
					Field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "numeric condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Code  int    `validate:"required"`
					Field string `validate:"required_if=Code 200"`
				}
				return validation.Struct(TestStruct{
					Code:  200,
					Field: "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "boolean condition matches and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Flag  bool   `validate:"required"`
					Field string `validate:"required_if=Flag true"`
				}
				return validation.Struct(TestStruct{
					Flag:  true,
					Field: "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name: "field under validation is a pointer and nil",
			validate: func() error {
				type TestStruct struct {
					Status string
					Field  *string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  nil,
				})
			},
			expectedErrorPart: "value is nil",
		},
		{
			name: "field under validation is unexported and empty",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					field:  "",
				})
			},
			expectedErrorPart: "the value is the zero-value",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.validate()
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestRequiredIfValidator_ConditionMatches_RequiredFieldPresent_Passes(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		validate func() error
	}

	testCases := []testCase{
		{
			name: "string condition matches and required field is set",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(&TestStruct{
					Status: "active",
					Field:  "some value",
				})
			},
		},
		{
			name: "field under validation is a pointer and set",
			validate: func() error {
				type TestStruct struct {
					Status string  `validate:"required"`
					Field  *string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  ptr.Of("some value"),
				})
			},
		},
		{
			name: "condition field is unexported and matches",
			validate: func() error {
				type TestStruct struct {
					status string
					Field  string `validate:"required_if=status active"`
				}
				return validation.Struct(TestStruct{
					status: "active",
					Field:  "not-empty",
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.validate()
			assert.NoError(t, err)
		})
	}
}

func TestRequiredIfValidator_ConditionDoesNotMatch_DoesNotEnforceRequired(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		validate func() error
	}

	testCases := []testCase{
		{
			name: "string condition does not match and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "inactive",
					Field:  "",
				})
			},
		},
		{
			name: "string condition does not match and required field is set",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: "inactive",
					Field:  "some value",
				})
			},
		},
		{
			name: "non-string condition does not match and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Count int
					Field string `validate:"required_if=Count 10"`
				}
				return validation.Struct(TestStruct{
					Count: 5,
					Field: "",
				})
			},
		},
		{
			name: "boolean condition does not match and required field is zero",
			validate: func() error {
				type TestStruct struct {
					Flag  bool
					Field string `validate:"required_if=Flag true"`
				}
				return validation.Struct(TestStruct{
					Flag:  false,
					Field: "",
				})
			},
		},
		{
			name: "condition field is any and does not match",
			validate: func() error {
				type TestStruct struct {
					Status any
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: any("inactive"),
					Field:  "",
				})
			},
		},
		{
			name: "condition field is any and nil",
			validate: func() error {
				type TestStruct struct {
					Status any
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: nil,
					Field:  "",
				})
			},
		},
		{
			name: "condition field is a pointer and nil",
			validate: func() error {
				type TestStruct struct {
					Status *string
					Field  string `validate:"required_if=Status active"`
				}
				return validation.Struct(TestStruct{
					Status: nil,
					Field:  "",
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.validate()
			assert.NoError(t, err)
		})
	}
}

func TestRequiredIfValidator_MissingConditionField_ReturnsError(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Field string `validate:"required_if=Status active"`
	}
	err := validation.Struct(TestStruct{
		Field: "some value",
	})
	assert.ErrorPart(t, err, "field Status does not exist in the struct")
}

func TestRequiredIfValidator_InvalidParameters_ReturnsError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		validate          func() error
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name: "missing comparison value",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  "",
				})
			},
			expectedErrorPart: "required_if requires a field name and a value to compare",
		},
		{
			name: "empty parameters",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if="`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  "",
				})
			},
			expectedErrorPart: "required_if requires a field name and a value to compare",
		},
		{
			name: "too many parts",
			validate: func() error {
				type TestStruct struct {
					Status string `validate:"required"`
					Field  string `validate:"required_if=Status active extra"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  "",
				})
			},
			expectedErrorPart: "required_if requires a field name and a value to compare",
		},
		{
			name: "poorly formatted after dive",
			validate: func() error {
				type TestStruct struct {
					Status string   `validate:"required"`
					Field  []string `validate:"dive,required_if=Status"`
				}
				return validation.Struct(TestStruct{
					Status: "active",
					Field:  []string{""},
				})
			},
			expectedErrorPart: "required_if requires a field name and a value to compare",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.validate()
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestRequiredIfValidator_OmitemptyPresent_SkipsRequiredIf(t *testing.T) {
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
}

func TestRequiredIfValidator_MultipleConditionsAnyMatch_EnforcesRequired(t *testing.T) {
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
}

func TestRequiredIfValidator_AfterDive_ValidatesElements(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		value             []string
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:              "element is empty",
			value:             []string{""},
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "element is not empty",
			value:             []string{"value"},
			expectedErrorPart: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			type TestStruct struct {
				Status string   `validate:"required"`
				Field  []string `validate:"dive,required_if=Status active"`
			}

			err := validation.Struct(TestStruct{
				Status: "active",
				Field:  testCase.value,
			})

			if testCase.expectedErrorPart != "" {
				assert.ErrorPart(t, err, testCase.expectedErrorPart)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestRequiredIfValidator_ConcurrentValidation_PassesConsistently(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Status string `validate:"required"`
		Field  string `validate:"required_if=Status active"`
	}

	const goroutineCount = 50
	errorsCh := make(chan error, goroutineCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)
	for range goroutineCount {
		go func() {
			defer waitGroup.Done()
			errorsCh <- validation.Struct(TestStruct{
				Status: "active",
				Field:  "some value",
			})
		}()
	}
	waitGroup.Wait()
	close(errorsCh)

	for err := range errorsCh {
		assert.NoError(t, err)
	}
}
