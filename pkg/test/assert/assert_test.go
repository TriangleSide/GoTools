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

func runAssertTest(t *testing.T, callback func(*testRecorder, ...assert.Option), expectLogs []string) {
	t.Helper()
	recorder := newTestRecorder(t)
	callback(recorder)
	if len(expectLogs) > 0 {
		checkRecorder(t, recorder, 0, 1, expectLogs)
	} else {
		checkRecorder(t, recorder, 0, 0, []string{})
	}
	recorder = newTestRecorder(t)
	callback(recorder, assert.Continue())
	if len(expectLogs) > 0 {
		checkRecorder(t, recorder, 1, 0, expectLogs)
	} else {
		checkRecorder(t, recorder, 0, 0, []string{})
	}
}

func TestEquals_EqualIntegers_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, 1, 1, opts...)
	}, []string{})
}

func TestEquals_DifferentIntegers_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, 1, 2, opts...)
	}, []string{"Expected 1 and 2 to be equal."})
}

func TestEquals_NilWithNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, nil, nil, opts...)
	}, []string{})
}

func TestEquals_NilWithNonNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, nil, 1, opts...)
	}, []string{"Expected <nil> and 1 to be equal."})
}

func TestEquals_DifferentTypes_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, 0, 0.0, opts...)
	}, []string{"Expected 0 and 0 to be equal."})
}

func TestEquals_EmptySlices_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Equals(tr, []int{}, []int{}, opts...)
	}, []string{})
}

func TestEquals_NilSlices_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var s1 []int
		var s2 []int
		assert.Equals(tr, s1, s2, opts...)
	}, []string{})
}

func TestEquals_NilSliceWithEmptySlice_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var s1 []int
		s2 := []int{}
		assert.Equals(tr, s1, s2, opts...)
	}, []string{"Expected [] and [] to be equal."})
}

func TestEquals_MapsWithSameContent_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"a": 1}
		assert.Equals(tr, m1, m2, opts...)
	}, []string{})
}

func TestEquals_NilMaps_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var m1 map[string]int
		var m2 map[string]int
		assert.Equals(tr, m1, m2, opts...)
	}, []string{})
}

func TestEquals_NilMapWithEmptyMap_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var m1 map[string]int
		m2 := map[string]int{}
		assert.Equals(tr, m1, m2, opts...)
	}, []string{"Expected map[] and map[] to be equal."})
}

func TestNotEquals_DifferentIntegers_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, 1, 2, opts...)
	}, []string{})
}

func TestNotEquals_EqualIntegers_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, 1, 1, opts...)
	}, []string{"Expected 1 and 1 to differ."})
}

func TestNotEquals_NilWithNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, nil, nil, opts...)
	}, []string{"Expected <nil> and <nil> to differ."})
}

func TestNotEquals_NilWithNonNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, nil, 1, opts...)
	}, []string{})
}

func TestNotEquals_ZeroValuesOfDifferentTypes_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, 0, 0.0, opts...)
	}, []string{})
}

func TestNotEquals_EmptySlices_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotEquals(tr, []int{}, []int{}, opts...)
	}, []string{"Expected [] and [] to differ."})
}

func TestPanic_FunctionPanics_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Panic(tr, func() { panic(errors.New("test panic")) }, opts...)
	}, []string{})
}

func TestPanic_FunctionDoesNotPanic_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Panic(tr, func() {}, opts...)
	}, []string{"Expected panic to occur but none occurred."})
}

func TestPanic_FunctionPanicsWithError_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Panic(tr, func() { panic(errors.New("42")) }, opts...)
	}, []string{})
}

func TestPanic_NilFunctionTriggersPanic_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var panicFunc func()
		assert.Panic(tr, panicFunc, opts...)
	}, []string{})
}

func TestPanicExact_PanicWithExactMessage_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicExact(tr, func() { panic(errors.New("exact message")) }, "exact message", opts...)
	}, []string{})
}

func TestPanicExact_PanicMessageDiffers_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicExact(tr, func() { panic(errors.New("wrong message")) }, "exact message", opts...)
	}, []string{"Expected panic message to equal 'exact message' but got 'wrong message'."})
}

func TestPanicExact_PanicWithErrorMessage_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicExact(tr, func() { panic(errors.New("42")) }, "42", opts...)
	}, []string{})
}

func TestPanicExact_PanicWithEmptyStringMessage_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicExact(tr, func() { panic(errors.New("")) }, "", opts...)
	}, []string{})
}

func TestPanicPart_PanicMessageContainsExpectedPart_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicPart(tr, func() { panic(errors.New("some panic message")) }, "panic", opts...)
	}, []string{})
}

func TestPanicPart_PanicMessageDoesNotContainExpectedPart_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicPart(tr, func() { panic(errors.New("wrong message")) }, "expected part", opts...)
	}, []string{"Expected panic message to contain 'expected part' but got 'wrong message'."})
}

func TestPanicPart_PanicMessageMissingPart_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicPart(tr, func() { panic(errors.New("other message")) }, "some part", opts...)
	}, []string{"Expected panic message to contain 'some part' but got 'other message'."})
}

func TestPanicPart_PanicWithErrorMessage_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.PanicPart(tr, func() { panic(errors.New("42")) }, "42", opts...)
	}, []string{})
}

func TestError_ErrorIsNotNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Error(tr, errors.New("some error"), opts...)
	}, []string{})
}

func TestError_ErrorIsNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Error(tr, nil, opts...)
	}, []string{"Expecting an error but none occurred."})
}

func TestError_ErrorInterfaceIsNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var err error
		assert.Error(tr, err, opts...)
	}, []string{"Expecting an error but none occurred."})
}

func TestErrorExact_ErrorMessageMatchesExactly_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.ErrorExact(tr, errors.New("exact error"), "exact error", opts...)
	}, []string{})
}

func TestErrorExact_ErrorMessageDiffers_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.ErrorExact(tr, errors.New("wrong error"), "exact error", opts...)
	}, []string{"Expected the error message 'exact error' but got 'wrong error'."})
}

func TestErrorExact_ErrorIsNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var err error
		assert.ErrorExact(tr, err, "some error", opts...)
	}, []string{"Expecting an error but none occurred."})
}

func TestErrorExact_ErrorMessageIsEmptyString_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		err := errors.New("")
		assert.ErrorExact(tr, err, "", opts...)
	}, []string{})
}

func TestErrorPart_ErrorMessageContainsPart_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.ErrorPart(tr, errors.New("some error occurred"), "error", opts...)
	}, []string{})
}

func TestErrorPart_ErrorMessageDoesNotContainPart_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.ErrorPart(tr, errors.New("wrong error"), "expected part", opts...)
	}, []string{"Expected the error message to contain 'expected part' but got 'wrong error'."})
}

func TestErrorPart_ErrorIsNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var err error
		assert.ErrorPart(tr, err, "some part", opts...)
	}, []string{"Expecting an error but none occurred."})
}

func TestNoError_ErrorIsNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NoError(tr, nil, opts...)
	}, []string{})
}

func TestNoError_ErrorIsNotNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NoError(tr, errors.New("unexpected error"), opts...)
	}, []string{"Not expecting an error to occur. Got unexpected error."})
}

func TestNoError_ErrorInterfaceIsNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var err error
		assert.NoError(tr, err, opts...)
	}, []string{})
}

func TestNoError_ErrorIsNonNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		err := errors.New("some error")
		assert.NoError(tr, err, opts...)
	}, []string{"Not expecting an error to occur. Got some error."})
}

func TestNil_ValueIsNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Nil(tr, nil, opts...)
	}, []string{})
}

func TestNil_ValueIsNilIntegerSlice_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value []int
		assert.Nil(tr, value, opts...)
	}, []string{})
}

func TestNil_ValueIsNilPointer_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value *int
		assert.Nil(tr, value, opts...)
	}, []string{})
}

func TestNil_ValueIsNilMap_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value map[string]int
		assert.Nil(tr, value, opts...)
	}, []string{})
}

func TestNil_ValueIsNilChan_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value chan int
		assert.Nil(tr, value, opts...)
	}, []string{})
}

func TestNil_ValueIsNilFunction_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value func()
		assert.Nil(tr, value, opts...)
	}, []string{})
}

func TestNil_ValueIsNotNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Nil(tr, "not nil", opts...)
	}, []string{"Expecting nil value but value is not nil."})
}

func TestNotNil_ValueIsNotNil_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotNil(tr, "not nil", opts...)
	}, []string{})
}

func TestNotNil_ValueIsNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.NotNil(tr, nil, opts...)
	}, []string{"Expecting the value to not be nil."})
}

func TestNotNil_ValueIsTypedNilPointer_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		var value *int
		assert.NotNil(tr, value, opts...)
	}, []string{"Expecting the value to not be nil."})
}

func TestTrue_ValueIsTrue_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.True(tr, true, opts...)
	}, []string{})
}

func TestTrue_ValueIsFalse_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.True(tr, false, opts...)
	}, []string{"Expecting the value to be true."})
}

func TestFalse_ValueIsFalse_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.False(tr, false, opts...)
	}, []string{})
}

func TestFalse_ValueIsTrue_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.False(tr, true, opts...)
	}, []string{"Expecting the value to be false."})
}

func TestContains_StringContainsSubstring_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Contains(tr, "test string", "test", opts...)
	}, []string{})
}

func TestContains_StringDoesNotContainSubstring_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Contains(tr, "test string", "not a substring", opts...)
	}, []string{"Expecting 'test string' to contain 'not a substring'."})
}

func TestContains_StringAndNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Contains(tr, "test string", nil, opts...)
	}, []string{"Unknown types for the contains check."})
}

func TestContains_NilAndString_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Contains(tr, nil, "test", opts...)
	}, []string{"Unknown types for the contains check."})
}

func TestContains_NilAndNil_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.Contains(tr, nil, nil, opts...)
	}, []string{"Unknown types for the contains check."})
}

func TestFloatEquals_ValuesWithinEpsilon_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.FloatEquals[float32](tr, 123.45, 123.46, 0.1, opts...)
	}, []string{})
}

func TestFloatEquals_DifferenceEqualsEpsilon_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.FloatEquals[float64](tr, 1.0, 1.5, 0.5, opts...)
	}, []string{})
}

func TestFloatEquals_ZeroEpsilonWithEqualValues_ShouldPass(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.FloatEquals[float64](tr, 1.0, 1.0, 0.0, opts...)
	}, []string{})
}

func TestFloatEquals_ZeroEpsilonWithDifferentValues_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.FloatEquals[float64](tr, 1.0, 2.0, 0.0, opts...)
	}, []string{"Expecting 1.", "to equal 2.", "within a margin of 0."})
}

func TestFloatEquals_ValuesOutsideEpsilon_ShouldFail(t *testing.T) {
	t.Parallel()
	runAssertTest(t, func(tr *testRecorder, opts ...assert.Option) {
		assert.FloatEquals[float32](tr, 123.45, 123.46, 0.0001, opts...)
	}, []string{"Expecting 123.", "to equal 123.", "within a margin of 0."})
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
