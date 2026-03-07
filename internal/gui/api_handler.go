/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

// APIHandler handles API requests.
type APIHandler struct {
	server *Server
}

// NewAPIHandler creates a new API handler.
func NewAPIHandler(server *Server) *APIHandler {
	return &APIHandler{server: server}
}

// ServeHTTP handles HTTP requests.
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	filteredParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filteredParts = append(filteredParts, part)
		}
	}

	if len(filteredParts) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	switch {
	case filteredParts[0] == "dataflows":
		h.handleDataFlows(w, r, filteredParts[1:])
	case filteredParts[0] == "logs":
		h.handleLogs(w, r, filteredParts[1:])
	case filteredParts[0] == "metrics":
		h.handleMetrics(w, r, filteredParts[1:])
	case filteredParts[0] == "status":
		h.handleStatus(w, r, filteredParts[1:])
	case filteredParts[0] == "namespaces":
		h.handleNamespaces(w, r)
	case filteredParts[0] == "events":
		h.handleEvents(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *APIHandler) handleDataFlows(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	switch r.Method {
	case "GET":
		if len(parts) == 0 {
			h.listDataFlows(w, r, namespace)
		} else {
			h.getDataFlow(w, r, namespace, parts[0])
		}
	case "POST":
		h.createDataFlow(w, r, namespace)
	case "PUT":
		if len(parts) > 0 {
			h.updateDataFlow(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	case "DELETE":
		if len(parts) > 0 {
			h.deleteDataFlow(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) listDataFlows(w http.ResponseWriter, r *http.Request, namespace string) {
	var list dataflowv1.DataFlowList
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := h.server.client.List(r.Context(), &list, opts...); err != nil {
		h.server.logger.Error(err, "Failed to list DataFlows")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list.Items)
}

func (h *APIHandler) getDataFlow(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(df)
}

func (h *APIHandler) createDataFlow(w http.ResponseWriter, r *http.Request, namespace string) {
	var df dataflowv1.DataFlow
	if err := json.NewDecoder(r.Body).Decode(&df); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	df.Namespace = namespace
	df.APIVersion = "dataflow.dataflow.io/v1"
	df.Kind = "DataFlow"

	if err := h.server.client.Create(r.Context(), &df); err != nil {
		h.server.logger.Error(err, "Failed to create DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(df)
}

func (h *APIHandler) updateDataFlow(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updates dataflowv1.DataFlow
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	df.Spec = updates.Spec

	if err := h.server.client.Update(r.Context(), &df); err != nil {
		h.server.logger.Error(err, "Failed to update DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(df)
}

func (h *APIHandler) deleteDataFlow(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.server.client.Delete(r.Context(), &df); err != nil {
		h.server.logger.Error(err, "Failed to delete DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) handleLogs(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	tailLines := r.URL.Query().Get("tailLines")
	follow := r.URL.Query().Get("follow") == "true"

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name required", http.StatusBadRequest)
		return
	}

	pods, err := h.server.k8sClient.CoreV1().Pods(namespace).List(r.Context(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("dataflow.dataflow.io/name=%s", name),
	})
	if err != nil {
		h.server.logger.Error(err, "Failed to list pods")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pods.Items) == 0 {
		podName := fmt.Sprintf("dataflow-%s", name)
		pod, err := h.server.k8sClient.CoreV1().Pods(namespace).Get(r.Context(), podName, metav1.GetOptions{})
		if err != nil {
			pods, err = h.server.k8sClient.CoreV1().Pods(namespace).List(r.Context(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=dataflow-processor,dataflow.dataflow.io/name=%s", name),
			})
			if err != nil || len(pods.Items) == 0 {
				http.Error(w, "Pod not found", http.StatusNotFound)
				return
			}
		} else {
			pods.Items = []corev1.Pod{*pod}
		}
	}

	if len(pods.Items) == 0 {
		http.Error(w, "No pods found", http.StatusNotFound)
		return
	}

	pod := pods.Items[0]
	containerName := "processor"

	opts := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    follow,
	}

	if tailLines != "" {
		if tail, err := strconv.ParseInt(tailLines, 10, 64); err == nil {
			opts.TailLines = &tail
		}
	} else {
		tail := int64(100)
		opts.TailLines = &tail
	}

	req := h.server.k8sClient.CoreV1().Pods(namespace).GetLogs(pod.Name, opts)
	logs, err := req.Stream(r.Context())
	if err != nil {
		h.server.logger.Error(err, "Failed to get logs")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	if follow {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		scanner := bufio.NewScanner(logs)
		const maxLogLineSize = 1024 * 1024 // 1 MB - log lines can be long (dumps, base64)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, maxLogLineSize)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(w, "data: %s\n\n", line)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			h.server.logger.Error(err, "Error reading logs")
		}
	} else {
		io.Copy(w, logs)
	}
}

func (h *APIHandler) handleMetrics(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name required", http.StatusBadRequest)
		return
	}

	if h.server.operatorMetricsURL == "" {
		metrics := map[string]interface{}{
			"namespace": namespace,
			"name":      name,
			"metrics":   map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(metrics)
		return
	}

	metricsURL := strings.TrimSuffix(h.server.operatorMetricsURL, "/") + "/metrics"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, metricsURL, nil)
	if err != nil {
		h.server.logger.Error(err, "Failed to create metrics request")
		http.Error(w, "Failed to create metrics request", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.server.logger.Error(err, "Failed to fetch metrics from operator")
		http.Error(w, "Failed to fetch metrics from operator", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.server.logger.Info("Operator metrics returned non-200", "status", resp.StatusCode)
		http.Error(w, fmt.Sprintf("Operator metrics returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	filterPrometheusByDataFlow(resp.Body, w, namespace, name)
}

// filterPrometheusByDataFlow reads Prometheus text format from src and writes to dst
// only lines for dataflow_* metrics with namespace and name labels matching the given DataFlow.
func filterPrometheusByDataFlow(src io.Reader, dst io.Writer, namespace, name string) {
	nsMatch := `namespace="` + namespace + `"`
	nameMatch := `name="` + name + `"`
	scanner := bufio.NewScanner(src)
	const maxLineSize = 64 * 1024
	buf := make([]byte, 0, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			if strings.Contains(line, "dataflow_") {
				fmt.Fprintln(dst, line)
			}
			continue
		}
		if strings.HasPrefix(line, "dataflow_") && strings.Contains(line, nsMatch) && strings.Contains(line, nameMatch) {
			fmt.Fprintln(dst, line)
		}
	}
}

func (h *APIHandler) handleStatus(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name required", http.StatusBadRequest)
		return
	}

	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"phase":             df.Status.Phase,
		"message":           df.Status.Message,
		"processedCount":    df.Status.ProcessedCount,
		"errorCount":        df.Status.ErrorCount,
		"lastProcessedTime": df.Status.LastProcessedTime,
	}

	json.NewEncoder(w).Encode(status)
}

func (h *APIHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	name := r.URL.Query().Get("name")

	fieldSelector := "involvedObject.kind=DataFlow"
	if name != "" {
		fieldSelector += ",involvedObject.name=" + name
	}

	events, err := h.server.k8sClient.CoreV1().Events(namespace).List(r.Context(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		h.server.logger.Error(err, "Failed to list events")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := events.Items
	sort.Slice(items, func(i, j int) bool {
		return items[j].LastTimestamp.Before(&items[i].LastTimestamp)
	})

	const maxEvents = 100
	if len(items) > maxEvents {
		items = items[:maxEvents]
	}

	json.NewEncoder(w).Encode(items)
}

func (h *APIHandler) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespaces, err := h.server.k8sClient.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		h.server.logger.Error(err, "Failed to list namespaces")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	namespaceNames := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		namespaceNames = append(namespaceNames, ns.Name)
	}

	json.NewEncoder(w).Encode(namespaceNames)
}
