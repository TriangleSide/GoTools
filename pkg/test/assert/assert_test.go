package assert_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type testRecorder struct {
	t           *testing.T
	helperCount int
	errorCount  int
	fatalCount  int
	logs        []string
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
	t.Helper()
	return &testRecorder{
		t:           t,
		helperCount: 0,
		errorCount:  0,
		fatalCount:  0,
		logs:        make([]string, 0),
	}
}

func checkRecorder(t *testing.T, recorder *testRecorder, errorCount int, fatalCount int, expectedLogs []string) {
	t.Helper()
	if recorder.errorCount != errorCount {
		t.Fatalf("Incorrect error count. Wanted %d but got %d.", errorCount, recorder.errorCount)
	}
	if recorder.fatalCount != fatalCount {
		t.Fatalf("Incorrect fatal count. Wanted %d but got %d.", fatalCount, recorder.fatalCount)
	}
	aggregatedLogs := strings.Join(recorder.logs, "\n")
	for _, log := range expectedLogs {
		if !strings.Contains(aggregatedLogs, log) {
			t.Fatalf("Incorrect error message. Wanted '%s' to contain '%s'.", aggregatedLogs, log)
		}
	}
}

func TestAssert_DifferentCases_ShouldWorkCorrectly(t *testing.T) {
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
			expectLogs: []string{"Expected 1 and 2 to be equal."},
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
			expectLogs: []string{"Expected <nil> and 1 to be equal."},
		},
		{
			name: "Equals negative case - Comparing different types",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.Equals(tr, 0, 0.0, opts...)
			},
			expectLogs: []string{"Expected 0 and 0 to be equal."},
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
			expectLogs: []string{"Expected [] and [] to be equal."},
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
			expectLogs: []string{"Expected map[] and map[] to be equal."},
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
			expectLogs: []string{"Expected 1 and 1 to differ."},
		},
		{
			name: "NotEquals negative case - Comparing nil with nil",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.NotEquals(tr, nil, nil, opts...)
			},
			expectLogs: []string{"Expected <nil> and <nil> to differ."},
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
			expectLogs: []string{"Expected [] and [] to differ."},
		},
		{
			name: "Panic positive case - Function panics as expected",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.Panic(tr, func() { panic(errors.New("test panic")) }, opts...)
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
			name: "Panic positive case - Function panics with error",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.Panic(tr, func() { panic(errors.New("42")) }, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "Panic positive case - Nil function triggers panic",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var panicFunc func()
				assert.Panic(tr, panicFunc, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "PanicExact positive case - Panic with exact message",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicExact(tr, func() { panic(errors.New("exact message")) }, "exact message", opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "PanicExact negative case - Panic message differs",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicExact(tr, func() { panic(errors.New("wrong message")) }, "exact message", opts...)
			},
			expectLogs: []string{
				"Expected panic message to equal 'exact message' but got 'wrong message'.",
			},
		},
		{
			name: "PanicExact positive case - Panic with error message",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicExact(tr, func() { panic(errors.New("42")) }, "42", opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "PanicExact positive case - Panic with empty string message",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicExact(tr, func() { panic(errors.New("")) }, "", opts...)
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
				assert.PanicPart(tr, func() { panic(errors.New("wrong message")) }, "expected part", opts...)
			},
			expectLogs: []string{
				"Expected panic message to contain 'expected part' but got 'wrong message'.",
			},
		},
		{
			name: "PanicPart negative case - Panic message missing part",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicPart(tr, func() { panic(errors.New("other message")) }, "some part", opts...)
			},
			expectLogs: []string{
				"Expected panic message to contain 'some part' but got 'other message'.",
			},
		},
		{
			name: "PanicPart positive case - Panic with error message",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.PanicPart(tr, func() { panic(errors.New("42")) }, "42", opts...)
			},
			expectLogs: []string{},
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
			name: "Nil positive case - Value is a nil integer slice",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value []int
				assert.Nil(tr, value, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "Nil positive case - Value is a nil pointer",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value *int
				assert.Nil(tr, value, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "Nil positive case - Value is a nil map",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value map[string]int
				assert.Nil(tr, value, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "Nil positive case - Value is a nil chan",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value chan int
				assert.Nil(tr, value, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "Nil positive case - Value is a nil function",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value func()
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
			name: "NotNil negative case - Value is a typed nil pointer",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				var value *int
				assert.NotNil(tr, value, opts...)
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
			name: "Contains negative case - String and nil",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.Contains(tr, "test string", nil, opts...)
			},
			expectLogs: []string{
				"Unknown types for the contains check.",
			},
		},
		{
			name: "Contains negative case - Nil and string",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.Contains(tr, nil, "test", opts...)
			},
			expectLogs: []string{
				"Unknown types for the contains check.",
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
			name: "FloatEquals positive case - Difference equals epsilon",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.FloatEquals[float64](tr, 1.0, 1.5, 0.5, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "FloatEquals positive case - Zero epsilon with equal values",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.FloatEquals[float64](tr, 1.0, 1.0, 0.0, opts...)
			},
			expectLogs: []string{},
		},
		{
			name: "FloatEquals negative case - Zero epsilon with different values",
			callback: func(tr *testRecorder, opts ...assert.Option) {
				assert.FloatEquals[float64](tr, 1.0, 2.0, 0.0, opts...)
			},
			expectLogs: []string{
				"Expecting 1.",
				"to equal 2.",
				"within a margin of 0.",
			},
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
			recorder := newTestRecorder(t)
			testCase.callback(recorder)
			if len(testCase.expectLogs) > 0 {
				checkRecorder(t, recorder, 0, 1, testCase.expectLogs)
			} else {
				checkRecorder(t, recorder, 0, 0, []string{})
			}
			recorder = newTestRecorder(t)
			testCase.callback(recorder, assert.Continue())
			if len(testCase.expectLogs) > 0 {
				checkRecorder(t, recorder, 1, 0, testCase.expectLogs)
			} else {
				checkRecorder(t, recorder, 0, 0, []string{})
			}
		})
	}
}

func TestAssert_ConcurrentUsage_ShouldWorkCorrectly(t *testing.T) {
	t.Parallel()
	const goroutines = 8
	const iterations = 1000
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				assert.Equals(t, 1, 1)
				assert.NotEquals(t, 1, 2)
				assert.True(t, true)
				assert.False(t, false)
				assert.Nil(t, nil)
				assert.NotNil(t, 1)
				assert.NoError(t, nil)
				assert.Contains(t, "hello world", "world")
				assert.FloatEquals(t, 1.0, 1.0, 0.1)
				assert.Panic(t, func() { panic(errors.New("test")) })
			}
		})
	}
	waitGroup.Wait()
}
