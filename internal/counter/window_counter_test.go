package counter

import (
	"testing"
	"time"

	"otlp-log-parser-assignment/internal/logger"
)

func TestWindowCounter_Increment(t *testing.T) {
	testLogger, _ := logger.New(false)
	wc := NewWindowCounter(1*time.Second, testLogger, false)

	wc.Increment("value1")
	wc.Increment("value2")
	wc.Increment("value1")

	counts := wc.GetCurrentCounts()

	if counts["value1"] != 2 {
		t.Errorf("Expected count for value1 to be 2, got %d", counts["value1"])
	}

	if counts["value2"] != 1 {
		t.Errorf("Expected count for value2 to be 1, got %d", counts["value2"])
	}
}

func TestWindowCounter_IncrementBatch(t *testing.T) {
	testLogger, _ := logger.New(false)
	wc := NewWindowCounter(1*time.Second, testLogger, false)

	values := []string{"a", "b", "a", "c", "b", "a"}
	wc.IncrementBatch(values)

	counts := wc.GetCurrentCounts()

	expected := map[string]int64{
		"a": 3,
		"b": 2,
		"c": 1,
	}

	for key, expectedCount := range expected {
		if counts[key] != expectedCount {
			t.Errorf("Expected count for %s to be %d, got %d", key, expectedCount, counts[key])
		}
	}
}

func TestWindowCounter_IncrementBatch_Empty(t *testing.T) {
	testLogger, _ := logger.New(false)
	wc := NewWindowCounter(1*time.Second, testLogger, false)

	wc.IncrementBatch([]string{})

	counts := wc.GetCurrentCounts()
	if len(counts) != 0 {
		t.Errorf("Expected empty counts, got %v", counts)
	}
}

func TestWindowCounter_StartStop(t *testing.T) {
	testLogger, _ := logger.New(false)
	wc := NewWindowCounter(100*time.Millisecond, testLogger, false)

	wc.Increment("test")

	wc.Start()
	time.Sleep(150 * time.Millisecond)

	// After window duration, counts should be reset
	counts := wc.GetCurrentCounts()
	if len(counts) != 0 {
		t.Errorf("Expected counts to be reset after window, got %v", counts)
	}

	wc.Stop()
}

func TestWindowCounter_ConcurrentAccess(t *testing.T) {
	testLogger, _ := logger.New(false)
	wc := NewWindowCounter(1*time.Second, testLogger, false)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				wc.Increment("concurrent")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	counts := wc.GetCurrentCounts()
	if counts["concurrent"] != 1000 {
		t.Errorf("Expected count to be 1000, got %d", counts["concurrent"])
	}
}
