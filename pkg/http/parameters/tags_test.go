package parameters_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/parameters"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestTagLookupKeyFollowsNamingConvention_TableDriven_ReturnsExpected(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		lookupKey string
		expect    bool
	}{
		{name: "ContentType header", lookupKey: headers.ContentType, expect: true},
		{name: "single lowercase letter", lookupKey: "a", expect: true},
		{name: "single uppercase letter", lookupKey: "A", expect: true},
		{name: "two lowercase letters", lookupKey: "aa", expect: true},
		{name: "lowercase followed by uppercase", lookupKey: "aA", expect: true},
		{name: "letter followed by digit", lookupKey: "a0", expect: true},
		{name: "letter followed by dash", lookupKey: "a-", expect: true},
		{name: "letter followed by underscore", lookupKey: "a_", expect: true},
		{name: "mixed valid characters", lookupKey: "aaA-_", expect: true},
		{name: "empty string", lookupKey: "", expect: false},
		{name: "single digit", lookupKey: "0", expect: false},
		{name: "digit followed by letter", lookupKey: "0a", expect: false},
		{name: "leading space", lookupKey: " name", expect: false},
		{name: "trailing space", lookupKey: "name ", expect: false},
		{name: "space in middle", lookupKey: "na me", expect: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			actual := parameters.TagLookupKeyFollowsNamingConvention(testCase.lookupKey)
			assert.Equals(t, testCase.expect, actual)
		})
	}
}

func TestExtractAndValidateFieldTagLookupKeys_ProperlyFormattedTags_ReturnsTagMappings(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		QueryField1  string `json:"-"               otherTag1:"value"      urlQuery:"Query1"`
		QueryField2  string `json:"-"               otherTag1:"value"      urlQuery:"Query2"`
		HeaderField1 string `httpHeader:"Header1"   json:"-"               otherTag2:"value1"`
		HeaderField2 string `httpHeader:"Header2"   json:"-"               otherTag2:"value2"`
		PathField1   string `json:"-"               otherTag3:""           urlPath:"Path1"`
		PathField2   string `json:"-"               otherTag4:"!@#$%^&*()" urlPath:"Path2"`
		JSONField1   string `json:"json1,omitempty"`
		JSONField2   string `json:"json2,omitempty"`
	}

	for range 3 {
		tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
		assert.NoError(t, err)

		_, hasQueryTag := tagToLookupKeyToFieldName[parameters.QueryTag]
		assert.True(t, hasQueryTag)
		_, hasHeaderTag := tagToLookupKeyToFieldName[parameters.HeaderTag]
		assert.True(t, hasHeaderTag)
		_, hasPathTag := tagToLookupKeyToFieldName[parameters.PathTag]
		assert.True(t, hasPathTag)

		assert.Equals(t, len(tagToLookupKeyToFieldName[parameters.QueryTag]), 2)
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.QueryTag]["query1"], "QueryField1")
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.QueryTag]["query2"], "QueryField2")

		assert.Equals(t, len(tagToLookupKeyToFieldName[parameters.HeaderTag]), 2)
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.HeaderTag]["header1"], "HeaderField1")
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.HeaderTag]["header2"], "HeaderField2")

		assert.Equals(t, len(tagToLookupKeyToFieldName[parameters.PathTag]), 2)
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.PathTag]["Path1"], "PathField1")
		assert.Equals(t, tagToLookupKeyToFieldName[parameters.PathTag]["Path2"], "PathField2")
	}
}

func TestExtractAndValidateFieldTagLookupKeys_DuplicateTagFields_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field1 string `json:"-" urlQuery:"QueryField"`
		Field2 int    `json:"-" urlQuery:"QueryField"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_DuplicateTagFieldsDifferentCases_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field1 string `json:"-" urlQuery:"QueryField"`
		Field2 int    `json:"-" urlQuery:"qUeRyfIeLd"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_OverlappingTags_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `httpHeader:"HeaderField" json:"-" urlQuery:"QueryField"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_MissingJsonTag_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `urlQuery:"QueryField"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_WrongJsonTagFormat_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `httpHeader:"QueryField" json:"notRight"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_EmptyTagValue_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `json:"-" urlPath:""`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_TagValueStartsWithDigit_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `json:"-" urlQuery:"0invalidName"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_TagValueWithSpace_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field string `httpHeader:"invalid name" json:"-"`
	}
	tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
	assert.Error(t, err)
	assert.Nil(t, tagToLookupKeyToFieldName)
}

func TestExtractAndValidateFieldTagLookupKeys_NonStructGeneric_Panics(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[string]()
	})
}
