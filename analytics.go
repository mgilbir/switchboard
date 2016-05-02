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

type AnalyticsMsg struct {
	Category string
	Time     time.Time
}

type Analytics struct {
	categoryCount map[string]uint64
	totalCount    uint64
	lock          sync.RWMutex
}

func NewAnalytics() *Analytics {
	r := &Analytics{
		categoryCount: make(map[string]uint64),
	}

	return r
}

func (a *Analytics) Handle(msg AnalyticsMsg) {
	a.lock.Lock()
	a.totalCount++
	if msg.Category != "" {
		a.categoryCount[msg.Category]++
	}

	a.lock.Unlock()
	fmt.Println("ADDED", msg)
}

func (a *Analytics) Count() uint64 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.totalCount
}
