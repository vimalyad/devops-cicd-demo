package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlanBatches(t *testing.T) {
	cases := []struct {
		name, body string
		expected   Plan
	}{
		{"partial final batch", `{"replicas":7,"batch_size":3,"seconds_per_batch":20}`, Plan{3, 1, 60}},
		{"even batches", `{"replicas":6,"batch_size":2,"seconds_per_batch":10}`, Plan{3, 2, 30}},
		{"single replica", `{"replicas":1,"batch_size":1,"seconds_per_batch":5}`, Plan{1, 1, 5}},
		{"upper bounds", `{"replicas":10000,"batch_size":1,"seconds_per_batch":3600}`, Plan{10000, 1, 36000000}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			New("test").ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/plan", strings.NewReader(tc.body)))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status=%d: %s", recorder.Code, recorder.Body.String())
			}
			var result Plan
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result != tc.expected {
				t.Fatalf("got %+v, want %+v", result, tc.expected)
			}
		})
	}
}

func TestRejectInvalidPlans(t *testing.T) {
	for _, body := range []string{
		`{"replicas":0,"batch_size":1,"seconds_per_batch":5}`,
		`{"replicas":4,"batch_size":0,"seconds_per_batch":5}`,
		`{"replicas":4,"batch_size":5,"seconds_per_batch":5}`,
		`{"replicas":4,"batch_size":2,"seconds_per_batch":0}`,
		`{"replicas":10001,"batch_size":1,"seconds_per_batch":5}`,
		`{"replicas":4,"batch_size":1,"seconds_per_batch":3601}`,
	} {
		r := httptest.NewRecorder()
		New("test").ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/plan", strings.NewReader(body)))
		if r.Code != http.StatusUnprocessableEntity {
			t.Fatalf("body=%s status=%d", body, r.Code)
		}
	}
}

func TestRejectMalformedJSON(t *testing.T) {
	for _, body := range []string{`{`, `{"unexpected":1}`, `{} {}`, `{"replicas":"seven"}`, strings.Repeat(" ", 4097) + "{}"} {
		r := httptest.NewRecorder()
		New("test").ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/plan", strings.NewReader(body)))
		if r.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", r.Code)
		}
	}
}

func TestHealthAndVersion(t *testing.T) {
	for path, expected := range map[string]string{"/healthz": `"status":"ok"`, "/": `"version":"test-sha"`} {
		r := httptest.NewRecorder()
		New("test-sha").ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != 200 || !strings.Contains(r.Body.String(), expected) {
			t.Fatalf("%s: %d %s", path, r.Code, r.Body.String())
		}
	}
}

func TestWrongMethodAndUnknownRoute(t *testing.T) {
	for path, status := range map[string]int{"/api/plan": 405, "/missing": 404} {
		r := httptest.NewRecorder()
		New("test").ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != status {
			t.Fatalf("%s: got %d want %d", path, r.Code, status)
		}
	}
}
