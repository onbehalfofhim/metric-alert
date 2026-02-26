package models

import (
	"math/rand"
	"runtime"
	"sync"
)

// сборщик метрик
type Collector struct {
	mu        sync.Mutex // для разделения потоков при работе с хранилищем метрик
	pollCount int64
	metrics   map[string]Metric
}

// констурктор для сборщика метрик
func NewCollector() *Collector {
	return &Collector{
		pollCount: 0,
		metrics:   make(map[string]Metric, 0),
	}
}

// метод выдачи метрик из сборщика
func (c *Collector) GetMetrics() []Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]Metric, 0, len(c.metrics))

	for _, v := range c.metrics {
		result = append(result, v)
	}

	return result
}

// метод сбора метрик
func (c *Collector) CollectMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pollCount++

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// ===== Основные аллокации =====
	c.metrics["Alloc"] = NewGauge("Alloc", float64(m.Alloc))                // текущие байты, выделенные под heap-объекты
	c.metrics["TotalAlloc"] = NewGauge("TotalAlloc", float64(m.TotalAlloc)) // всего выделено байт за всё время
	c.metrics["Sys"] = NewGauge("Sys", float64(m.Sys))                      // всего байт, полученных от ОС
	c.metrics["Lookups"] = NewGauge("Lookups", float64(m.Lookups))          // количество pointer lookups
	c.metrics["Mallocs"] = NewGauge("Mallocs", float64(m.Mallocs))          // количество malloc
	c.metrics["Frees"] = NewGauge("Frees", float64(m.Frees))                // количество free

	// ===== Heap =====
	c.metrics["HeapAlloc"] = NewGauge("HeapAlloc", float64(m.HeapAlloc))          // байты, выделенные под heap
	c.metrics["HeapSys"] = NewGauge("HeapSys", float64(m.HeapSys))                // байты, запрошенные у ОС под heap
	c.metrics["HeapIdle"] = NewGauge("HeapIdle", float64(m.HeapIdle))             // неиспользуемые байты heap
	c.metrics["HeapInuse"] = NewGauge("HeapInuse", float64(m.HeapInuse))          // используемые байты heap
	c.metrics["HeapReleased"] = NewGauge("HeapReleased", float64(m.HeapReleased)) // байты heap, возвращённые ОС
	c.metrics["HeapObjects"] = NewGauge("HeapObjects", float64(m.HeapObjects))    // количество объектов в heap

	// ===== GC =====
	c.metrics["GCCPUFraction"] = NewGauge("GCCPUFraction", float64(m.GCCPUFraction)) // доля CPU, потраченная на GC
	c.metrics["GCSys"] = NewGauge("GCSys", float64(m.GCSys))                         // байты под GC структуры
	c.metrics["NextGC"] = NewGauge("NextGC", float64(m.NextGC))                      // порог heap для следующего GC
	c.metrics["LastGC"] = NewGauge("LastGC", float64(m.LastGC))                      // timestamp последнего GC (ns)
	c.metrics["PauseTotalNs"] = NewGauge("PauseTotalNs", float64(m.PauseTotalNs))    // суммарные паузы GC
	c.metrics["NumGC"] = NewGauge("NumGC", float64(m.NumGC))                         // количество GC
	c.metrics["NumForcedGC"] = NewGauge("NumForcedGC", float64(m.NumForcedGC))       // количество принудительных GC

	// ===== Stack =====
	c.metrics["StackInuse"] = NewGauge("StackInuse", float64(m.StackInuse)) // используемые байты stack
	c.metrics["StackSys"] = NewGauge("StackSys", float64(m.StackSys))       // байты под stack от ОС

	// ===== MCache / MSpan =====
	c.metrics["MCacheInuse"] = NewGauge("MCacheInuse", float64(m.MCacheInuse)) // используемые MCache
	c.metrics["MCacheSys"] = NewGauge("MCacheSys", float64(m.MCacheSys))       // MCache от ОС
	c.metrics["MSpanInuse"] = NewGauge("MSpanInuse", float64(m.MSpanInuse))    // используемые MSpan
	c.metrics["MSpanSys"] = NewGauge("MSpanSys", float64(m.MSpanSys))          // MSpan от ОС

	// ===== Прочее =====
	c.metrics["BuckHashSys"] = NewGauge("BuckHashSys", float64(m.MSpanSys)) // байты под hash buckets
	c.metrics["OtherSys"] = NewGauge("OtherSys", float64(m.MSpanSys))       // прочие аллокации рантайма

	//== Random Value ==
	c.metrics["RandomValue"] = NewGauge("RandomValue", rand.Float64())

	//== Poll Counter ==
	c.metrics["PollCount"] = NewCounter("PollCount", c.pollCount)
}
