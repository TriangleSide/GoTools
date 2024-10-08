package assert_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

type testRecorder struct {
	t           *testing.T
	helperCount int
	errorCount  int
	fatalCount  int
	logs        []string
}

func (tr *testRecorder) Name() string {
	return tr.t.Name()
}

func (tr *testRecorder) Helper() {
	tr.helperCount++
}

func (tr *testRecorder) Error(args ...any) {
	tr.errorCount++
	tr.logs = append(tr.logs, fmt.Sprint(args...))
}

func (tr *testRecorder) Fatal(args ...any) {
	tr.fatalCount++
	tr.logs = append(tr.logs, fmt.Sprint(args...))
}

func newTestRecorder(t *testing.T) *testRecorder {
	return &testRecorder{
		t:           t,
		helperCount: 0,
		errorCount:  0,
		fatalCount:  0,
		logs:        make([]string, 0),
	}
}

func TestAssertFunctions(t *testing.T) {
	t.Parallel()

	checkRecorder := func(t *testing.T, tr *testRecorder, errorCount int, fatalCount int, expectedLogs []string) {
		t.Helper()
		if tr.errorCount != errorCount {
			t.Fatalf("Incorrect error count. Wanted %d but got %d.", errorCount, tr.errorCount)
		}
		if tr.fatalCount != fatalCount {
			t.Fatalf("Incorrect fatal count. Wanted %d but got %d.", fatalCount, tr.fatalCount)
		}
		aggregatedLogs := strings.Join(tr.logs, "\n")
		for _, log := range expectedLogs {
			if !strings.Contains(aggregatedLogs, log) {
				t.Fatalf("Incorrect error message. Wanted '%s' to contain '%s'.", aggregatedLogs, log)
			}
		}
	}

	t.Run("assert test cases", func(t *testing.T) {
		t.Parallel()
		var testCases = []struct {
			name       string
			callback   func(*testRecorder, ...assert.Option)
			expectLogs []string
		}{
			{
				name: "Equals positive case - 1 and 1",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, 1, 1, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals negative case - 1 and 2",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, 1, 2, opts...)
				},
				expectLogs: []string{"Expected 1 to equal 2."},
			},
			{
				name: "Equals positive case - Comparing nil with nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, nil, nil, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals negative case - Comparing nil with non-nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, nil, 1, opts...)
				},
				expectLogs: []string{"Expected <nil> to equal 1."},
			},
			{
				name: "Equals negative case - Comparing different types",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, 0, 0.0, opts...)
				},
				expectLogs: []string{"Expected 0 to equal 0."},
			},
			{
				name: "Equals positive case - Comparing empty slices",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Equals(tr, []int{}, []int{}, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals positive case - Comparing nil slices",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var s1 []int
					var s2 []int
					assert.Equals(tr, s1, s2, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals negative case - Comparing nil slice with empty slice",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var s1 []int
					s2 := []int{}
					assert.Equals(tr, s1, s2, opts...)
				},
				expectLogs: []string{"Expected [] to equal []."},
			},
			{
				name: "Equals positive case - Comparing maps with same content",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					m1 := map[string]int{"a": 1}
					m2 := map[string]int{"a": 1}
					assert.Equals(tr, m1, m2, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals positive case - Comparing nil maps",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var m1 map[string]int
					var m2 map[string]int
					assert.Equals(tr, m1, m2, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Equals negative case - Comparing nil map with empty map",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var m1 map[string]int
					m2 := map[string]int{}
					assert.Equals(tr, m1, m2, opts...)
				},
				expectLogs: []string{"Expected map[] to equal map[]."},
			},
			{
				name: "NotEquals positive case - Comparing different integers",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, 1, 2, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NotEquals negative case - Comparing equal integers",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, 1, 1, opts...)
				},
				expectLogs: []string{"Expected arguments 1 to differ."},
			},
			{
				name: "NotEquals negative case - Comparing nil with nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, nil, nil, opts...)
				},
				expectLogs: []string{"Expected arguments <nil> to differ."},
			},
			{
				name: "NotEquals positive case - Comparing nil with non-nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, nil, 1, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NotEquals positive case - Comparing zero values of different types",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, 0, 0.0, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NotEquals negative case - Comparing empty slices",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotEquals(tr, []int{}, []int{}, opts...)
				},
				expectLogs: []string{"Expected arguments [] to differ."},
			},
			{
				name: "Panic positive case - Function panics as expected",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Panic(tr, func() { panic("test panic") }, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Panic negative case - Function does not panic",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Panic(tr, func() {}, opts...)
				},
				expectLogs: []string{"Expected panic to occur but none occurred."},
			},
			{
				name: "Panic positive case - Function panics with integer",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Panic(tr, func() { panic(42) }, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "PanicExact positive case - Panic with exact message",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicExact(tr, func() { panic("exact message") }, "exact message", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "PanicExact negative case - Panic message differs",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicExact(tr, func() { panic("wrong message") }, "exact message", opts...)
				},
				expectLogs: []string{
					"Expected panic message to equal 'exact message' but got 'wrong message'.",
				},
			},
			{
				name: "PanicExact negative case - Panic with non-string message",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicExact(tr, func() { panic(42) }, "42", opts...)
				},
				expectLogs: []string{
					"Could not extract error message from panic.",
				},
			},
			{
				name: "PanicExact positive case - Panic with empty string message",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicExact(tr, func() { panic("") }, "", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "PanicPart positive case - Panic message contains expected part",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicPart(tr, func() { panic(errors.New("some panic message")) }, "panic", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "PanicPart negative case - Panic message does not contain expected part",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicPart(tr, func() { panic("wrong message") }, "expected part", opts...)
				},
				expectLogs: []string{
					"Expected panic message to contain 'expected part' but got 'wrong message'.",
				},
			},
			{
				name: "PanicPart negative case - Panic message missing part",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.PanicPart(tr, func() { panic("other message") }, "some part", opts...)
				},
				expectLogs: []string{
					"Expected panic message to contain 'some part' but got 'other message'.",
				},
			},
			{
				name: "Error positive case - Error is not nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Error(tr, errors.New("some error"), opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Error negative case - Error is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Error(tr, nil, opts...)
				},
				expectLogs: []string{
					"Expecting an error but none occurred.",
				},
			},
			{
				name: "Error negative case - Error interface is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var err error = nil
					assert.Error(tr, err, opts...)
				},
				expectLogs: []string{
					"Expecting an error but none occurred.",
				},
			},
			{
				name: "ErrorExact positive case - Error message matches exactly",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.ErrorExact(tr, errors.New("exact error"), "exact error", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "ErrorExact negative case - Error message differs",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.ErrorExact(tr, errors.New("wrong error"), "exact error", opts...)
				},
				expectLogs: []string{
					"Expected the error message 'exact error' but got 'wrong error'.",
				},
			},
			{
				name: "ErrorExact negative case - Error is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var err error = nil
					assert.ErrorExact(tr, err, "some error", opts...)
				},
				expectLogs: []string{
					"Expecting an error but none occurred.",
				},
			},
			{
				name: "ErrorExact positive case - Error message is empty string",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					err := errors.New("")
					assert.ErrorExact(tr, err, "", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "ErrorPart positive case - Error message contains part",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.ErrorPart(tr, errors.New("some error occurred"), "error", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "ErrorPart negative case - Error message does not contain part",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.ErrorPart(tr, errors.New("wrong error"), "expected part", opts...)
				},
				expectLogs: []string{
					"Expected the error message to contain 'expected part' but got 'wrong error'.",
				},
			},
			{
				name: "ErrorPart negative case - Error is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var err error = nil
					assert.ErrorPart(tr, err, "some part", opts...)
				},
				expectLogs: []string{
					"Expecting an error but none occurred.",
				},
			},
			{
				name: "NoError positive case - Error is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NoError(tr, nil, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NoError negative case - Error is not nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NoError(tr, errors.New("unexpected error"), opts...)
				},
				expectLogs: []string{
					"Not expecting an error to occur. Got unexpected error.",
				},
			},
			{
				name: "NoError positive case - Error interface is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var err error = nil
					assert.NoError(tr, err, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NoError negative case - Error is non-nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					err := errors.New("some error")
					assert.NoError(tr, err, opts...)
				},
				expectLogs: []string{
					"Not expecting an error to occur. Got some error.",
				},
			},
			{
				name: "Nil positive case - Value is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Nil(tr, nil, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Nil negative case - Value is a nil integer slice",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					var value []int
					assert.Nil(tr, value, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Nil negative case - Value is not nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Nil(tr, "not nil", opts...)
				},
				expectLogs: []string{
					"Expecting nil value but value is not nil.",
				},
			},
			{
				name: "NotNil positive case - Value is not nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotNil(tr, "not nil", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "NotNil negative case - Value is nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.NotNil(tr, nil, opts...)
				},
				expectLogs: []string{
					"Expecting the value to not be nil.",
				},
			},
			{
				name: "True positive case",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.True(tr, true, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "True negative case",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.True(tr, false, opts...)
				},
				expectLogs: []string{
					"Expecting the value to be true.",
				},
			},
			{
				name: "False positive case",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.False(tr, false, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "False negative case",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.False(tr, true, opts...)
				},
				expectLogs: []string{
					"Expecting the value to be false.",
				},
			},
			{
				name: "Contains positive case - String and string",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Contains(tr, "test string", "test", opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "Contains negative case - String and string",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Contains(tr, "test string", "not a substring", opts...)
				},
				expectLogs: []string{
					"Expecting 'test string' to contain 'not a substring'.",
				},
			},
			{
				name: "Contains negative case - nil and nil",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.Contains(tr, nil, nil, opts...)
				},
				expectLogs: []string{
					"Unknown types for the contains check.",
				},
			},
			{
				name: "FloatEquals positive case - 123.45 with 123.46 and 0.1 epsilon ",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.FloatEquals[float32](tr, 123.45, 123.46, 0.1, opts...)
				},
				expectLogs: []string{},
			},
			{
				name: "FloatEquals negative case - 123.45 with 123.46 and 0.001 epsilon ",
				callback: func(tr *testRecorder, opts ...assert.Option) {
					assert.FloatEquals[float32](tr, 123.45, 123.46, 0.0001, opts...)
				},
				expectLogs: []string{
					"Expecting 123.",
					"to equal 123.",
					"within a margin of 0.",
				},
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				tr := newTestRecorder(t)
				testCase.callback(tr)
				if len(testCase.expectLogs) > 0 {
					checkRecorder(t, tr, 0, 1, testCase.expectLogs)
				} else {
					checkRecorder(t, tr, 0, 0, []string{})
				}
				tr = newTestRecorder(t)
				testCase.callback(tr, assert.Continue())
				if len(testCase.expectLogs) > 0 {
					checkRecorder(t, tr, 1, 0, testCase.expectLogs)
				} else {
					checkRecorder(t, tr, 0, 0, []string{})
				}
			})
		}
	})
}
