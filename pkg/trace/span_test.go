package trace_test

import (
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/trace"
)

func TestStartSpan_EmptyContext_CreatesRootSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resultCtx, span := trace.Start(ctx, t.Name())
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, span)
	assert.Equals(t, t.Name(), span.Name())
	assert.Nil(t, span.Parent())
}

func TestStartSpan_WithParent_CreatesChildSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := trace.Start(ctx, "parent")
	_, child := trace.Start(ctx, "child")
	assert.Equals(t, parent, child.Parent())
	children := parent.Children()
	assert.Equals(t, 1, len(children))
	assert.Equals(t, child, children[0])
}

func TestStartSpan_MultipleChildren_AllAddedToParent(t *testing.T) {
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

func TestStartSpan_NestedSpans_CreatesHierarchy(t *testing.T) {
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

func TestStartSpan_RecordsStartTime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	before := time.Now()
	_, span := trace.Start(ctx, "test")
	after := time.Now()
	assert.True(t, !span.StartTime().Before(before))
	assert.True(t, !span.StartTime().After(after))
}

func TestSpanEnd_RecordsEndTime(t *testing.T) {
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

func TestSpanDuration_BeforeEnd_ReturnsDurationSinceStart(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	time.Sleep(10 * time.Millisecond)
	duration := span.Duration()
	assert.True(t, duration >= 10*time.Millisecond)
}

func TestSpanDuration_AfterEnd_ReturnsFixedDuration(t *testing.T) {
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

func TestFromContext_NoSpan_ReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	span := trace.FromContext(ctx)
	assert.Nil(t, span)
}

func TestFromContext_WithSpan_ReturnsSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, expectedSpan := trace.Start(ctx, "test")
	actualSpan := trace.FromContext(ctx)
	assert.Equals(t, expectedSpan, actualSpan)
}

func TestFromContext_AfterNestedSpan_ReturnsInnerSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, _ = trace.Start(ctx, "outer")
	ctx, inner := trace.Start(ctx, "inner")
	actualSpan := trace.FromContext(ctx)
	assert.Equals(t, inner, actualSpan)
}

func TestSpanChildren_ReturnsDefensiveCopy(t *testing.T) {
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

func TestStartSpan_ConcurrentChildCreation_IsThreadSafe(t *testing.T) {
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

func TestSpanEnd_ConcurrentCalls_IsThreadSafe(t *testing.T) {
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

func TestSpanDuration_ConcurrentReads_IsThreadSafe(t *testing.T) {
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

func TestSpanChildren_ConcurrentReads_IsThreadSafe(t *testing.T) {
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

func TestSpan_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
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

func TestSetAttribute_SingleAttribute_CanBeRetrieved(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("key", "value")
	value, ok := span.Attribute("key")
	assert.True(t, ok)
	assert.Equals(t, "value", value)
}

func TestSetAttribute_MultipleTypes_AllSupported(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("string", "hello")
	span.SetAttribute("int", 42)
	span.SetAttribute("float", 3.14)
	span.SetAttribute("bool", true)
	span.SetAttribute("slice", []int{1, 2, 3})
	stringVal, exists := span.Attribute("string")
	assert.True(t, exists)
	assert.Equals(t, "hello", stringVal)
	intVal, exists := span.Attribute("int")
	assert.True(t, exists)
	assert.Equals(t, 42, intVal)
	floatVal, exists := span.Attribute("float")
	assert.True(t, exists)
	assert.Equals(t, 3.14, floatVal)
	boolVal, exists := span.Attribute("bool")
	assert.True(t, exists)
	assert.Equals(t, true, boolVal)
	sliceVal, exists := span.Attribute("slice")
	assert.True(t, exists)
	assert.Equals(t, []int{1, 2, 3}, sliceVal)
}

func TestSetAttribute_OverwriteExisting_UpdatesValue(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("key", "original")
	span.SetAttribute("key", "updated")
	value, ok := span.Attribute("key")
	assert.True(t, ok)
	assert.Equals(t, "updated", value)
}

func TestAttribute_NonExistentKey_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	value, ok := span.Attribute("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestAttributes_NoAttributes_ReturnsEmptyMap(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	attrs := span.Attributes()
	assert.Equals(t, 0, len(attrs))
}

func TestAttributes_MultipleAttributes_ReturnsAll(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("key1", "value1")
	span.SetAttribute("key2", "value2")
	span.SetAttribute("key3", "value3")
	attrs := span.Attributes()
	assert.Equals(t, 3, len(attrs))
	assert.Equals(t, "value1", attrs["key1"])
	assert.Equals(t, "value2", attrs["key2"])
	assert.Equals(t, "value3", attrs["key3"])
}

func TestAttributes_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("key", "value")
	attrs := span.Attributes()
	attrs["key"] = "modified"
	attrs["newkey"] = "newvalue"
	originalValue, ok := span.Attribute("key")
	assert.True(t, ok)
	assert.Equals(t, "value", originalValue)
	_, ok = span.Attribute("newkey")
	assert.False(t, ok)
}

func TestSetAttribute_ConcurrentWrites_IsThreadSafe(t *testing.T) {
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
				span.SetAttribute(key, j)
			}
		})
	}
	waitGroup.Wait()
	attrs := span.Attributes()
	assert.Equals(t, goroutines, len(attrs))
}

func TestAttribute_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	ctx := t.Context()
	_, span := trace.Start(ctx, "test")
	span.SetAttribute("key", "value")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				value, ok := span.Attribute("key")
				assert.True(t, ok, assert.Continue())
				assert.Equals(t, "value", value, assert.Continue())
			}
		})
	}
	waitGroup.Wait()
}

func TestAttributes_ConcurrentReadAndWrite_IsThreadSafe(t *testing.T) {
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
				span.SetAttribute(key, j)
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				_ = span.Attributes()
			}
		})
		waitGroup.Go(func() {
			for range iterations {
				span.Attribute("key")
			}
		})
	}
	waitGroup.Wait()
}
