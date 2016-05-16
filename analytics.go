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
	Time          TimeBin
	CategoryCount map[string]EntryStats
	TotalCount    uint64
}

type AnalyticsAPI interface {
	LastQueries() ([]AnalyticsMsg, error)
	CategoryStatsAll() ([]CategoryStats, error)
	CategoryStats(categories []string) ([]CategoryStats, error)
}

type AnalyticsMsg struct {
	Domain    string
	QueryType string
	Category  string
	Time      time.Time
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

type TimedAnalytics struct {
	categoryCount map[string]EntryStats
	totalCount    uint64
}

func NewTimedAnalytics() TimedAnalytics {
	return TimedAnalytics{
		categoryCount: make(map[string]EntryStats),
	}
}

type Analytics struct {
	data             map[TimeBin]TimedAnalytics
	lastTimeModified time.Time
	lastN            *LastEntries
	lock             sync.RWMutex
}

func NewAnalytics() *Analytics {
	r := &Analytics{
		data:  make(map[TimeBin]TimedAnalytics),
		lastN: NewLastEntries(50),
	}

	return r
}

func (a *Analytics) Handle(msg AnalyticsMsg) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.lastN.Add(msg)

	bin := TimeToHourBin(msg.Time)
	ta, ok := a.data[bin]
	if !ok {
		ta = NewTimedAnalytics()
	}
	ta.totalCount++
	if msg.Category != "" {
		m := ta.categoryCount[msg.Category]
		m.Count++
		m.LastTimeModified = msg.Time
		ta.categoryCount[msg.Category] = m
	}

	if msg.Time.After(a.lastTimeModified) {
		a.lastTimeModified = msg.Time
	}
	a.data[bin] = ta
}

func (a *Analytics) CategoryStatsAll() ([]CategoryStats, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var r []CategoryStats

	for tb, data := range a.data {
		r = append(r, CategoryStats{
			//TODO: return a copy!
			Time:          tb,
			CategoryCount: data.categoryCount,
			TotalCount:    data.totalCount,
		})
	}

	return r, nil
}

func (a *Analytics) CategoryStats(categories []string) ([]CategoryStats, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return []CategoryStats{}, fmt.Errorf("Not implemented")
}

func (a *Analytics) Count() uint64 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	bin := TimeToHourBin(Now())
	return a.data[bin].totalCount
}

func (a *Analytics) LastQueries() ([]AnalyticsMsg, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.lastN.All(), nil
}
