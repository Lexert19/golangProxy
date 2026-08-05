package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type StatusResponse struct {
	IsRunning      bool       `json:"is_running"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	CheckedCount   int64      `json:"checked_count"`
	RemainingCount int64      `json:"remaining_count"`
	WorkingCount   int64      `json:"working_count"`
}

type MetricsResponse struct {
	TotalChecks        int64   `json:"total_checks"`
	SuccessCount       int64   `json:"success_count"`
	ErrorCount         int64   `json:"error_count"`
	AvgCheckDurationMs float64 `json:"avg_check_duration_ms"`
	ActiveWorkers      int64   `json:"active_workers"`
	UptimeSeconds      float64 `json:"uptime_seconds"`
}

type MetricsTracker struct {
	mu                 sync.RWMutex
	programStartTime   time.Time
	isRunning          bool
	currentRunStart    time.Time
	checkedCount       int64
	remainingCount     int64
	workingCount       int64
	totalChecks        int64
	successCount       int64
	errorCount         int64
	totalCheckDuration int64
	activeWorkers      int64
	workingProxies     []Proxy
}

func NewMetricsTracker() *MetricsTracker {
	return &MetricsTracker{
		programStartTime: time.Now(),
		workingProxies:   make([]Proxy, 0),
	}
}

func (m *MetricsTracker) StartRun(totalProxies int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
	m.currentRunStart = time.Now()
	m.checkedCount = 0
	m.remainingCount = int64(totalProxies)
	m.workingCount = 0
}

func (m *MetricsTracker) FinishRun() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
}

func (m *MetricsTracker) RecordCheck(p Proxy, err error, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkedCount++
	if m.remainingCount > 0 {
		m.remainingCount--
	}
	m.totalChecks++
	atomic.AddInt64(&m.totalCheckDuration, int64(duration))

	if err == nil {
		m.workingCount++
		m.successCount++
		m.workingProxies = append(m.workingProxies, p)
	} else {
		m.errorCount++
	}
}

func (m *MetricsTracker) IncWorker() {
	atomic.AddInt64(&m.activeWorkers, 1)
}

func (m *MetricsTracker) DecWorker() {
	atomic.AddInt64(&m.activeWorkers, -1)
}

func (m *MetricsTracker) GetStatus() StatusResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var startTime *time.Time
	if m.isRunning {
		t := m.currentRunStart
		startTime = &t
	}

	return StatusResponse{
		IsRunning:      m.isRunning,
		StartTime:      startTime,
		CheckedCount:   m.checkedCount,
		RemainingCount: m.remainingCount,
		WorkingCount:   m.workingCount,
	}
}

func (m *MetricsTracker) GetMetrics() MetricsResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDurationMs := float64(0)
	if m.totalChecks > 0 {
		avgDurationMs = float64(m.totalCheckDuration) / float64(m.totalChecks) / float64(time.Millisecond)
	}

	return MetricsResponse{
		TotalChecks:        m.totalChecks,
		SuccessCount:       m.successCount,
		ErrorCount:         m.errorCount,
		AvgCheckDurationMs: avgDurationMs,
		ActiveWorkers:      atomic.LoadInt64(&m.activeWorkers),
		UptimeSeconds:      time.Since(m.programStartTime).Seconds(),
	}
}

func (m *MetricsTracker) GetResults() []Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Proxy, len(m.workingProxies))
	copy(result, m.workingProxies)
	return result
}