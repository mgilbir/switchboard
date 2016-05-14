package switchboard

import (
	"fmt"
	"sync"
	"time"

	"container/ring"
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
	LastQueries() ([]AnalyticsMsg, error)
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

type LastEntries struct {
	r *ring.Ring
}

func NewLastEntries(n int) *LastEntries {
	return &LastEntries{
		r: ring.New(n),
	}
}

func (l LastEntries) All() []AnalyticsMsg {
	var r []AnalyticsMsg
	l.r.Do(func(i interface{}) {
		switch v := i.(type) {
		case AnalyticsMsg:
			r = append(r, v)
		case *AnalyticsMsg:
			r = append(r, *v)
		}
	})
	return r
}

func (l *LastEntries) Add(m AnalyticsMsg) error {
	n := l.r.Next()
	n.Value = m
	l.r = n
	return nil
}

type Analytics struct {
	categoryCount    map[string]EntryStats
	totalCount       uint64
	lastTimeModified time.Time
	lastN            *LastEntries
	lock             sync.RWMutex
}

func NewAnalytics() *Analytics {
	r := &Analytics{
		categoryCount: make(map[string]EntryStats),
		lastN:         NewLastEntries(50),
	}

	return r
}

func (a *Analytics) Handle(msg AnalyticsMsg) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.lastN.Add(msg)

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

func (a *Analytics) LastQueries() ([]AnalyticsMsg, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.lastN.All(), nil
}
