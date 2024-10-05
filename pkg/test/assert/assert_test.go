package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

type TestRecorder struct {
	helperCount int
	errorCount  int
	fatalCount  int
	logs        []string
}

func (tr *TestRecorder) Helper() {
	tr.helperCount++
}

func (tr *TestRecorder) Error(args ...any) {
	tr.errorCount++
	tr.logs = append(tr.logs, fmt.Sprint(args...))
}

func (tr *TestRecorder) Fatal(args ...any) {
	tr.fatalCount++
	tr.logs = append(tr.logs, fmt.Sprint(args...))
}

func newTestRecorder() *TestRecorder {
	return &TestRecorder{
		helperCount: 0,
		errorCount:  0,
		fatalCount:  0,
		logs:        make([]string, 0),
	}
}

func TestAssertFunctions(t *testing.T) {
	t.Parallel()

	checkRecorder := func(t *testing.T, tr *TestRecorder, helperCount int, errorCount int, fatalCount int, logs []string) {
		t.Helper()
		if tr.helperCount != helperCount {
			t.Fatalf("Incorrect helper count. Wanted %d but got %d.", helperCount, tr.helperCount)
		}
		if tr.errorCount != errorCount {
			t.Fatalf("Incorrect error count. Wanted %d but got %d.", errorCount, tr.errorCount)
		}
		if tr.fatalCount != fatalCount {
			t.Fatalf("Incorrect fatal count. Wanted %d but got %d.", fatalCount, tr.fatalCount)
		}
		if len(tr.logs) != len(logs) {
			t.Fatalf("Incorrect log count. Wanted %d but got %d.", len(logs), len(tr.logs))
		}
		for i, log := range logs {
			if tr.logs[i] != log {
				t.Fatalf("Incorrect error message. Wanted '%s' but got '%s'.", log, tr.logs[i])
			}
		}
	}

	t.Run("Equals positive case - Comparing equal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, 1, 1)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals negative case - Comparing unequal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, 1, 2)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected 1 to equal 2.",
		})
	})

	t.Run("Equals positive case - Comparing nil with nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, nil, nil)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals negative case - Comparing nil with non-nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, nil, 1)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected <nil> to equal 1.",
		})
	})

	t.Run("Equals negative case - Comparing different types", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, 0, 0.0)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected 0 to equal 0.",
		})
	})

	t.Run("Equals positive case - Comparing empty slices", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, []int{}, []int{})
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals positive case - Comparing nil slices", func(t *testing.T) {
		tr := newTestRecorder()
		var s1 []int
		var s2 []int
		assert.Equals(tr, s1, s2)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals negative case - Comparing nil slice with empty slice", func(t *testing.T) {
		tr := newTestRecorder()
		var s1 []int
		s2 := []int{}
		assert.Equals(tr, s1, s2)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected [] to equal [].",
		})
	})

	t.Run("Equals positive case - Comparing maps with same content", func(t *testing.T) {
		tr := newTestRecorder()
		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"a": 1}
		assert.Equals(tr, m1, m2)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals positive case - Comparing nil maps", func(t *testing.T) {
		tr := newTestRecorder()
		var m1 map[string]int
		var m2 map[string]int
		assert.Equals(tr, m1, m2)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals negative case - Comparing nil map with empty map", func(t *testing.T) {
		tr := newTestRecorder()
		var m1 map[string]int
		m2 := map[string]int{}
		assert.Equals(tr, m1, m2)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected map[] to equal map[].",
		})
	})

	t.Run("NotEquals positive case - Comparing different integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, 1, 2)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotEquals negative case - Comparing equal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, 1, 1)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected arguments 1 to differ.",
		})
	})

	t.Run("NotEquals negative case - Comparing nil with nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, nil, nil)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected arguments <nil> to differ.",
		})
	})

	t.Run("NotEquals positive case - Comparing nil with non-nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, nil, 1)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotEquals positive case - Comparing zero values of different types", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, 0, 0.0)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotEquals negative case - Comparing empty slices", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, []int{}, []int{})
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected arguments [] to differ.",
		})
	})

	t.Run("Panic positive case - Function panics as expected", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Panic(tr, func() { panic("test panic") })
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("Panic negative case - Function does not panic", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Panic(tr, func() {})
		checkRecorder(t, tr, 3, 0, 1, []string{
			"Expected panic to occur but none occurred.",
		})
	})

	t.Run("Panic positive case - Function panics with integer", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Panic(tr, func() { panic(42) })
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicExact positive case - Panic with exact message", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic("exact message") }, "exact message")
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicExact negative case - Panic message differs", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic("wrong message") }, "exact message")
		checkRecorder(t, tr, 3, 0, 1, []string{
			"Expected panic message to equal 'exact message' but got 'wrong message'.",
		})
	})

	t.Run("PanicExact negative case - Panic with non-string message", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic(42) }, "42")
		checkRecorder(t, tr, 3, 0, 1, []string{
			"Could not extract error message from panic.",
		})
	})

	t.Run("PanicExact positive case - Panic with empty string message", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic("") }, "")
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicPart positive case - Panic message contains expected part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicPart(tr, func() { panic(errors.New("some panic message")) }, "panic")
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicPart negative case - Panic message does not contain expected part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicPart(tr, func() { panic("wrong message") }, "expected part")
		checkRecorder(t, tr, 3, 0, 1, []string{
			"Expected panic message to contain 'expected part' but got 'wrong message'.",
		})
	})

	t.Run("PanicPart negative case - Panic message missing part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicPart(tr, func() { panic("other message") }, "some part")
		checkRecorder(t, tr, 3, 0, 1, []string{
			"Expected panic message to contain 'some part' but got 'other message'.",
		})
	})

	t.Run("Error positive case - Error is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Error(tr, errors.New("some error"))
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Error negative case - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Error(tr, nil)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting an error but none occurred.",
		})
	})

	t.Run("Error negative case - Error interface is nil", func(t *testing.T) {
		tr := newTestRecorder()
		var err error = nil
		assert.Error(tr, err)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting an error but none occurred.",
		})
	})

	t.Run("ErrorExact positive case - Error message matches exactly", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorExact(tr, errors.New("exact error"), "exact error")
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("ErrorExact negative case - Error message differs", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorExact(tr, errors.New("wrong error"), "exact error")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected the error message 'exact error' but got 'wrong error'.",
		})
	})

	t.Run("ErrorExact negative case - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		var err error = nil
		assert.ErrorExact(tr, err, "some error")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting an error but none occurred.",
		})
	})

	t.Run("ErrorExact positive case - Error message is empty string", func(t *testing.T) {
		tr := newTestRecorder()
		err := errors.New("")
		assert.ErrorExact(tr, err, "")
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("ErrorPart positive case - Error message contains part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorPart(tr, errors.New("some error occurred"), "error")
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("ErrorPart negative case - Error message does not contain part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorPart(tr, errors.New("wrong error"), "expected part")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expected the error message to contain 'expected part' but got 'wrong error'.",
		})
	})

	t.Run("ErrorPart negative case - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		var err error = nil
		assert.ErrorPart(tr, err, "some part")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting an error but none occurred.",
		})
	})

	t.Run("NoError positive case - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NoError(tr, nil)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NoError negative case - Error is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NoError(tr, errors.New("unexpected error"))
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Not expecting an error to occur. Got unexpected error.",
		})
	})

	t.Run("NoError positive case - Error interface is nil", func(t *testing.T) {
		tr := newTestRecorder()
		var err error = nil
		assert.NoError(tr, err)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NoError negative case - Error is non-nil", func(t *testing.T) {
		tr := newTestRecorder()
		err := errors.New("some error")
		assert.NoError(tr, err)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Not expecting an error to occur. Got some error.",
		})
	})

	t.Run("Nil positive case - Value is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Nil(tr, nil)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Nil negative case - Value is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Nil(tr, "not nil")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting nil value but value is not nil.",
		})
	})

	t.Run("Nil positive case - Nil pointer", func(t *testing.T) {
		tr := newTestRecorder()
		var p *int = nil
		assert.Nil(tr, p)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Nil negative case - Non-nil pointer", func(t *testing.T) {
		tr := newTestRecorder()
		var i int = 0
		var p *int = &i
		assert.Nil(tr, p)
		checkRecorder(t, tr, 2, 0, 1, []string{
			fmt.Sprintf("Expecting nil value but value is %v.", p),
		})
	})

	t.Run("Nil positive case - Nil slice", func(t *testing.T) {
		tr := newTestRecorder()
		var s []int = nil
		assert.Nil(tr, s)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Nil negative case - Empty slice", func(t *testing.T) {
		tr := newTestRecorder()
		s := []int{}
		assert.Nil(tr, s)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting nil value but value is [].",
		})
	})

	t.Run("Nil negative case - Zero value", func(t *testing.T) {
		tr := newTestRecorder()
		var i int = 0
		assert.Nil(tr, i)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting nil value but value is 0.",
		})
	})

	t.Run("NotNil positive case - Value is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotNil(tr, "not nil")
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotNil negative case - Value is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotNil(tr, nil)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting the value to not be nil.",
		})
	})

	t.Run("NotNil negative case - Nil slice", func(t *testing.T) {
		tr := newTestRecorder()
		var s []int = nil
		assert.NotNil(tr, s)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting the value to not be nil.",
		})
	})

	t.Run("NotNil positive case - Empty slice", func(t *testing.T) {
		tr := newTestRecorder()
		s := []int{}
		assert.NotNil(tr, s)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotNil positive case - Zero value", func(t *testing.T) {
		tr := newTestRecorder()
		var i int = 0
		assert.NotNil(tr, i)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("True positive case", func(t *testing.T) {
		tr := newTestRecorder()
		assert.True(tr, true)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("True negative case", func(t *testing.T) {
		tr := newTestRecorder()
		assert.True(tr, false)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting the value to be true.",
		})
	})

	t.Run("False positive case", func(t *testing.T) {
		tr := newTestRecorder()
		assert.False(tr, false)
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("False negative case", func(t *testing.T) {
		tr := newTestRecorder()
		assert.False(tr, true)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting the value to be false.",
		})
	})

	t.Run("Contains positive case - String and string", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Contains(tr, "test string", "test")
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Contains negative case - String and string", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Contains(tr, "test string", "not a substring")
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Expecting 'test string' to contain 'not a substring'.",
		})
	})

	t.Run("Contains negative case - nil and nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Contains(tr, nil, nil)
		checkRecorder(t, tr, 2, 0, 1, []string{
			"Unknown types for the contains check.",
		})
	})

	t.Run("Equals positive case with Continue - Comparing equal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, 1, 1, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Equals negative case with Continue - Comparing unequal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Equals(tr, 1, 2, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expected 1 to equal 2.",
		})
	})

	t.Run("NotEquals positive case with Continue - Comparing different integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, 1, 2, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotEquals negative case with Continue - Comparing equal integers", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotEquals(tr, 1, 1, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expected arguments 1 to differ.",
		})
	})

	t.Run("Panic positive case with Continue - Function panics as expected", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Panic(tr, func() { panic("test panic") }, assert.Continue())
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("Panic negative case with Continue - Function does not panic", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Panic(tr, func() {}, assert.Continue())
		checkRecorder(t, tr, 3, 1, 0, []string{
			"Expected panic to occur but none occurred.",
		})
	})

	t.Run("PanicExact positive case with Continue - Panic with exact message", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic("exact message") }, "exact message", assert.Continue())
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicExact negative case with Continue - Panic message differs", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicExact(tr, func() { panic("wrong message") }, "exact message", assert.Continue())
		checkRecorder(t, tr, 3, 1, 0, []string{
			"Expected panic message to equal 'exact message' but got 'wrong message'.",
		})
	})

	t.Run("PanicPart positive case with Continue - Panic message contains expected part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicPart(tr, func() { panic("some panic message") }, "panic", assert.Continue())
		checkRecorder(t, tr, 2, 0, 0, []string{})
	})

	t.Run("PanicPart negative case with Continue - Panic message does not contain expected part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.PanicPart(tr, func() { panic("wrong message") }, "expected part", assert.Continue())
		checkRecorder(t, tr, 3, 1, 0, []string{
			"Expected panic message to contain 'expected part' but got 'wrong message'.",
		})
	})

	t.Run("Error positive case with Continue - Error is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Error(tr, errors.New("some error"), assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Error negative case with Continue - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Error(tr, nil, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting an error but none occurred.",
		})
	})

	t.Run("ErrorExact positive case with Continue - Error message matches exactly", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorExact(tr, errors.New("exact error"), "exact error", assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("ErrorExact negative case with Continue - Error message differs", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorExact(tr, errors.New("wrong error"), "exact error", assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expected the error message 'exact error' but got 'wrong error'.",
		})
	})

	t.Run("ErrorPart positive case with Continue - Error message contains part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorPart(tr, errors.New("some error occurred"), "error", assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("ErrorPart negative case with Continue - Error message does not contain part", func(t *testing.T) {
		tr := newTestRecorder()
		assert.ErrorPart(tr, errors.New("wrong error"), "expected part", assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expected the error message to contain 'expected part' but got 'wrong error'.",
		})
	})

	t.Run("NoError positive case with Continue - Error is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NoError(tr, nil, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NoError negative case with Continue - Error is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NoError(tr, errors.New("unexpected error"), assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Not expecting an error to occur. Got unexpected error.",
		})
	})

	t.Run("Nil positive case with Continue - Value is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Nil(tr, nil, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Nil negative case with Continue - Value is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Nil(tr, "not nil", assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting nil value but value is not nil.",
		})
	})

	t.Run("NotNil positive case with Continue - Value is not nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotNil(tr, "not nil", assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("NotNil negative case with Continue - Value is nil", func(t *testing.T) {
		tr := newTestRecorder()
		assert.NotNil(tr, nil, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting the value to not be nil.",
		})
	})

	t.Run("True positive case with Continue", func(t *testing.T) {
		tr := newTestRecorder()
		assert.True(tr, true, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("True negative case with Continue", func(t *testing.T) {
		tr := newTestRecorder()
		assert.True(tr, false, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting the value to be true.",
		})
	})

	t.Run("False positive case with Continue", func(t *testing.T) {
		tr := newTestRecorder()
		assert.False(tr, false, assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("False negative case with Continue", func(t *testing.T) {
		tr := newTestRecorder()
		assert.False(tr, true, assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting the value to be false.",
		})
	})

	t.Run("Contains positive case with Continue - String and string", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Contains(tr, "test string", "test", assert.Continue())
		checkRecorder(t, tr, 1, 0, 0, []string{})
	})

	t.Run("Contains negative case with Continue - String and string", func(t *testing.T) {
		tr := newTestRecorder()
		assert.Contains(tr, "test string", "not a substring", assert.Continue())
		checkRecorder(t, tr, 2, 1, 0, []string{
			"Expecting 'test string' to contain 'not a substring'.",
		})
	})
}
