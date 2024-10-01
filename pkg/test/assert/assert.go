package assert

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

func Equals(t Testing, expected any, actual any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if !reflect.DeepEqual(expected, actual) {
		tCtx.fail(fmt.Sprintf("Expected %+v to equal %+v.", expected, actual))
	}
}

func NotEquals(t Testing, expected any, actual any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if reflect.DeepEqual(expected, actual) {
		tCtx.fail(fmt.Sprintf("Expected arguments %+v to differ.", actual))
	}
}

func assertPanic(tCtx *testContext, panicFunc func(), msg *string, exact bool) {
	tCtx.Helper()

	panicOccurred := false
	gotRecoverMsg := false
	recoverMsg := ""

	wg := sync.WaitGroup{}
	wg.Add(1)
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
			wg.Done()
		}()
		panicFunc()
	}()
	wg.Wait()

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

func Panic(t Testing, panicFunc func(), options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, nil, false)
}

func PanicExact(t Testing, panicFunc func(), msg string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, &msg, true)
}

func PanicPart(t Testing, panicFunc func(), part string, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	assertPanic(tCtx, panicFunc, &part, false)
}

func Error(t Testing, err error, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err == nil {
		tCtx.fail("Expecting an error but none occurred.")
	}
}

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

func NoError(t Testing, err error, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if err != nil {
		tCtx.fail(fmt.Sprintf("Not expecting an error to occur. Got %s.", err.Error()))
	}
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Interface ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Chan ||
		v.Kind() == reflect.Func {
		if v.IsNil() {
			return true
		}
	}
	return false
}

func Nil(t Testing, value any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if !isNil(value) {
		tCtx.fail(fmt.Sprintf("Expecting nil value but value is %+v.", value))
	}
}

func NotNil(t Testing, value any, options ...Option) {
	tCtx := newTestContext(t, options...)
	tCtx.Helper()
	if isNil(value) {
		tCtx.fail("Expecting the value to not be nil.")
	}
}
