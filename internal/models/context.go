package models

import (
	"runtime"
	"time"
)

// Context represents the operational context when an error occurs
type Context struct {
	Environment    string                 `json:"environment"`              // e.g., "production", "staging"
	SystemInfo     SystemInfo            `json:"system_info"`              // System-level information
	RuntimeMetrics RuntimeMetrics        `json:"runtime_metrics"`          // Runtime performance metrics
	Variables      map[string]string     `json:"variables,omitempty"`      // Relevant variables and their values
	CustomData     map[string]interface{} `json:"custom_data,omitempty"`   // Any additional context-specific data
}

// SystemInfo contains system-level information
type SystemInfo struct {
	OS            string    `json:"os"`
	Architecture  string    `json:"architecture"`
	NumCPU        int       `json:"num_cpu"`
	Hostname      string    `json:"hostname"`
	GoVersion     string    `json:"go_version"`
	StartTime     time.Time `json:"start_time"`
	NumGoroutines int       `json:"num_goroutines"`
}

// RuntimeMetrics contains runtime performance metrics
type RuntimeMetrics struct {
	HeapAlloc    uint64    `json:"heap_alloc"`     // Current heap allocation
	HeapSys      uint64    `json:"heap_sys"`       // Total heap size
	NumGC        uint32    `json:"num_gc"`         // Number of completed GC cycles
	PauseTotalNs uint64    `json:"pause_total_ns"` // Total GC pause time
	NumThreads   int       `json:"num_threads"`    // Number of OS threads
	CPUUsage     float64   `json:"cpu_usage"`      // CPU usage percentage
	MemoryUsage  float64   `json:"memory_usage"`   // Memory usage percentage
	Timestamp    time.Time `json:"timestamp"`      // When these metrics were collected
}

// NewContext creates a new Context instance with current system information
func NewContext() Context {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return Context{
		Environment: getEnvironment(),
		SystemInfo: SystemInfo{
			OS:            runtime.GOOS,
			Architecture:  runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			Hostname:     getHostname(),
			GoVersion:    runtime.Version(),
			StartTime:    time.Now(),
			NumGoroutines: runtime.NumGoroutine(),
		},
		RuntimeMetrics: RuntimeMetrics{
			HeapAlloc:    memStats.HeapAlloc,
			HeapSys:      memStats.HeapSys,
			NumGC:        memStats.NumGC,
			PauseTotalNs: memStats.PauseTotalNs,
			NumThreads:   runtime.NumCPU(), // This is a simplification
			CPUUsage:     getCPUUsage(),
			MemoryUsage:  getMemoryUsage(),
			Timestamp:    time.Now(),
		},
		Variables:  make(map[string]string),
		CustomData: make(map[string]interface{}),
	}
}

// AddVariable adds a variable to the context
func (c *Context) AddVariable(key, value string) {
	c.Variables[key] = value
}

// AddCustomData adds custom data to the context
func (c *Context) AddCustomData(key string, value interface{}) {
	c.CustomData[key] = value
}

// UpdateRuntimeMetrics updates the runtime metrics with current values
func (c *Context) UpdateRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.RuntimeMetrics = RuntimeMetrics{
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		NumGC:        memStats.NumGC,
		PauseTotalNs: memStats.PauseTotalNs,
		NumThreads:   runtime.NumCPU(),
		CPUUsage:     getCPUUsage(),
		MemoryUsage:  getMemoryUsage(),
		Timestamp:    time.Now(),
	}
}

// Helper functions that need to be implemented
func getEnvironment() string {
	// Implementation to get the current environment
	return "development"
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getCPUUsage() float64 {
	// Implementation to get CPU usage
	// This would typically use something like github.com/shirou/gopsutil
	return 0.0
}

func getMemoryUsage() float64 {
	// Implementation to get memory usage
	// This would typically use something like github.com/shirou/gopsutil
	return 0.0
}
