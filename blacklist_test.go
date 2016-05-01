package switchboard

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRetrieveURL(t *testing.T) {
	server := mockBlacklistURLServer(t)
	bl, err := RetrieveBlacklistURL(server.URL+"/5", "TEST")
	if err != nil {
		t.Fatal(err)
	}

	gotCount := len(bl.Domains())

	if gotCount != 5 {
		t.Errorf("Expected %d domains, got %d\n", 5, gotCount)
	}
}

func mockBlacklistURLServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/5", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 5; i++ {
			f := strings.NewReader(fmt.Sprintf("%d.test.domain\n", i))
			io.Copy(w, f)
		}
	})

	return httptest.NewServer(mux)
}
