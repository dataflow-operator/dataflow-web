/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
*/

package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dataflow-operator/dataflow-web/internal/gui/promparse"
)

type promQueryResponse struct {
	Mode     string           `json:"mode"`
	Degraded bool             `json:"degraded,omitempty"`
	Message  string           `json:"message,omitempty"`
	Series   []promTimeSeries `json:"series"`
}

type promTimeSeries struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

func (h *APIHandler) handlePrometheus(w http.ResponseWriter, r *http.Request, parts []string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if len(parts) == 0 {
		http.Error(w, "prometheus subpath required", http.StatusBadRequest)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	panel := r.URL.Query().Get("panel")
	if namespace == "" || name == "" || panel == "" {
		http.Error(w, "namespace, name and panel required", http.StatusBadRequest)
		return
	}
	if !isValidPanel(panel) {
		http.Error(w, "invalid panel", http.StatusBadRequest)
		return
	}

	switch parts[0] {
	case "range":
		h.handlePrometheusRange(w, r, namespace, name, panel)
	case "instant":
		h.handlePrometheusInstant(w, r, namespace, name, panel)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *APIHandler) handlePrometheusRange(w http.ResponseWriter, r *http.Request, namespace, name, panel string) {
	query, err := panelQuery(panel, namespace, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	end := time.Now().Unix()
	start := end - 3600
	step := "60"
	if v := r.URL.Query().Get("start"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			start = n
		}
	}
	if v := r.URL.Query().Get("end"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			end = n
		}
	}
	if v := r.URL.Query().Get("step"); v != "" {
		step = v
	}

	if h.server.prometheusURL != "" {
		resp, err := h.queryPrometheusRange(r, query, start, end, step)
		if err == nil {
			writePromResponse(w, promQueryResponse{Mode: "prometheus", Series: resp})
			return
		}
		h.server.logger.Info("Prometheus range query failed, falling back to instant", "error", err)
	}

	series, err := h.instantSeriesFromOperator(r, namespace, name, panel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writePromResponse(w, promQueryResponse{
		Mode:     "instant",
		Degraded: h.server.prometheusURL != "",
		Message:  degradedMessage(h.server.prometheusURL),
		Series:   series,
	})
}

func (h *APIHandler) handlePrometheusInstant(w http.ResponseWriter, r *http.Request, namespace, name, panel string) {
	if h.server.prometheusURL != "" {
		query, err := panelQuery(panel, namespace, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		series, err := h.queryPrometheusInstant(r, query)
		if err == nil {
			writePromResponse(w, promQueryResponse{Mode: "prometheus", Series: series})
			return
		}
		h.server.logger.Info("Prometheus instant query failed, falling back", "error", err)
	}

	series, err := h.instantSeriesFromOperator(r, namespace, name, panel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writePromResponse(w, promQueryResponse{
		Mode:     "instant",
		Degraded: true,
		Message:  degradedMessage(h.server.prometheusURL),
		Series:   series,
	})
}

func degradedMessage(prometheusURL string) string {
	if prometheusURL == "" {
		return "Historical metrics require PROMETHEUS_URL; showing current values from operator"
	}
	return "Prometheus unavailable; showing current values from operator"
}

func writePromResponse(w http.ResponseWriter, resp promQueryResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) queryPrometheusRange(r *http.Request, query string, start, end int64, step string) ([]promTimeSeries, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start, 10))
	params.Set("end", strconv.FormatInt(end, 10))
	params.Set("step", step)

	body, err := h.prometheusGET(r, "/api/v1/query_range", params)
	if err != nil {
		return nil, err
	}
	return parsePrometheusMatrix(body)
}

func (h *APIHandler) queryPrometheusInstant(r *http.Request, query string) ([]promTimeSeries, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", strconv.FormatInt(time.Now().Unix(), 10))

	body, err := h.prometheusGET(r, "/api/v1/query", params)
	if err != nil {
		return nil, err
	}
	return parsePrometheusVector(body)
}

func (h *APIHandler) prometheusGET(r *http.Request, path string, params url.Values) ([]byte, error) {
	base := strings.TrimSuffix(h.server.prometheusURL, "/")
	reqURL := base + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if h.server.prometheusBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.server.prometheusBearerToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func parsePrometheusMatrix(body []byte) ([]promTimeSeries, error) {
	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed")
	}
	out := make([]promTimeSeries, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		out = append(out, promTimeSeries{Metric: r.Metric, Values: r.Values})
	}
	return out, nil
}

func parsePrometheusVector(body []byte) ([]promTimeSeries, error) {
	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed")
	}
	out := make([]promTimeSeries, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		out = append(out, promTimeSeries{Metric: r.Metric, Values: [][]interface{}{r.Value}})
	}
	return out, nil
}

func (h *APIHandler) instantSeriesFromOperator(r *http.Request, namespace, name, panel string) ([]promTimeSeries, error) {
	if h.server.operatorMetricsURL == "" {
		return nil, fmt.Errorf("operator metrics URL not configured")
	}
	metricsURL := strings.TrimSuffix(h.server.operatorMetricsURL, "/") + "/metrics"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, metricsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("operator metrics returned %d", resp.StatusCode)
	}

	samples, err := promparse.ParseText(resp.Body)
	if err != nil {
		return nil, err
	}
	samples = promparse.FilterByDataFlow(samples, namespace, name)

	now := float64(time.Now().Unix())
	byLabels := make(map[string]promTimeSeries)
	for _, s := range samples {
		if !sampleMatchesPanel(panel, s.Name) {
			continue
		}
		key := labelsKey(s.Labels)
		ts, ok := byLabels[key]
		if !ok {
			ts = promTimeSeries{Metric: s.Labels}
		}
		ts.Metric["__name__"] = s.Name
		ts.Values = append(ts.Values, []interface{}{now, fmt.Sprintf("%g", s.Value)})
		byLabels[key] = ts
	}

	out := make([]promTimeSeries, 0, len(byLabels))
	for _, ts := range byLabels {
		if ts.Metric == nil {
			ts.Metric = map[string]string{}
		}
		out = append(out, ts)
	}
	return out, nil
}

func labelsKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
