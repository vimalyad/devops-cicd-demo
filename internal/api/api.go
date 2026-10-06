package api

import "net/http"

type PlanRequest struct {
	Replicas        int `json:"replicas"`
	BatchSize       int `json:"batch_size"`
	SecondsPerBatch int `json:"seconds_per_batch"`
}
type Plan struct {
	Batches          int `json:"batches"`
	LastBatchSize    int `json:"last_batch_size"`
	EstimatedSeconds int `json:"estimated_seconds"`
}

func New(version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, map[string]string{"application": "Rollout Planner", "student": "Vimal Kumar Yadav", "version": version})
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/plan", func(w http.ResponseWriter, r *http.Request) {
		var request PlanRequest
		if !decode(w, r, &request) {
			return
		}
		if request.Replicas < 1 || request.Replicas > 10000 || request.BatchSize < 1 || request.BatchSize > request.Replicas || request.SecondsPerBatch < 1 || request.SecondsPerBatch > 3600 {
			respond(w, http.StatusUnprocessableEntity, map[string]string{"error": "replicas must be 1..10000, batch_size 1..replicas, seconds_per_batch 1..3600"})
			return
		}
		batches := (request.Replicas + request.BatchSize - 1) / request.BatchSize
		respond(w, http.StatusOK, Plan{batches, request.Replicas - (batches-1)*request.BatchSize, batches * request.SecondsPerBatch})
	})
	return mux
}
