package parameters_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestTags(t *testing.T) {
	t.Parallel()

	t.Run("lookup key name validation", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			lookupKey string
			expect    bool
		}{
			{lookupKey: headers.ContentType, expect: true},
			{lookupKey: "a", expect: true},
			{lookupKey: "A", expect: true},
			{lookupKey: "aa", expect: true},
			{lookupKey: "aA", expect: true},
			{lookupKey: "a0", expect: true},
			{lookupKey: "a-", expect: true},
			{lookupKey: "a_", expect: true},
			{lookupKey: "aaA-_", expect: true},
			{lookupKey: "", expect: false},
			{lookupKey: "0", expect: false},
			{lookupKey: "0a", expect: false},
			{lookupKey: " name", expect: false},
			{lookupKey: "name ", expect: false},
			{lookupKey: "na me", expect: false},
		}

		for _, testCase := range testCases {
			actual := parameters.TagLookupKeyFollowsNamingConvention(testCase.lookupKey)
			assert.Equals(t, testCase.expect, actual)
		}
	})

	t.Run("it should succeed when the tags are properly formatted", func(t *testing.T) {
		t.Parallel()

		type testStruct struct {
			QueryField1  string `urlQuery:"Query1" json:"-" otherTag1:"value"`
			QueryField2  string `urlQuery:"Query2" json:"-" otherTag1:"value"`
			HeaderField1 string `httpHeader:"Header1" json:"-" otherTag2:"value1"`
			HeaderField2 string `httpHeader:"Header2" json:"-" otherTag2:"value2"`
			PathField1   string `urlPath:"Path1" json:"-" otherTag3:""`
			PathField2   string `urlPath:"Path2" json:"-" otherTag4:"!@#$%^&*()"`
			JSONField1   string `json:"JSON1,omitempty"`
			JSONField2   string `json:"JSON2,omitempty"`
		}

		for i := 0; i < 3; i++ {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
			assert.NoError(t, err)

			assert.True(t, tagToLookupKeyToFieldName.Has(parameters.QueryTag))
			assert.True(t, tagToLookupKeyToFieldName.Has(parameters.HeaderTag))
			assert.True(t, tagToLookupKeyToFieldName.Has(parameters.PathTag))

			assert.Equals(t, len(tagToLookupKeyToFieldName.Get(parameters.QueryTag)), 2)
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.QueryTag)["query1"], "QueryField1")
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.QueryTag)["query2"], "QueryField2")

			assert.Equals(t, len(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)), 2)
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.HeaderTag)["header1"], "HeaderField1")
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.HeaderTag)["header2"], "HeaderField2")

			assert.Equals(t, len(tagToLookupKeyToFieldName.Get(parameters.PathTag)), 2)
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.PathTag)["Path1"], "PathField1")
			assert.Equals(t, tagToLookupKeyToFieldName.Get(parameters.PathTag)["Path2"], "PathField2")
		}
	})

	t.Run("it should fail when validating a struct that has two fields with the same tag", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field1 string `urlQuery:"QueryField" json:"-"`
			Field2 int    `urlQuery:"QueryField" json:"-"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should fail when validating a struct that has two fields with the same tag in different cases", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field1 string `urlQuery:"QueryField" json:"-"`
			Field2 int    `urlQuery:"qUeRyfIeLd" json:"-"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should fail when validating a struct that has overlapping tags", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field string `urlQuery:"QueryField" httpHeader:"HeaderField" json:"-"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should fail when validating a struct that has no accompanying json tag", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field string `urlQuery:"QueryField"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should fail when validating a struct that has an accompanying json tag with the wrong format", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field string `httpHeader:"QueryField" json:"notRight"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should fail when validating a struct that has a tag that is empty", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Field string `urlPath:"" json:"-"`
		}
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.Error(t, err)
		assert.Nil(t, tagToLookupKeyToFieldName)
	})

	t.Run("it should panic when the generic isn't a struct", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[string]()
		})
	})

	t.Run("it should panic when the generic is a struct pointer", func(t *testing.T) {
		t.Parallel()
		type parameterStruct struct {
			Field string `urlQuery:"" json:"-"`
		}
		assert.Panic(t, func() {
			_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[*parameterStruct]()
		})
	})
}
