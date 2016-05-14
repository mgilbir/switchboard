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

func TestLastEntries(t *testing.T) {
	l := NewLastEntries(3)

	if len(l.All()) != 0 {
		t.Fatalf("Expected an empty list")
	}

	m1 := AnalyticsMsg{Category: "TEST1", Time: Now()}
	m2 := AnalyticsMsg{Category: "TEST2", Time: Now()}
	m3 := AnalyticsMsg{Category: "TEST3", Time: Now()}
	m4 := AnalyticsMsg{Category: "TEST4", Time: Now()}

	l.Add(m1)
	l.Add(m2)

	if v := len(l.All()); v != 2 {
		t.Fatalf("Expected a list of 2 items, got %d", v)
	}

	l.Add(m3)
	l.Add(m4)

	if v := len(l.All()); v != 3 {
		t.Fatalf("Expected a list of 3 items, got %d", v)
	}

	expected := []AnalyticsMsg{m4, m2, m3}
	got := l.All()

	for i, j := range got {
		if j != expected[i] {
			t.Fatalf("At position %d: Expected %v, got %v", i, expected[i], j)
		}
	}
}
