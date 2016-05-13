package switchboard

import (
	"encoding/json"
	"net/http"
)

type Api struct {
	analytics AnalyticsAPI
	*http.ServeMux
}

func NewApi(analytics AnalyticsAPI) *Api {
	a := &Api{
		analytics: analytics,
		ServeMux:  http.NewServeMux(),
	}

	// Initialize all the handlers
	a.Handle("/all", http.HandlerFunc(a.handleAll))

	return a
}

func (a *Api) handleAll(w http.ResponseWriter, r *http.Request) {
	j := json.NewEncoder(w)
	data, err := a.analytics.CategoryStatsAll()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	j.Encode(data)
}
