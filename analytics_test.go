package switchboard

import "testing"

func TestImplementsInterfaces(t *testing.T) {
	var _ AnalyticsHandler = NewAnalytics()
	var _ AnalyticsAPI = NewAnalytics()
}

func TestAnalytics(t *testing.T) {
	a := NewAnalytics()
	if a.totalCount != 0 {
		t.Fatalf("Expected an empty totalCount and got %d", a.totalCount)
	}
	if len(a.categoryCount) != 0 {
		t.Fatalf("Expected an empty categoryCount and got %d", a.categoryCount)
	}

	a.Handle(AnalyticsMsg{Category: "TEST", Time: Now()})
	a.Handle(AnalyticsMsg{Category: "", Time: Now()})

	// time.Sleep(time.Second)
	if a.Count() != 2 {
		t.Fatalf("Expected a totalCount of 2 and got %d", a.Count())
	}
	if len(a.categoryCount) != 1 {
		t.Fatalf("Expected one categoryCount and got %d", a.categoryCount)
	}
}
