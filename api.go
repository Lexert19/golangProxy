package main

import (
	"encoding/json"
	"net/http"
	"context"
)

type Server struct {
	scheduler *Scheduler
	metrics   *MetricsTracker
	baseCtx   context.Context
}

func NewServer(scheduler *Scheduler, metrics *MetricsTracker, baseCtx context.Context) *Server {
    return &Server{
        scheduler: scheduler,
        metrics:   metrics,
        baseCtx:   baseCtx,
    }
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /start", s.handleStart)
	mux.HandleFunc("GET /status", s.handleStatus)
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	mux.HandleFunc("GET /results", s.handleResults)
	mux.HandleFunc("POST /cancel", s.handleCancel)
	return mux
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	started := s.scheduler.StartSingleRun(s.baseCtx)
	
	w.Header().Set("Content-Type", "application/json")
	if !started {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Check process is already running"})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Check process started"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.metrics.GetStatus())
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.metrics.GetMetrics())
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.metrics.GetResults())
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	cancelled := s.scheduler.CancelCurrent()
	w.Header().Set("Content-Type", "application/json")
	if !cancelled {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "No process currently running to cancel"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Check process cancelled"})
}