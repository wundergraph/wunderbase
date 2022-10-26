package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
)

func TestApi(t *testing.T) {
	fakeDB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := NewHandler(false, false, fakeDB.URL, fakeDB.URL+"/sdl", "/health", 0, 10000, 2000, cancel)

	fakeAPI := httptest.NewServer(handler)

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: fakeAPI.URL,
		Client: &http.Client{
			Jar:     httpexpect.NewJar(),
			Timeout: time.Second * 2,
		},
		Reporter: httpexpect.NewRequireReporter(t),
	})

	e.GET(fakeAPI.URL).Expect().Status(http.StatusOK).Body().Contains("GraphQL").Contains(fakeAPI.URL)
	e.GET(fakeAPI.URL + "/health").Expect().Status(http.StatusOK).Body().Equal("OK")
}
