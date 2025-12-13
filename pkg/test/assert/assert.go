package assert

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Equals checks if two values are equal.
func Equals(t Testing, first any, second any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if !reflect.DeepEqual(first, second) {
		tCtx.fail(fmt.Sprintf("Expected %+v and %+v to be equal.", first, second))
	}
}

// NotEquals checks if two values are not equal.
func NotEquals(t Testing, first any, second any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if reflect.DeepEqual(first, second) {
		tCtx.fail(fmt.Sprintf("Expected %+v and %+v to differ.", first, second))
	}
}

// assertPanic checks if a function panics with an optional message.
func assertPanic(tCtx *testContext, panicFunc func(), msg *string, exact bool) {
	tCtx.Helper()

	panicOccurred := false
	gotRecoverMsg := false
	recoverMsg := ""

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
				if castErrStr, castErrStrOk := r.(string); castErrStrOk {
					gotRecoverMsg = true
					recoverMsg = castErrStr
				} else if castErr, castErrOk := r.(error); castErrOk {
					gotRecoverMsg = true
					recoverMsg = castErr.Error()
				}
			}
			waitGroup.Done()
		}()
		panicFunc()
	}()
	waitGroup.Wait()

	if !panicOccurred {
		tCtx.fail("Expected panic to occur but none occurred.")
		return
	}

	if msg != nil {
		if !gotRecoverMsg {
			tCtx.fail("Could not extract error message from panic.")
			return
		}
		if exact {
			if recoverMsg != *msg {
				tCtx.fail(fmt.Sprintf("Expected panic message to equal '%s' but got '%s'.", *msg, recoverMsg))
			}
		} else {
			if !strings.Contains(recoverMsg, *msg) {
				tCtx.fail(fmt.Sprintf("Expected panic message to contain '%s' but got '%s'.", *msg, recoverMsg))
			}
		}
	}
}

// Panic checks if a function panics.
func Panic(t Testing, panicFunc func(), options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, nil, false)
}

// PanicExact checks if a function panics with an exact message.
func PanicExact(t Testing, panicFunc func(), msg string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, &msg, true)
}

// PanicPart checks if a function panics with a message containing a part.
func PanicPart(t Testing, panicFunc func(), part string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, &part, false)
}

// Error checks if an error occurred.
func Error(t Testing, err error, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err == nil {
		tCtx.fail("Expecting an error but none occurred.")
	}
}

// ErrorExact checks if an error occurred with an exact message.
func ErrorExact(t Testing, err error, msg string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err == nil {
		tCtx.fail("Expecting an error but none occurred.")
		return
	}
	if msg != err.Error() {
		tCtx.fail(fmt.Sprintf("Expected the error message '%s' but got '%s'.", msg, err.Error()))
	}
}

// ErrorPart checks if an error occurred with a message containing a part.
func ErrorPart(t Testing, err error, part string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err == nil {
		tCtx.fail("Expecting an error but none occurred.")
		return
	}
	if !strings.Contains(err.Error(), part) {
		tCtx.fail(fmt.Sprintf("Expected the error message to contain '%s' but got '%s'.", part, err.Error()))
	}
}

// NoError checks if no error occurred.
func NoError(t Testing, err error, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err != nil {
		tCtx.fail(fmt.Sprintf("Not expecting an error to occur. Got %s.", err.Error()))
	}
}

// isNil checks if a value is nil.
func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() == reflect.Ptr ||
		reflectValue.Kind() == reflect.Interface ||
		reflectValue.Kind() == reflect.Slice ||
		reflectValue.Kind() == reflect.Map ||
		reflectValue.Kind() == reflect.Chan ||
		reflectValue.Kind() == reflect.Func {
		if reflectValue.IsNil() {
			return true
		}
	}
	return false
}

// Nil checks if a value is nil.
func Nil(t Testing, value any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if !isNil(value) {
		tCtx.fail(fmt.Sprintf("Expecting nil value but value is %+v.", value))
	}
}

// NotNil checks if a value is not nil.
func NotNil(t Testing, value any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if isNil(value) {
		tCtx.fail("Expecting the value to not be nil.")
	}
}

// True checks if a value is true.
func True(t Testing, value bool, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if !value {
		tCtx.fail("Expecting the value to be true.")
	}
}

// False checks if a value is false.
func False(t Testing, value bool, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if value {
		tCtx.fail("Expecting the value to be false.")
	}
}

// Contains checks if an interface contains the contents of another.
// If the value is a string, it checks for a substring.
func Contains(t Testing, value any, check any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()

	valueStr, valueIsString := value.(string)
	checkStr, checkIsString := check.(string)
	if valueIsString && checkIsString {
		if !strings.Contains(valueStr, checkStr) {
			tCtx.fail(fmt.Sprintf("Expecting '%v' to contain '%v'.", value, check))
		}
		return
	}

	tCtx.fail("Unknown types for the contains check.")
}

// FloatEquals checks if two floating point values are equal within a given epsilon.
func FloatEquals[T ~float64 | ~float32](t Testing, first T, second T, epsilon T, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	abs := first - second
	if abs < 0 {
		abs = -abs
	}
	if abs > epsilon {
		tCtx.fail(fmt.Sprintf("Expecting %f to equal %f within a margin of %f.", first, second, epsilon))
	}
}
