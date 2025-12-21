package tracer_test

import (
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/tracer"
)

func TestStartSpan_EmptyContext_CreatesRootSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resultCtx, span := tracer.StartSpan(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, span)
	assert.Contains(t, span.Name(), t.Name())
	assert.Nil(t, span.Parent())
}

func TestStartSpan_WithParent_CreatesChildSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := tracer.StartSpan(ctx)
	_, child := tracer.StartSpan(ctx)
	assert.Equals(t, parent, child.Parent())
	children := parent.Children()
	assert.Equals(t, 1, len(children))
	assert.Equals(t, child, children[0])
}

func TestStartSpan_MultipleChildren_AllAddedToParent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := tracer.StartSpan(ctx)
	_, child1 := tracer.StartSpan(ctx)
	_, child2 := tracer.StartSpan(ctx)
	_, child3 := tracer.StartSpan(ctx)
	children := parent.Children()
	assert.Equals(t, 3, len(children))
	assert.Equals(t, child1, children[0])
	assert.Equals(t, child2, children[1])
	assert.Equals(t, child3, children[2])
}

func TestStartSpan_NestedSpans_CreatesHierarchy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, root := tracer.StartSpan(ctx)
	ctx, child := tracer.StartSpan(ctx)
	_, grandchild := tracer.StartSpan(ctx)
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
	_, span := tracer.StartSpan(ctx)
	after := time.Now()
	assert.True(t, !span.StartTime().Before(before))
	assert.True(t, !span.StartTime().After(after))
}

func TestSpanEnd_RecordsEndTime(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := tracer.StartSpan(ctx)
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
	_, span := tracer.StartSpan(ctx)
	time.Sleep(10 * time.Millisecond)
	duration := span.Duration()
	assert.True(t, duration >= 10*time.Millisecond)
}

func TestSpanDuration_AfterEnd_ReturnsFixedDuration(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, span := tracer.StartSpan(ctx)
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
	span := tracer.FromContext(ctx)
	assert.Nil(t, span)
}

func TestFromContext_WithSpan_ReturnsSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, expectedSpan := tracer.StartSpan(ctx)
	actualSpan := tracer.FromContext(ctx)
	assert.Equals(t, expectedSpan, actualSpan)
}

func TestFromContext_AfterNestedSpan_ReturnsInnerSpan(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, _ = tracer.StartSpan(ctx)
	ctx, inner := tracer.StartSpan(ctx)
	actualSpan := tracer.FromContext(ctx)
	assert.Equals(t, inner, actualSpan)
}

func TestSpanChildren_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, parent := tracer.StartSpan(ctx)
	tracer.StartSpan(ctx)
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
	ctx, parent := tracer.StartSpan(ctx)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				tracer.StartSpan(ctx)
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
	_, span := tracer.StartSpan(ctx)
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
	_, span := tracer.StartSpan(ctx)
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
	ctx, parent := tracer.StartSpan(ctx)
	for range 5 {
		tracer.StartSpan(ctx)
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
	ctx, parent := tracer.StartSpan(ctx)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				tracer.StartSpan(ctx)
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
