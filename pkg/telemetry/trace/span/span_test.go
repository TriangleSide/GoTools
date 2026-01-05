package span_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/event"
	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/span"
	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/status"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestNew_NilParent_CreatesRootSpan(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test-span", "trace-123", nil)
	assert.NotNil(t, testSpan)
	assert.Equals(t, "test-span", testSpan.Name())
	assert.Equals(t, "trace-123", testSpan.TraceID())
	assert.Nil(t, testSpan.Parent())
	assert.Equals(t, 0, len(testSpan.Children()))
	assert.Equals(t, 0, len(testSpan.Attributes()))
	assert.Equals(t, 0, len(testSpan.Events()))
	assert.Equals(t, status.Unset, testSpan.StatusCode())
}

func TestNew_WithParent_CreatesChildSpan(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", "", nil)
	child := span.New("child", "", parent)
	assert.Equals(t, parent, child.Parent())
	assert.Equals(t, 1, len(parent.Children()))
	assert.Equals(t, child, parent.Children()[0])
}

func TestNew_MultipleChildren_AllAddedToParent(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", "", nil)
	child1 := span.New("child1", "", parent)
	child2 := span.New("child2", "", parent)
	child3 := span.New("child3", "", parent)
	children := parent.Children()
	assert.Equals(t, 3, len(children))
	assert.Equals(t, child1, children[0])
	assert.Equals(t, child2, children[1])
	assert.Equals(t, child3, children[2])
}

func TestNew_NestedSpans_CreatesHierarchy(t *testing.T) {
	t.Parallel()
	root := span.New("root", "", nil)
	child := span.New("child", "", root)
	grandchild := span.New("grandchild", "", child)
	assert.Nil(t, root.Parent())
	assert.Equals(t, root, child.Parent())
	assert.Equals(t, child, grandchild.Parent())
	assert.Equals(t, 1, len(root.Children()))
	assert.Equals(t, 1, len(child.Children()))
	assert.Equals(t, 0, len(grandchild.Children()))
}

func TestNew_RecordsStartTime(t *testing.T) {
	t.Parallel()
	before := time.Now()
	s := span.New("test", "", nil)
	after := time.Now()
	assert.True(t, !s.StartTime().Before(before))
	assert.True(t, !s.StartTime().After(after))
}

func TestNew_ConcurrentChildCreation_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	parent := span.New("parent", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.New("child", "", parent)
			}
		})
	}
	waitGroup.Wait()
	children := parent.Children()
	assert.Equals(t, goroutines*iterations, len(children))
}

func TestSpanEnd_RecordsEndTime(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", "", nil)
	assert.True(t, testSpan.EndTime().IsZero())
	before := time.Now()
	testSpan.End()
	after := time.Now()
	assert.False(t, testSpan.EndTime().IsZero())
	assert.True(t, !testSpan.EndTime().Before(before))
	assert.True(t, !testSpan.EndTime().After(after))
}

func TestSpanDuration_BeforeEnd_ReturnsDurationSinceStart(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	time.Sleep(10 * time.Millisecond)
	duration := s.Duration()
	assert.True(t, duration >= 10*time.Millisecond)
}

func TestSpanDuration_AfterEnd_ReturnsFixedDuration(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	time.Sleep(10 * time.Millisecond)
	s.End()
	duration1 := s.Duration()
	time.Sleep(10 * time.Millisecond)
	duration2 := s.Duration()
	assert.Equals(t, duration1, duration2)
}

func TestSpanChildren_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", "", nil)
	span.New("child", "", parent)
	children1 := parent.Children()
	children2 := parent.Children()
	assert.Equals(t, len(children1), len(children2))
	children1[0] = nil
	assert.NotNil(t, parent.Children()[0])
}

func TestSpanEnd_ConcurrentCalls_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			testSpan.End()
		})
	}
	waitGroup.Wait()
	assert.False(t, testSpan.EndTime().IsZero())
}

func TestSpanDuration_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	s := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				_ = s.Duration()
			}
		})
	}
	waitGroup.Wait()
}

func TestSpanChildren_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	parent := span.New("parent", "", nil)
	for range 5 {
		span.New("child", "", parent)
	}
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				children := parent.Children()
				assert.Equals(t, 5, len(children), assert.Continue())
			}
		})
	}
	waitGroup.Wait()
}

func TestSpan_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	parent := span.New("parent", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.New("child", "", parent)
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = parent.Children()
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = parent.Duration()
			}
		})
	}
	waitGroup.Wait()
	children := parent.Children()
	assert.Equals(t, goroutines*iterations, len(children))
}

func TestSetAttributes_SingleAttribute_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.SetAttributes(attribute.String("key", "value"))
	attrs := s.Attributes()
	assert.Equals(t, 1, len(attrs))
	assert.Equals(t, "key", attrs[0].Key())
	assert.Equals(t, "value", attrs[0].StringValue())
}

func TestSetAttributes_MultipleTypes_AllSupported(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", "", nil)
	testSpan.SetAttributes(
		attribute.String("string", "hello"),
		attribute.Int("int", 42),
		attribute.Float("float", 3.14),
		attribute.Bool("bool", true),
	)
	attrs := testSpan.Attributes()
	assert.Equals(t, 4, len(attrs))
	assert.Equals(t, "string", attrs[0].Key())
	assert.Equals(t, "hello", attrs[0].StringValue())
	assert.Equals(t, "int", attrs[1].Key())
	assert.Equals(t, int64(42), attrs[1].IntValue())
	assert.Equals(t, "float", attrs[2].Key())
	assert.Equals(t, 3.14, attrs[2].FloatValue())
	assert.Equals(t, "bool", attrs[3].Key())
	assert.Equals(t, true, attrs[3].BoolValue())
}

func TestAttributes_NoAttributes_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	attrs := s.Attributes()
	assert.Equals(t, 0, len(attrs))
}

func TestAttributes_MultipleAttributes_ReturnsAll(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", "", nil)
	testSpan.SetAttributes(
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
		attribute.String("key3", "value3"),
	)
	attrs := testSpan.Attributes()
	assert.Equals(t, 3, len(attrs))
	assert.Equals(t, "key1", attrs[0].Key())
	assert.Equals(t, "value1", attrs[0].StringValue())
	assert.Equals(t, "key2", attrs[1].Key())
	assert.Equals(t, "value2", attrs[1].StringValue())
	assert.Equals(t, "key3", attrs[2].Key())
	assert.Equals(t, "value3", attrs[2].StringValue())
}

func TestAttributes_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.SetAttributes(attribute.String("key", "value"))
	attrs := s.Attributes()
	attrs[0] = attribute.String("key", "modified")
	originalAttrs := s.Attributes()
	assert.Equals(t, 1, len(originalAttrs))
	assert.Equals(t, "value", originalAttrs[0].StringValue())
}

func TestSetAttributes_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for i := range goroutines {
		waitGroup.Go(func() {
			for j := range iterations {
				key := "key" + string(rune('A'+i))
				testSpan.SetAttributes(attribute.Int(key, int64(j)))
			}
		})
	}
	waitGroup.Wait()
	attrs := testSpan.Attributes()
	assert.Equals(t, goroutines*iterations, len(attrs))
}

func TestAttributes_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for i := range goroutines {
		waitGroup.Go(func() {
			for j := range iterations {
				key := "key" + string(rune('A'+i))
				testSpan.SetAttributes(attribute.Int(key, int64(j)))
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = testSpan.Attributes()
			}
		})
	}
	waitGroup.Wait()
}

func TestAddEvent_SingleEvent_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	e := event.New("test-event")
	s.AddEvent(e)
	events := s.Events()
	assert.Equals(t, 1, len(events))
	assert.Equals(t, "test-event", events[0].Name())
}

func TestAddEvent_MultipleEvents_AllRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.AddEvent(event.New("event1"))
	s.AddEvent(event.New("event2"))
	s.AddEvent(event.New("event3"))
	events := s.Events()
	assert.Equals(t, 3, len(events))
	assert.Equals(t, "event1", events[0].Name())
	assert.Equals(t, "event2", events[1].Name())
	assert.Equals(t, "event3", events[2].Name())
}

func TestEvents_NoEvents_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	events := s.Events()
	assert.Equals(t, 0, len(events))
}

func TestEvents_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.AddEvent(event.New("original"))
	events := s.Events()
	events[0] = event.New("modified")
	originalEvents := s.Events()
	assert.Equals(t, "original", originalEvents[0].Name())
}

func TestAddEvent_WithAttributes_PreservesAttributes(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", "", nil)
	e := event.New("test-event",
		attribute.String("key", "value"),
		attribute.Int("count", 42),
	)
	testSpan.AddEvent(e)
	events := testSpan.Events()
	attrs := events[0].Attributes()
	assert.Equals(t, 2, len(attrs))
	assert.Equals(t, "key", attrs[0].Key())
	assert.Equals(t, "value", attrs[0].StringValue())
}

func TestAddEvent_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.AddEvent(event.New("concurrent-event"))
			}
		})
	}
	waitGroup.Wait()
	events := testSpan.Events()
	assert.Equals(t, goroutines*iterations, len(events))
}

func TestEvents_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.AddEvent(event.New("concurrent-event"))
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = testSpan.Events()
			}
		})
	}
	waitGroup.Wait()
}

func TestStatus_DefaultValue_IsUnset(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	assert.Equals(t, status.Unset, s.StatusCode())
}

func TestSetStatus_Error_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, s.StatusCode())
}

func TestSetStatus_Success_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, s.StatusCode())
}

func TestSetStatus_CanBeOverwritten(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, s.StatusCode())
	s.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, s.StatusCode())
}

func TestSetStatus_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.SetStatusCode(status.Success)
			}
		})
	}
	waitGroup.Wait()
	code := testSpan.StatusCode()
	assert.True(t, code == status.Success || code == status.Unset || code == status.Error)
}

func TestStatus_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.SetStatusCode(status.Error)
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = testSpan.StatusCode()
			}
		})
	}
	waitGroup.Wait()
}

func TestTraceID_EmptyString_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	assert.Equals(t, "", s.TraceID())
}

func TestTraceID_NonEmptyString_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", "abc-123-def", nil)
	assert.Equals(t, "abc-123-def", s.TraceID())
}

func TestSpanID_RootSpan_ReturnsZero(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	assert.Equals(t, "0", s.SpanID())
}

func TestSpanID_ChildSpan_ReturnsIncrementingID(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", "", nil)
	child1 := span.New("child1", "", parent)
	child2 := span.New("child2", "", parent)
	child3 := span.New("child3", "", parent)
	assert.Equals(t, "0", parent.SpanID())
	assert.Equals(t, "1", child1.SpanID())
	assert.Equals(t, "2", child2.SpanID())
	assert.Equals(t, "3", child3.SpanID())
}

func TestSpanID_NestedSpans_ReturnsIncrementingID(t *testing.T) {
	t.Parallel()
	root := span.New("root", "", nil)
	child := span.New("child", "", root)
	grandchild := span.New("grandchild", "", child)
	assert.Equals(t, "0", root.SpanID())
	assert.Equals(t, "1", child.SpanID())
	assert.Equals(t, "2", grandchild.SpanID())
}

func TestSpanID_SeparateHierarchies_HaveIndependentIDs(t *testing.T) {
	t.Parallel()
	root1 := span.New("root1", "", nil)
	child1 := span.New("child1", "", root1)
	root2 := span.New("root2", "", nil)
	child2 := span.New("child2", "", root2)
	assert.Equals(t, "0", root1.SpanID())
	assert.Equals(t, "1", child1.SpanID())
	assert.Equals(t, "0", root2.SpanID())
	assert.Equals(t, "1", child2.SpanID())
}

func TestSpanID_ConcurrentChildCreation_AssignsUniqueIDs(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	parent := span.New("parent", "", nil)
	children := make([]*span.Span, 0, goroutines*iterations)
	var childrenMutex sync.Mutex
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				child := span.New("child", "", parent)
				childrenMutex.Lock()
				children = append(children, child)
				childrenMutex.Unlock()
			}
		})
	}
	waitGroup.Wait()
	ids := make(map[string]bool)
	ids["0"] = true
	for _, child := range children {
		id := child.SpanID()
		assert.False(t, ids[id])
		ids[id] = true
	}
	assert.Equals(t, goroutines*iterations+1, len(ids))
}

func TestRecordError_StandardError_SetsStatusAndAddsEvent(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	testErr := errors.New("test error")
	s.RecordError(testErr)
	assert.Equals(t, status.Error, s.StatusCode())
	events := s.Events()
	assert.Equals(t, 1, len(events))
	assert.Equals(t, "error", events[0].Name())
	attrs := events[0].Attributes()
	assert.Equals(t, 2, len(attrs))
	assert.Equals(t, "error.message", attrs[0].Key())
	assert.Equals(t, "test error", attrs[0].StringValue())
	assert.Equals(t, "error.type", attrs[1].Key())
	assert.Equals(t, "*errors.errorString", attrs[1].StringValue())
}

func TestRecordError_NilError_IsNoOp(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.RecordError(nil)
	assert.Equals(t, status.Unset, s.StatusCode())
	events := s.Events()
	assert.Equals(t, 0, len(events))
}

func TestRecordError_WrappedError_RecordsWrappedMessage(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	innerErr := errors.New("inner error")
	wrappedErr := fmt.Errorf("outer context: %w", innerErr)
	s.RecordError(wrappedErr)
	assert.Equals(t, status.Error, s.StatusCode())
	events := s.Events()
	assert.Equals(t, 1, len(events))
	attrs := events[0].Attributes()
	assert.Equals(t, "error.message", attrs[0].Key())
	assert.Equals(t, "outer context: inner error", attrs[0].StringValue())
	assert.Equals(t, "error.type", attrs[1].Key())
	assert.Equals(t, "*fmt.wrapError", attrs[1].StringValue())
}

type customError struct {
	code    int
	message string
}

func (e *customError) Error() string {
	return fmt.Sprintf("error %d: %s", e.code, e.message)
}

func TestRecordError_CustomErrorType_RecordsTypeInfo(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	customErr := &customError{code: 404, message: "not found"}
	s.RecordError(customErr)
	assert.Equals(t, status.Error, s.StatusCode())
	events := s.Events()
	assert.Equals(t, 1, len(events))
	attrs := events[0].Attributes()
	assert.Equals(t, "error.message", attrs[0].Key())
	assert.Equals(t, "error 404: not found", attrs[0].StringValue())
	assert.Equals(t, "error.type", attrs[1].Key())
	assert.Equals(t, "*span_test.customError", attrs[1].StringValue())
}

func TestRecordError_MultipleErrors_AllRecorded(t *testing.T) {
	t.Parallel()
	s := span.New("test", "", nil)
	s.RecordError(errors.New("first error"))
	s.RecordError(errors.New("second error"))
	s.RecordError(errors.New("third error"))
	assert.Equals(t, status.Error, s.StatusCode())
	events := s.Events()
	assert.Equals(t, 3, len(events))
	attrs0 := events[0].Attributes()
	assert.Equals(t, "first error", attrs0[0].StringValue())
	attrs1 := events[1].Attributes()
	assert.Equals(t, "second error", attrs1[0].StringValue())
	attrs2 := events[2].Attributes()
	assert.Equals(t, "third error", attrs2[0].StringValue())
}

func TestRecordError_ConcurrentCalls_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.RecordError(errors.New("concurrent error"))
			}
		})
	}
	waitGroup.Wait()
	assert.Equals(t, status.Error, testSpan.StatusCode())
	events := testSpan.Events()
	assert.Equals(t, goroutines*iterations, len(events))
}

func TestRecordError_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	testSpan := span.New("test", "", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				testSpan.RecordError(errors.New("concurrent error"))
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = testSpan.StatusCode()
				_ = testSpan.Events()
			}
		})
	}
	waitGroup.Wait()
}

func TestWithEndCallback_OnEnd_CallbackInvoked(t *testing.T) {
	t.Parallel()
	var callbackSpan *span.Span
	callbackInvoked := false
	testSpan := span.New("test", "trace-123", nil, span.WithEndCallback(func(s *span.Span) {
		callbackInvoked = true
		callbackSpan = s
	}))
	assert.False(t, callbackInvoked)
	testSpan.End()
	assert.True(t, callbackInvoked)
	assert.Equals(t, testSpan, callbackSpan)
}

func TestWithEndCallback_MultipleEndCalls_CallbackInvokedOnce(t *testing.T) {
	t.Parallel()
	callCount := 0
	testSpan := span.New("test", "trace-123", nil, span.WithEndCallback(func(_ *span.Span) {
		callCount++
	}))
	testSpan.End()
	testSpan.End()
	testSpan.End()
	assert.Equals(t, 1, callCount)
}

func TestWithEndCallback_NilCallback_NoError(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", "trace-123", nil)
	testSpan.End()
	assert.False(t, testSpan.EndTime().IsZero())
}

func TestNew_WithOptions_DoesNotAffectOtherBehavior(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test-span", "trace-123", nil, span.WithEndCallback(func(_ *span.Span) {}))
	assert.NotNil(t, testSpan)
	assert.Equals(t, "test-span", testSpan.Name())
	assert.Equals(t, "trace-123", testSpan.TraceID())
	assert.Nil(t, testSpan.Parent())
	assert.Equals(t, 0, len(testSpan.Children()))
	assert.Equals(t, 0, len(testSpan.Attributes()))
	assert.Equals(t, 0, len(testSpan.Events()))
	assert.Equals(t, status.Unset, testSpan.StatusCode())
}

func TestWithEndCallback_ChildSpan_CallbackInvoked(t *testing.T) {
	t.Parallel()
	callbackInvoked := false
	parent := span.New("parent", "trace-123", nil)
	child := span.New("child", "trace-123", parent, span.WithEndCallback(func(_ *span.Span) {
		callbackInvoked = true
	}))
	child.End()
	assert.True(t, callbackInvoked)
}

func TestWithEndCallback_ConcurrentCalls_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	var callCount int64
	var mu sync.Mutex
	testSpan := span.New("test", "trace-123", nil, span.WithEndCallback(func(_ *span.Span) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}))
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			testSpan.End()
		})
	}
	waitGroup.Wait()
	assert.Equals(t, int64(1), callCount)
}
