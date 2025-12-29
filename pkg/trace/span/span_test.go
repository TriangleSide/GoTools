package span_test

import (
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/trace/event"
	"github.com/TriangleSide/GoTools/pkg/trace/span"
	"github.com/TriangleSide/GoTools/pkg/trace/status"
)

func TestNew_NilParent_CreatesRootSpan(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test-span", nil)
	assert.NotNil(t, testSpan)
	assert.Equals(t, "test-span", testSpan.Name())
	assert.Nil(t, testSpan.Parent())
	assert.Equals(t, 0, len(testSpan.Children()))
	assert.Equals(t, 0, len(testSpan.Attributes()))
	assert.Equals(t, 0, len(testSpan.Events()))
	assert.Equals(t, status.Unset, testSpan.StatusCode())
}

func TestNew_WithParent_CreatesChildSpan(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", nil)
	child := span.New("child", parent)
	assert.Equals(t, parent, child.Parent())
	assert.Equals(t, 1, len(parent.Children()))
	assert.Equals(t, child, parent.Children()[0])
}

func TestNew_MultipleChildren_AllAddedToParent(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", nil)
	child1 := span.New("child1", parent)
	child2 := span.New("child2", parent)
	child3 := span.New("child3", parent)
	children := parent.Children()
	assert.Equals(t, 3, len(children))
	assert.Equals(t, child1, children[0])
	assert.Equals(t, child2, children[1])
	assert.Equals(t, child3, children[2])
}

func TestNew_NestedSpans_CreatesHierarchy(t *testing.T) {
	t.Parallel()
	root := span.New("root", nil)
	child := span.New("child", root)
	grandchild := span.New("grandchild", child)
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
	s := span.New("test", nil)
	after := time.Now()
	assert.True(t, !s.StartTime().Before(before))
	assert.True(t, !s.StartTime().After(after))
}

func TestNew_ConcurrentChildCreation_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	parent := span.New("parent", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.New("child", parent)
			}
		})
	}
	waitGroup.Wait()
	children := parent.Children()
	assert.Equals(t, goroutines*iterations, len(children))
}

func TestSpanEnd_RecordsEndTime(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
	time.Sleep(10 * time.Millisecond)
	duration := s.Duration()
	assert.True(t, duration >= 10*time.Millisecond)
}

func TestSpanDuration_AfterEnd_ReturnsFixedDuration(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
	time.Sleep(10 * time.Millisecond)
	s.End()
	duration1 := s.Duration()
	time.Sleep(10 * time.Millisecond)
	duration2 := s.Duration()
	assert.Equals(t, duration1, duration2)
}

func TestSpanChildren_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	parent := span.New("parent", nil)
	span.New("child", parent)
	children1 := parent.Children()
	children2 := parent.Children()
	assert.Equals(t, len(children1), len(children2))
	children1[0] = nil
	assert.NotNil(t, parent.Children()[0])
}

func TestSpanEnd_ConcurrentCalls_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
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
	parent := span.New("parent", nil)
	for range 5 {
		span.New("child", parent)
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
	parent := span.New("parent", nil)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				span.New("child", parent)
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
	s := span.New("test", nil)
	s.SetAttributes(attribute.String("key", "value"))
	attrs := s.Attributes()
	assert.Equals(t, 1, len(attrs))
	assert.Equals(t, "key", attrs[0].Key())
	assert.Equals(t, "value", attrs[0].StringValue())
}

func TestSetAttributes_MultipleTypes_AllSupported(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
	attrs := s.Attributes()
	assert.Equals(t, 0, len(attrs))
}

func TestAttributes_MultipleAttributes_ReturnsAll(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
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
	testSpan := span.New("test", nil)
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
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
	e := event.New("test-event")
	s.AddEvent(e)
	events := s.Events()
	assert.Equals(t, 1, len(events))
	assert.Equals(t, "test-event", events[0].Name())
}

func TestAddEvent_MultipleEvents_AllRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
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
	s := span.New("test", nil)
	events := s.Events()
	assert.Equals(t, 0, len(events))
}

func TestEvents_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
	s.AddEvent(event.New("original"))
	events := s.Events()
	events[0] = event.New("modified")
	originalEvents := s.Events()
	assert.Equals(t, "original", originalEvents[0].Name())
}

func TestAddEvent_WithAttributes_PreservesAttributes(t *testing.T) {
	t.Parallel()
	testSpan := span.New("test", nil)
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
	testSpan := span.New("test", nil)
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
	testSpan := span.New("test", nil)
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
	s := span.New("test", nil)
	assert.Equals(t, status.Unset, s.StatusCode())
}

func TestSetStatus_Error_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
	s.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, s.StatusCode())
}

func TestSetStatus_Success_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
	s.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, s.StatusCode())
}

func TestSetStatus_CanBeOverwritten(t *testing.T) {
	t.Parallel()
	s := span.New("test", nil)
	s.SetStatusCode(status.Error)
	assert.Equals(t, status.Error, s.StatusCode())
	s.SetStatusCode(status.Success)
	assert.Equals(t, status.Success, s.StatusCode())
}

func TestSetStatus_ConcurrentWrites_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	testSpan := span.New("test", nil)
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
	testSpan := span.New("test", nil)
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
