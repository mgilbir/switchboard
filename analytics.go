package switchboard

import (
	"fmt"
	"sync"
	"time"
)

var (
	NoOpAnalytics AnalyticsHandlerFunc = func(AnalyticsMsg) {}
)

type AnalyticsHandlerFunc func(AnalyticsMsg)

type AnalyticsHandler interface {
	Handle(AnalyticsMsg)
}

type CategoryStats struct {
	CategoryCount map[string]EntryStats
	TotalCount    uint64
}

type AnalyticsAPI interface {
	CategoryStatsAll() (CategoryStats, error)
	CategoryStats(categories []string) (CategoryStats, error)
}

type AnalyticsMsg struct {
	Category string
	Time     time.Time
}

type EntryStats struct {
	Count            uint64
	LastTimeModified time.Time
}

type Analytics struct {
	categoryCount    map[string]EntryStats
	totalCount       uint64
	lastTimeModified time.Time
	lock             sync.RWMutex
}

func NewAnalytics() *Analytics {
	r := &Analytics{
		categoryCount: make(map[string]EntryStats),
	}

	return r
}

func (a *Analytics) Handle(msg AnalyticsMsg) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.totalCount++
	if msg.Category != "" {
		m := a.categoryCount[msg.Category]
		m.Count++
		m.LastTimeModified = msg.Time
		a.categoryCount[msg.Category] = m
	}

	if msg.Time.After(a.lastTimeModified) {
		a.lastTimeModified = msg.Time
	}
}

func (a *Analytics) CategoryStatsAll() (CategoryStats, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return CategoryStats{
		//TODO: return a copy!
		CategoryCount: a.categoryCount,
		TotalCount:    a.totalCount,
	}, nil
}

func (a *Analytics) CategoryStats(categories []string) (CategoryStats, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return CategoryStats{}, fmt.Errorf("Not implemented")
}

func (a *Analytics) Count() uint64 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.totalCount
}
