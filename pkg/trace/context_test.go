package trace_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/trace"
	"github.com/TriangleSide/GoTools/pkg/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/trace/event"
	"github.com/TriangleSide/GoTools/pkg/trace/span"
	"github.com/TriangleSide/GoTools/pkg/trace/status"
)

type mockExporter struct {
	mu            sync.Mutex
	exportedSpans []*span.Span
}

func (m *mockExporter) Export(_ context.Context, s *span.Span) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exportedSpans = append(m.exportedSpans, s)
}

func TestStart_EmptyContext_CreatesRootSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resultCtx, span := trace.Start(ctx, t.Name())
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, span)
	assert.Equals(t, t.Name(), span.Name())
	assert.Nil(t, span.Parent())
}

func TestStart_WithParent_CreatesChildSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	_, child := trace.Start(ctx, "child")
	assert.Equals(t, parent, child.Parent())
	children := parent.Children()
	assert.Equals(t, 1, len(children))
	assert.Equals(t, child, children[0])
}

func TestStart_MultipleChildren_AllAddedToParent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	_, child1 := trace.Start(ctx, "child1")
	_, child2 := trace.Start(ctx, "child2")
	_, child3 := trace.Start(ctx, "child3")
	children := parent.Children()
	assert.Equals(t, 3, len(children))
	assert.Equals(t, child1, children[0])
	assert.Equals(t, child2, children[1])
	assert.Equals(t, child3, children[2])
}

func TestStart_NestedSpans_CreatesHierarchy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, root := trace.Start(ctx, "root")
	ctx, child := trace.Start(ctx, "child")
	_, grandchild := trace.Start(ctx, "grandchild")
	assert.Nil(t, root.Parent())
	assert.Equals(t, root, child.Parent())
	assert.Equals(t, child, grandchild.Parent())
	assert.Equals(t, 1, len(root.Children()))
	assert.Equals(t, 1, len(child.Children()))
	assert.Equals(t, 0, len(grandchild.Children()))
}

func TestStart_RecordsStartTime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	before := time.Now()
	_, span := trace.Start(ctx, "test")
	after := time.Now()
	assert.True(t, !span.StartTime().Before(before))
	assert.True(t, !span.StartTime().After(after))
}

func TestStart_SpanEnd_RecordsEndTime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	assert.True(t, span.EndTime().IsZero())
	before := time.Now()
	span.End()
	after := time.Now()
	assert.False(t, span.EndTime().IsZero())
	assert.True(t, !span.EndTime().Before(before))
	assert.True(t, !span.EndTime().After(after))
}

func TestStart_SpanDuration_BeforeEnd_ReturnsDurationSinceStart(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	time.Sleep(10 * time.Millisecond)
	duration := span.Duration()
	assert.True(t, duration >= 10*time.Millisecond)
}

func TestStart_SpanDuration_AfterEnd_ReturnsFixedDuration(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	time.Sleep(10 * time.Millisecond)
	span.End()
	duration1 := span.Duration()
	time.Sleep(10 * time.Millisecond)
	duration2 := span.Duration()
	assert.Equals(t, duration1, duration2)
}

func TestStart_SpanChildren_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	trace.Start(ctx, "child")
	children1 := parent.Children()
	children2 := parent.Children()
	assert.Equals(t, len(children1), len(children2))
	children1[0] = nil
	assert.NotNil(t, parent.Children()[0])
}

func TestStart_ConcurrentChildCreation_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				trace.Start(ctx, "child")
			}
		})
	}
	waitGroup.Wait()
	children := parent.Children()
	assert.Equals(t, goroutines*iterations, len(children))
}

func TestStart_SpanEnd_ConcurrentCalls_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			span.End()
		})
	}
	waitGroup.Wait()
	assert.False(t, span.EndTime().IsZero())
}

func TestStart_SpanDuration_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				_ = span.Duration()
			}
		})
	}
	waitGroup.Wait()
}

func TestStart_SpanChildren_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	for range 5 {
		trace.Start(ctx, "child")
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

func TestStart_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				trace.Start(ctx, "child")
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

func TestStart_SetAttributes_SingleAttribute_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttributes(attribute.String("key", "value"))
	attrs := span.Attributes()
	assert.Equals(t, 1, len(attrs))
	assert.Equals(t, "key", attrs[0].Key())
	assert.Equals(t, "value", attrs[0].StringValue())
}

func TestStart_SetAttributes_MultipleTypes_AllSupported(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttributes(
		attribute.String("string", "hello"),
		attribute.Int("int", 42),
		attribute.Float("float", 3.14),
		attribute.Bool("bool", true),
	)
	attrs := span.Attributes()
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

func TestStart_Attributes_NoAttributes_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	attrs := span.Attributes()
	assert.Equals(t, 0, len(attrs))
}

func TestStart_Attributes_MultipleAttributes_ReturnsAll(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttributes(
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
		attribute.String("key3", "value3"),
	)
	attrs := span.Attributes()
	assert.Equals(t, 3, len(attrs))
	assert.Equals(t, "key1", attrs[0].Key())
	assert.Equals(t, "value1", attrs[0].StringValue())
	assert.Equals(t, "key2", attrs[1].Key())
	assert.Equals(t, "value2", attrs[1].StringValue())
	assert.Equals(t, "key3", attrs[2].Key())
	assert.Equals(t, "value3", attrs[2].StringValue())
}

func TestStart_Attributes_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttributes(attribute.String("key", "value"))
	attrs := span.Attributes()
	attrs[0] = attribute.String("key", "modified")
	originalAttrs := span.Attributes()
	assert.Equals(t, 1, len(originalAttrs))
	assert.Equals(t, "value", originalAttrs[0].StringValue())
}

func TestStart_SetAttributes_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for i := range goroutines {
		waitGroup.Go(func() {
			for j := range iterations {
				key := "key" + string(rune('A'+i))
				span.SetAttributes(attribute.Int(key, int64(j)))
			}
		})
	}
	waitGroup.Wait()
	attrs := span.Attributes()
	assert.Equals(t, goroutines*iterations, len(attrs))
}

func TestStart_Attributes_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for i := range goroutines {
		waitGroup.Go(func() {
			for j := range iterations {
				key := "key" + string(rune('A'+i))
				span.SetAttributes(attribute.Int(key, int64(j)))
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = span.Attributes()
			}
		})
	}
	waitGroup.Wait()
}

func TestStart_AddEvent_SingleEvent_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	evt := event.New("test-event")
	span.AddEvent(evt)
	events := span.Events()
	assert.Equals(t, 1, len(events))
	assert.Equals(t, "test-event", events[0].Name())
}

func TestStart_AddEvent_MultipleEvents_AllRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.AddEvent(event.New("event1"))
	span.AddEvent(event.New("event2"))
	span.AddEvent(event.New("event3"))
	events := span.Events()
	assert.Equals(t, 3, len(events))
	assert.Equals(t, "event1", events[0].Name())
	assert.Equals(t, "event2", events[1].Name())
	assert.Equals(t, "event3", events[2].Name())
}

func TestStart_Events_NoEvents_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	events := span.Events()
	assert.Equals(t, 0, len(events))
}

func TestStart_Events_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.AddEvent(event.New("original"))
	events := span.Events()
	events[0] = event.New("modified")
	originalEvents := span.Events()
	assert.Equals(t, "original", originalEvents[0].Name())
}

func TestStart_AddEvent_WithAttributes_PreservesAttributes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	evt := event.New("test-event",
		attribute.String("key", "value"),
		attribute.Int("count", 42),
	)
	span.AddEvent(evt)
	events := span.Events()
	attrs := events[0].Attributes()
	assert.Equals(t, 2, len(attrs))
	assert.Equals(t, "key", attrs[0].Key())
	assert.Equals(t, "value", attrs[0].StringValue())
}

func TestStart_AddEvent_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.AddEvent(event.New("concurrent-event"))
			}
		})
	}
	waitGroup.Wait()
	events := span.Events()
	assert.Equals(t, goroutines*iterations, len(events))
}

func TestStart_Events_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.AddEvent(event.New("concurrent-event"))
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = span.Events()
			}
		})
	}
	waitGroup.Wait()
}

func TestStart_Status_DefaultValue_IsUnset(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	assert.Equals(t, status.Unset, span.StatusCode())
}

func TestStart_SetStatus_Error_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, span.StatusCode())
}

func TestStart_SetStatus_Success_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, span.StatusCode())
}

func TestStart_SetStatus_CanBeOverwritten(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, span.StatusCode())
	span.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, span.StatusCode())
}

func TestStart_SetStatus_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.SetStatusCode(status.Success)
			}
		})
	}
	waitGroup.Wait()
	code := span.StatusCode()
	assert.True(t, code == status.Success || code == status.Unset || code == status.Error)
}

func TestStart_Status_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 50
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.SetStatusCode(status.Error)
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = span.StatusCode()
			}
		})
	}
	waitGroup.Wait()
}

func TestSetTraceID_PassedToSpan_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx = trace.SetTraceID(ctx, "abc123")
	_, span := trace.Start(ctx, "test")
	assert.Equals(t, "abc123", span.TraceID())
}

func TestStart_NoTraceIDSet_SpanHasEmptyTraceID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	assert.Equals(t, "", span.TraceID())
}

func TestSetTraceID_Overwrite_LatestPassedToSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx = trace.SetTraceID(ctx, "first")
	ctx = trace.SetTraceID(ctx, "second")
	_, span := trace.Start(ctx, "test")
	assert.Equals(t, "second", span.TraceID())
}

func TestSetTraceID_AllSpansReceiveSameID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx = trace.SetTraceID(ctx, "trace-123")
	ctx, parent := trace.Start(ctx, "parent")
	_, child := trace.Start(ctx, "child")
	assert.Equals(t, "trace-123", parent.TraceID())
	assert.Equals(t, "trace-123", child.TraceID())
}

func TestStart_SpanID_RootSpan_ReturnsZero(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	assert.Equals(t, "0", span.SpanID())
}

func TestStart_SpanID_ChildSpan_ReturnsIncrementingID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	_, child1 := trace.Start(ctx, "child1")
	_, child2 := trace.Start(ctx, "child2")
	_, child3 := trace.Start(ctx, "child3")
	assert.Equals(t, "0", parent.SpanID())
	assert.Equals(t, "1", child1.SpanID())
	assert.Equals(t, "2", child2.SpanID())
	assert.Equals(t, "3", child3.SpanID())
}

func TestStart_SpanID_NestedSpans_ReturnsIncrementingID(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, root := trace.Start(ctx, "root")
	ctx, child := trace.Start(ctx, "child")
	_, grandchild := trace.Start(ctx, "grandchild")
	assert.Equals(t, "0", root.SpanID())
	assert.Equals(t, "1", child.SpanID())
	assert.Equals(t, "2", grandchild.SpanID())
}

func TestStart_SpanID_ConcurrentChildCreation_AssignsUniqueIDs(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	children := parent.Children()
	assert.Equals(t, 0, len(children))
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				trace.Start(ctx, "child")
			}
		})
	}
	waitGroup.Wait()
	children = parent.Children()
	ids := make(map[string]bool)
	ids["0"] = true
	for _, child := range children {
		id := child.SpanID()
		assert.False(t, ids[id], assert.Continue())
		ids[id] = true
	}
	assert.Equals(t, goroutines*iterations+1, len(ids))
}

func TestSetExporter_SpanEnd_InvokesExporter(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	exp := &mockExporter{}
	ctx = trace.SetExporter(ctx, exp)
	_, testSpan := trace.Start(ctx, "test")
	assert.Equals(t, 0, len(exp.exportedSpans))
	testSpan.End()
	assert.Equals(t, 1, len(exp.exportedSpans))
	assert.Equals(t, testSpan, exp.exportedSpans[0])
}

func TestSetExporter_NoExporter_DoesNotPanic(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, testSpan := trace.Start(ctx, "test")
	testSpan.End()
}

func TestSetExporter_NestedSpans_ExportsEachSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	exp := &mockExporter{}
	ctx = trace.SetExporter(ctx, exp)
	ctx, root := trace.Start(ctx, "root")
	ctx, child := trace.Start(ctx, "child")
	_, grandchild := trace.Start(ctx, "grandchild")
	grandchild.End()
	child.End()
	root.End()
	assert.Equals(t, 3, len(exp.exportedSpans))
	assert.Equals(t, grandchild, exp.exportedSpans[0])
	assert.Equals(t, child, exp.exportedSpans[1])
	assert.Equals(t, root, exp.exportedSpans[2])
}
