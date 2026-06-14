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
	"github.com/dataflow-operator/dataflow/pkg/k8snames"
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
	case filteredParts[0] == "dataflowcrons":
		h.handleDataFlowCrons(w, r, filteredParts[1:])
	case filteredParts[0] == "secrets":
		h.handleSecrets(w, r, filteredParts[1:])
	case filteredParts[0] == "logs":
		h.handleLogs(w, r, filteredParts[1:])
	case filteredParts[0] == "metrics":
		h.handleMetrics(w, r, filteredParts[1:])
	case filteredParts[0] == "status":
		h.handleStatus(w, r, filteredParts[1:])
	case filteredParts[0] == "runtime":
		h.handleRuntime(w, r)
	case filteredParts[0] == "prometheus":
		h.handlePrometheus(w, r, filteredParts[1:])
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
	kind := r.URL.Query().Get("kind")
	tailLines := r.URL.Query().Get("tailLines")
	follow := r.URL.Query().Get("follow") == "true"
	step := r.URL.Query().Get("step")
	runID := r.URL.Query().Get("runId")

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name required", http.StatusBadRequest)
		return
	}

	var pods *corev1.PodList
	var err error
	var containerName string

	if kind == "dataflowcron" {
		pods, containerName, err = h.listDataFlowCronPods(r, namespace, name, step, runID)
	} else {
		pods, err = h.server.k8sClient.CoreV1().Pods(namespace).List(r.Context(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("dataflow.dataflow.io/name=%s", name),
		})
		containerName = "processor"
	}
	if err != nil {
		h.server.logger.Error(err, "Failed to list pods")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pods.Items) == 0 && kind != "dataflowcron" {
		podName := k8snames.ProcessorDeployment(name)
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

	pod := selectNewestPod(pods.Items)
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}
	if containerName == "" {
		containerName = "processor"
	}

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

	persistence := checkpointPersistenceEnabled(df.Spec.CheckpointPersistence)
	status := map[string]interface{}{
		"phase":                  df.Status.Phase,
		"message":                df.Status.Message,
		"processedCount":         df.Status.ProcessedCount,
		"errorCount":             df.Status.ErrorCount,
		"lastProcessedTime":      df.Status.LastProcessedTime,
		"conditions":             df.Status.Conditions,
		"checkpointPersistence":  persistence,
		"sourceType":             df.Spec.Source.Type,
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

func (h *APIHandler) handleSecrets(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	switch r.Method {
	case "GET":
		if len(parts) == 0 {
			h.listSecrets(w, r, namespace)
		} else {
			h.getSecret(w, r, namespace, parts[0])
		}
	case "POST":
		h.createSecret(w, r, namespace)
	case "PUT":
		if len(parts) > 0 {
			h.updateSecret(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	case "DELETE":
		if len(parts) > 0 {
			h.deleteSecret(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// secretListItem is a safe representation of a Secret for listing (no sensitive values).
type secretListItem struct {
	Metadata *metav1.ObjectMeta `json:"metadata"`
	Keys     []string          `json:"keys"`
}

func (h *APIHandler) listSecrets(w http.ResponseWriter, r *http.Request, namespace string) {
	var list corev1.SecretList
	if err := h.server.client.List(r.Context(), &list, client.InNamespace(namespace)); err != nil {
		h.server.logger.Error(err, "Failed to list Secrets")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]secretListItem, 0, len(list.Items))
	for _, s := range list.Items {
		keys := make([]string, 0, len(s.Data)+len(s.StringData))
		for k := range s.Data {
			keys = append(keys, k)
		}
		for k := range s.StringData {
			if _, ok := s.Data[k]; !ok {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		meta := s.ObjectMeta.DeepCopy()
		items = append(items, secretListItem{Metadata: meta, Keys: keys})
	}

	json.NewEncoder(w).Encode(items)
}

// secretSafeView is a Secret representation with keys but masked values for sensitive fields.
type secretSafeView struct {
	Metadata   *metav1.ObjectMeta      `json:"metadata"`
	StringData map[string]string       `json:"stringData,omitempty"`
	Data       map[string]string       `json:"data,omitempty"`
}

var sensitiveKeys = map[string]bool{
	"password": true, "token": true, "connectionstring": true,
	"clientsecret": true, "username": true, "key": true,
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	return sensitiveKeys[lower] || strings.Contains(lower, "password") ||
		strings.Contains(lower, "token") || strings.Contains(lower, "secret")
}

func (h *APIHandler) getSecret(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var secret corev1.Secret
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build safe view: for sensitive keys mask values, for others return as-is
	stringData := make(map[string]string)
	for k, v := range secret.StringData {
		if isSensitiveKey(k) {
			stringData[k] = "****"
		} else {
			stringData[k] = v
		}
	}
	for k, v := range secret.Data {
		if _, ok := stringData[k]; ok {
			continue
		}
		if isSensitiveKey(k) {
			stringData[k] = "****"
		} else {
			stringData[k] = string(v)
		}
	}

	if len(stringData) == 0 {
		stringData = nil
	}

	view := secretSafeView{
		Metadata:   secret.ObjectMeta.DeepCopy(),
		StringData: stringData,
	}
	json.NewEncoder(w).Encode(view)
}

func (h *APIHandler) createSecret(w http.ResponseWriter, r *http.Request, namespace string) {
	var secret corev1.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secret.Namespace = namespace
	secret.APIVersion = "v1"
	secret.Kind = "Secret"
	if secret.Type == "" {
		secret.Type = corev1.SecretTypeOpaque
	}

	if err := h.server.client.Create(r.Context(), &secret); err != nil {
		h.server.logger.Error(err, "Failed to create Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(secret)
}

func (h *APIHandler) updateSecret(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var secret corev1.Secret
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updates corev1.Secret
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Merge updates: do not overwrite sensitive fields with masked placeholder "****"
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	for k, v := range updates.StringData {
		if v == "****" && isSensitiveKey(k) {
			if existing, ok := secret.StringData[k]; ok && existing != "" {
				continue
			}
			if existing, ok := secret.Data[k]; ok && len(existing) > 0 {
				secret.StringData[k] = string(existing)
				continue
			}
		}
		secret.StringData[k] = v
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	for k, v := range updates.Data {
		secret.Data[k] = v
	}
	if updates.Type != "" {
		secret.Type = updates.Type
	}

	if err := h.server.client.Update(r.Context(), &secret); err != nil {
		h.server.logger.Error(err, "Failed to update Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(secret)
}

func (h *APIHandler) deleteSecret(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var secret corev1.Secret
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.server.client.Delete(r.Context(), &secret); err != nil {
		h.server.logger.Error(err, "Failed to delete Secret")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
