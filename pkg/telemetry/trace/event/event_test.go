package event_test

import (
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/event"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestNew_WithName_RecordsName(t *testing.T) {
	t.Parallel()
	e := event.New("test-event")
	assert.Equals(t, "test-event", e.Name())
}

func TestNew_WithoutAttributes_HasEmptyAttributes(t *testing.T) {
	t.Parallel()
	e := event.New("test-event")
	attrs := e.Attributes()
	assert.Equals(t, 0, len(attrs))
}

func TestNew_WithAttributes_RecordsAttributes(t *testing.T) {
	t.Parallel()
	e := event.New("test-event",
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
	attrs := e.Attributes()
	assert.Equals(t, 2, len(attrs))
	assert.Equals(t, "key1", attrs[0].Key())
	assert.Equals(t, "value1", attrs[0].StringValue())
	assert.Equals(t, "key2", attrs[1].Key())
	assert.Equals(t, int64(42), attrs[1].IntValue())
}

func TestNew_RecordsTimestamp(t *testing.T) {
	t.Parallel()
	before := time.Now()
	e := event.New("test-event")
	after := time.Now()
	assert.True(t, !e.Timestamp().Before(before))
	assert.True(t, !e.Timestamp().After(after))
}

func TestTimestamp_IsImmutable(t *testing.T) {
	t.Parallel()
	e := event.New("test-event")
	ts1 := e.Timestamp()
	time.Sleep(10 * time.Millisecond)
	ts2 := e.Timestamp()
	assert.Equals(t, ts1, ts2)
}

func TestAttributes_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	e := event.New("test-event", attribute.String("key", "value"))
	attrs := e.Attributes()
	attrs[0] = attribute.String("modified", "modified")
	originalAttrs := e.Attributes()
	assert.Equals(t, "key", originalAttrs[0].Key())
	assert.Equals(t, "value", originalAttrs[0].StringValue())
}

func TestNew_DoesNotModifyInputSlice(t *testing.T) {
	t.Parallel()
	inputAttrs := []*attribute.Attribute{
		attribute.String("key", "value"),
	}
	e := event.New("test-event", inputAttrs...)
	inputAttrs[0] = attribute.String("modified", "modified")
	eventAttrs := e.Attributes()
	assert.Equals(t, "key", eventAttrs[0].Key())
	assert.Equals(t, "value", eventAttrs[0].StringValue())
}

func TestEvent_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 5000
	evt := event.New("test-event",
		attribute.String("key", "value"),
		attribute.Int("count", 42),
	)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				_ = evt.Name()
				_ = evt.Timestamp()
				_ = evt.Attributes()
			}
		})
	}
	waitGroup.Wait()
}

func TestNew_AllAttributeTypes_Supported(t *testing.T) {
	t.Parallel()
	evt := event.New("test-event",
		attribute.String("string", "hello"),
		attribute.Int("int", 42),
		attribute.Float("float", 3.14),
		attribute.Bool("bool", true),
	)
	attrs := evt.Attributes()
	assert.Equals(t, 4, len(attrs))
	assert.Equals(t, attribute.TypeString, attrs[0].Type())
	assert.Equals(t, attribute.TypeInt, attrs[1].Type())
	assert.Equals(t, attribute.TypeFloat, attrs[2].Type())
	assert.Equals(t, attribute.TypeBool, attrs[3].Type())
}
