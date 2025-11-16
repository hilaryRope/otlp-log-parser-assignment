package counter

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"otlp-log-parser-assignment/internal/logger"
)

// WindowCounter tracks counts of attribute values within time windows
type WindowCounter struct {
	mu             sync.RWMutex
	currentCounts  map[string]int64
	windowDuration time.Duration
	ticker         *time.Ticker
	stopCh         chan struct{}
	logger         *logger.Logger
	windowStart    time.Time
	totalWindows   int64
	debug          bool
}

func NewWindowCounter(windowDuration time.Duration, logger *logger.Logger, debug bool) *WindowCounter {
	return &WindowCounter{
		currentCounts:  make(map[string]int64),
		windowDuration: windowDuration,
		stopCh:         make(chan struct{}),
		logger:         logger.With("component", "counter"),
		windowStart:    time.Now(),
		debug:          debug,
	}
}

func (wc *WindowCounter) Start() {
	wc.ticker = time.NewTicker(wc.windowDuration)

	go func() {
		for {
			select {
			case <-wc.ticker.C:
				wc.reportAndReset()
			case <-wc.stopCh:
				return
			}
		}
	}()

	wc.logger.Infow("Window counter started", "duration", wc.windowDuration)
}

func (wc *WindowCounter) Stop() {
	if wc.ticker != nil {
		wc.ticker.Stop()
	}
	close(wc.stopCh)

	wc.reportAndReset()

	wc.logger.Infow("Window counter stopped")
}

// Increment increments the count for a given attribute value
func (wc *WindowCounter) Increment(attributeValue string) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.currentCounts[attributeValue]++
}

func (wc *WindowCounter) IncrementBatch(attributeValues []string) {
	if len(attributeValues) == 0 {
		return
	}

	wc.mu.Lock()
	defer wc.mu.Unlock()

	for _, value := range attributeValues {
		wc.currentCounts[value]++
	}
}

// reportAndReset reports the current counts and resets the counter
func (wc *WindowCounter) reportAndReset() {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if len(wc.currentCounts) == 0 {
		wc.logger.Infow("No data to report in this window")
		return
	}

	keys := make([]string, 0, len(wc.currentCounts))
	for key := range wc.currentCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	now := time.Now()
	windowEnd := now
	windowDuration := windowEnd.Sub(wc.windowStart)
	totalLogs := int64(0)
	for _, count := range wc.currentCounts {
		totalLogs += count
	}
	wc.totalWindows++

	// Create detailed counts with percentages
	type AttributeCount struct {
		Count      int64   `json:"count"`
		Percentage float64 `json:"percentage"`
	}

	detailedCounts := make(map[string]AttributeCount)
	for _, key := range keys {
		count := wc.currentCounts[key]
		percentage := float64(count) / float64(totalLogs) * 100
		detailedCounts[key] = AttributeCount{
			Count:      count,
			Percentage: percentage,
		}
	}

	wc.logger.Infow("Log attribute counts report",
		"window_number", wc.totalWindows,
		"time_range", fmt.Sprintf("%s - %s", wc.windowStart.Format("15:04:05"), windowEnd.Format("15:04:05")),
		"duration", windowDuration.Round(time.Millisecond).String(),
		"total_logs", totalLogs,
		"unique_values", len(keys),
		"attribute_counts", detailedCounts,
	)

	// Show beautiful ASCII table in debug mode
	if wc.debug {
		wc.printASCIITable(windowEnd, windowDuration, totalLogs, keys)
	}

	// Reset counts for next window
	wc.currentCounts = make(map[string]int64)
	wc.windowStart = now
}

// printASCIITable prints a beautiful ASCII table for debug mode
func (wc *WindowCounter) printASCIITable(windowEnd time.Time, windowDuration time.Duration, totalLogs int64, keys []string) {
	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Log Attribute Counts Report                      ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Window #%-3d                                               ║\n", wc.totalWindows)
	fmt.Printf("║ Time Range: %-45s ║\n", wc.windowStart.Format("15:04:05")+" - "+windowEnd.Format("15:04:05"))
	fmt.Printf("║ Duration: %-47s ║\n", windowDuration.Round(time.Millisecond).String())
	fmt.Printf("║ Total Logs: %-45d ║\n", totalLogs)
	fmt.Printf("║ Unique Values: %-42d ║\n", len(keys))
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║ Attribute Value Counts:                                   ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	for _, key := range keys {
		count := wc.currentCounts[key]
		percentage := float64(count) / float64(totalLogs) * 100
		fmt.Printf("║ %-40s %8d (%5.1f%%) ║\n", truncate(key, 40), count, percentage)
	}
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println("")
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetCurrentCounts returns a copy of the current counts (for testing)
func (wc *WindowCounter) GetCurrentCounts() map[string]int64 {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	counts := make(map[string]int64, len(wc.currentCounts))
	for k, v := range wc.currentCounts {
		counts[k] = v
	}
	return counts
}
