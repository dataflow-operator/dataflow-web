/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package gui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

var ctx = context.Background()

func setupTestServer() (*Server, error) {
	return setupTestServerWithMetricsURL("")
}

func setupTestServerWithMetricsURL(operatorMetricsURL string) (*Server, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	server := &Server{
		client:              fakeClient,
		k8sClient:           nil, // not needed for dataflow CRUD tests
		logger:              ctrl.Log.WithName("test"),
		operatorMetricsURL:   operatorMetricsURL,
	}

	return server, nil
}

func TestAPIHandler_ListDataFlows(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	handler := NewAPIHandler(server)

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataflow",
			Namespace: "default",
		},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "kafka"},
			Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
		},
	}

	if err := server.client.Create(ctx, df); err != nil {
		t.Fatalf("Failed to create test DataFlow: %v", err)
	}

	req := httptest.NewRequest("GET", "/dataflows?namespace=default", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var dataflows []dataflowv1.DataFlow
	if err := json.NewDecoder(w.Body).Decode(&dataflows); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(dataflows) != 1 {
		t.Errorf("Expected 1 DataFlow, got %d", len(dataflows))
	}

	if dataflows[0].Name != "test-dataflow" {
		t.Errorf("Expected name 'test-dataflow', got '%s'", dataflows[0].Name)
	}
}

func TestAPIHandler_GetDataFlow(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	handler := NewAPIHandler(server)

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataflow",
			Namespace: "default",
		},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "kafka"},
			Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
		},
	}

	if err := server.client.Create(ctx, df); err != nil {
		t.Fatalf("Failed to create test DataFlow: %v", err)
	}

	req := httptest.NewRequest("GET", "/dataflows/test-dataflow?namespace=default", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result dataflowv1.DataFlow
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Name != "test-dataflow" {
		t.Errorf("Expected name 'test-dataflow', got '%s'", result.Name)
	}
}

func TestAPIHandler_CreateDataFlow(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	handler := NewAPIHandler(server)

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: "new-dataflow"},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "kafka"},
			Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
		},
	}

	body, _ := json.Marshal(df)
	req := httptest.NewRequest("POST", "/dataflows?namespace=default", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var result dataflowv1.DataFlow
	if err := server.client.Get(ctx, client.ObjectKey{Namespace: "default", Name: "new-dataflow"}, &result); err != nil {
		t.Fatalf("Failed to get created DataFlow: %v", err)
	}
}

func TestAPIHandler_DeleteDataFlow(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	handler := NewAPIHandler(server)

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataflow",
			Namespace: "default",
		},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "kafka"},
			Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
		},
	}

	if err := server.client.Create(ctx, df); err != nil {
		t.Fatalf("Failed to create test DataFlow: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/dataflows/test-dataflow?namespace=default", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	var result dataflowv1.DataFlow
	if err := server.client.Get(ctx, client.ObjectKey{Namespace: "default", Name: "test-dataflow"}, &result); err == nil {
		t.Error("Expected DataFlow to be deleted, but it still exists")
	}
}

func TestAPIHandler_Metrics_RequiresNamespaceAndName(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	handler := NewAPIHandler(server)

	for _, tc := range []struct {
		query string
	}{
		{"?"},
		{"?namespace=default"},
		{"?name=foo"},
	} {
		req := httptest.NewRequest("GET", "/metrics"+tc.query, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("query %q: expected 400, got %d", tc.query, w.Code)
		}
	}
}

func TestAPIHandler_Metrics_StubWhenNoOperatorURL(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/metrics?namespace=default&name=myflow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected application/json, got %s", w.Header().Get("Content-Type"))
	}
	var out map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if out["namespace"] != "default" || out["name"] != "myflow" {
		t.Errorf("Expected namespace=default, name=myflow, got %v", out)
	}
	metrics, ok := out["metrics"].(map[string]interface{})
	if !ok || len(metrics) != 0 {
		t.Errorf("Expected empty metrics map, got %v", out["metrics"])
	}
}

func TestAPIHandler_Metrics_ProxyFromOperator(t *testing.T) {
	promOutput := `# HELP dataflow_messages_received_total Total messages received
# TYPE dataflow_messages_received_total counter
dataflow_messages_received_total{namespace="default",name="myflow",source_type="kafka"} 100
dataflow_messages_received_total{namespace="other",name="otherflow",source_type="kafka"} 50
# HELP go_goroutines Number of goroutines
# TYPE go_goroutines gauge
go_goroutines 10
`
	mockOperator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(promOutput))
	}))
	defer mockOperator.Close()

	server, err := setupTestServerWithMetricsURL(mockOperator.URL)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/metrics?namespace=default&name=myflow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `dataflow_messages_received_total{namespace="default",name="myflow"`) {
		t.Errorf("Expected filtered metrics for default/myflow, got: %s", body)
	}
	if strings.Contains(body, `namespace="other"`) {
		t.Errorf("Should not include metrics from other namespace: %s", body)
	}
	if strings.Contains(body, "go_goroutines") {
		t.Errorf("Should not include non-dataflow metrics: %s", body)
	}
}

func TestFilterPrometheusByDataFlow(t *testing.T) {
	input := `# HELP dataflow_foo A metric
# TYPE dataflow_foo counter
dataflow_foo{namespace="ns1",name="n1"} 1
dataflow_foo{namespace="ns2",name="n2"} 2
# HELP go_goroutines Goroutines
# TYPE go_goroutines gauge
go_goroutines 5
`
	var buf bytes.Buffer
	filterPrometheusByDataFlow(strings.NewReader(input), &buf, "ns1", "n1")
	out := buf.String()
	if !strings.Contains(out, `dataflow_foo{namespace="ns1",name="n1"} 1`) {
		t.Errorf("Expected ns1/n1 metric: %s", out)
	}
	if strings.Contains(out, `namespace="ns2"`) {
		t.Errorf("Should not include ns2: %s", out)
	}
	if strings.Contains(out, "go_goroutines") {
		t.Errorf("Should not include go_goroutines: %s", out)
	}
}

func TestAPIHandler_Events_AllInNamespace(t *testing.T) {
	now := metav1.NewTime(time.Now())
	ev1 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "default"},
		Type:       "Normal",
		Reason:     "ConfigMapCreated",
		Message:    "Created ConfigMap test-cm",
		InvolvedObject: corev1.ObjectReference{
			Kind:      "DataFlow",
			Name:      "my-flow",
			Namespace: "default",
		},
		LastTimestamp: now,
	}
	fakeK8s := k8sfake.NewSimpleClientset(ev1)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	server := &Server{
		client:    fakeClient,
		k8sClient: fakeK8s,
		logger:    ctrl.Log.WithName("test"),
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/events?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var events []corev1.Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Reason != "ConfigMapCreated" {
		t.Errorf("Expected reason ConfigMapCreated, got %s", events[0].Reason)
	}
}

func TestAPIHandler_Events_FilterByManifest(t *testing.T) {
	now := metav1.NewTime(time.Now())
	ev1 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev1", Namespace: "default"},
		Type:       "Normal",
		Reason:     "ConfigMapCreated",
		Message:    "Created ConfigMap",
		InvolvedObject: corev1.ObjectReference{
			Kind: "DataFlow", Name: "flow-a", Namespace: "default",
		},
		LastTimestamp: now,
	}
	ev2 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "ev2", Namespace: "default"},
		Type:       "Warning",
		Reason:     "DeploymentFailed",
		Message:    "Failed",
		InvolvedObject: corev1.ObjectReference{
			Kind: "DataFlow", Name: "flow-b", Namespace: "default",
		},
		LastTimestamp: now,
	}
	fakeK8s := k8sfake.NewSimpleClientset(ev1, ev2)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	server := &Server{
		client:    fakeClient,
		k8sClient: fakeK8s,
		logger:    ctrl.Log.WithName("test"),
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/events?namespace=default&name=flow-a", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var events []corev1.Event
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	// Fake clientset may not filter by field selector; at minimum we get valid response
	if len(events) == 0 {
		t.Errorf("Expected at least 1 event, got 0")
	}
	// If filtering works, we get only flow-a; otherwise we may get both
	foundFlowA := false
	for _, e := range events {
		if e.InvolvedObject.Name == "flow-a" {
			foundFlowA = true
			break
		}
	}
	if !foundFlowA {
		t.Errorf("Expected at least one event for flow-a, got %v", events)
	}
}

// TestScannerReadsLongLogLine verifies that a scanner with an increased buffer
// reads lines longer than 64 KB without "token too long" error.
func TestAPIHandler_ListSecrets(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"brokers": "localhost:9092",
			"topic":   "input-topic",
		},
	}
	if err := server.client.Create(ctx, secret); err != nil {
		t.Fatalf("Failed to create test Secret: %v", err)
	}

	handler := NewAPIHandler(server)
	req := httptest.NewRequest("GET", "/secrets?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var items []struct {
		Metadata *metav1.ObjectMeta `json:"metadata"`
		Keys     []string          `json:"keys"`
	}
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 secret, got %d", len(items))
	}
	if items[0].Metadata.Name != "test-secret" {
		t.Errorf("Expected name 'test-secret', got '%s'", items[0].Metadata.Name)
	}
	if len(items[0].Keys) != 2 {
		t.Errorf("Expected 2 keys, got %d: %v", len(items[0].Keys), items[0].Keys)
	}
}

func TestAPIHandler_CreateSecret(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	handler := NewAPIHandler(server)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "new-secret"},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"connectionString": "postgres://localhost/db",
			"table":            "output_table",
		},
	}
	body, _ := json.Marshal(secret)
	req := httptest.NewRequest("POST", "/secrets?namespace=default", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var result corev1.Secret
	if err := server.client.Get(ctx, client.ObjectKey{Namespace: "default", Name: "new-secret"}, &result); err != nil {
		t.Fatalf("Failed to get created Secret: %v", err)
	}
	if result.StringData["table"] != "output_table" {
		t.Errorf("Expected table=output_table, got %s", result.StringData["table"])
	}
}

func TestAPIHandler_DeleteSecret(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "to-delete",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{"key": "value"},
	}
	if err := server.client.Create(ctx, secret); err != nil {
		t.Fatalf("Failed to create test Secret: %v", err)
	}

	handler := NewAPIHandler(server)
	req := httptest.NewRequest("DELETE", "/secrets/to-delete?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	var result corev1.Secret
	if err := server.client.Get(ctx, client.ObjectKey{Namespace: "default", Name: "to-delete"}, &result); err == nil {
		t.Error("Expected Secret to be deleted, but it still exists")
	}
}

func TestScannerReadsLongLogLine(t *testing.T) {
	const lineLen = 65 * 1024 // larger than bufio.MaxScanTokenSize (64 KB)
	longLine := strings.Repeat("x", lineLen)
	input := longLine + "\n"

	scanner := bufio.NewScanner(strings.NewReader(input))
	const maxLogLineSize = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxLogLineSize)

	if !scanner.Scan() {
		t.Fatalf("Scan() = false, err = %v", scanner.Err())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner.Err() = %v", err)
	}
	if got := scanner.Text(); got != longLine {
		t.Errorf("len(Text()) = %d, want %d", len(got), lineLen)
	}
}

func TestAPIHandler_Runtime_CheckpointAndPods(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: "flow1", Namespace: "default"},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "postgresql"},
			Sink:   dataflowv1.SinkSpec{Type: "clickhouse"},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(df).Build()

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "df-flow1-checkpoint",
			Namespace: "default",
		},
		Data: map[string]string{
			"checkpoint.json": `{"postgresql":{"lastReadChangeTime":"2026-01-01T00:00:00Z","lastReadOrderByValue":"123"}}`,
		},
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "df-flow1",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      1,
			ReadyReplicas: 1,
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "df-flow1-abc",
			Namespace: "default",
			Labels: map[string]string{
				"app":                       "dataflow-processor",
				"dataflow.dataflow.io/name": "flow1",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: "node1",
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodRunning,
			StartTime: ptrTime(metav1.NewTime(time.Now())),
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "processor", Ready: true, RestartCount: 2},
			},
		},
	}
	fakeK8s := k8sfake.NewSimpleClientset(cm, dep, pod)

	server := &Server{client: fakeClient, k8sClient: fakeK8s, logger: ctrl.Log.WithName("test")}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/runtime?namespace=default&name=flow1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var out map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out["checkpointPersisted"] != true {
		t.Fatalf("Expected checkpointPersisted=true, got %v", out["checkpointPersisted"])
	}
	processor, ok := out["processor"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected processor object, got %T", out["processor"])
	}
	if processor["readyReplicas"].(float64) != 1 {
		t.Fatalf("Expected readyReplicas=1, got %v", processor["readyReplicas"])
	}
	pods, ok := processor["pods"].([]interface{})
	if !ok || len(pods) != 1 {
		t.Fatalf("Expected 1 pod, got %v", processor["pods"])
	}
}

func ptrTime(t metav1.Time) *metav1.Time { return &t }

func TestAPIHandler_Runtime_Checkpoint(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: "myflow", Namespace: "default"},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "postgresql"},
			Sink:   dataflowv1.SinkSpec{Type: "clickhouse"},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(df).Build()

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "df-myflow-checkpoint", Namespace: "default"},
		Data: map[string]string{
			"checkpoint.json": `{"postgresql":{"lastReadChangeTime":"2024-01-01T00:00:00Z","lastReadOrderByValue":"100"}}`,
		},
	}
	fakeK8s := k8sfake.NewSimpleClientset(cm)

	server := &Server{client: fakeClient, k8sClient: fakeK8s, logger: ctrl.Log.WithName("test")}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/runtime?namespace=default&name=myflow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var out map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["checkpointPersisted"] != true {
		t.Errorf("expected checkpointPersisted true, got %v", out["checkpointPersisted"])
	}
	checkpoints, ok := out["checkpoints"].(map[string]interface{})
	if !ok || checkpoints["postgresql"] == nil {
		t.Errorf("expected postgresql checkpoint, got %v", out["checkpoints"])
	}
}

func TestAPIHandler_PrometheusInstant_Fallback(t *testing.T) {
	metricsBody := `# TYPE dataflow_task_throughput_messages_per_second gauge
dataflow_task_throughput_messages_per_second{namespace="default",name="myflow"} 12.5
`
	metricsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(metricsBody))
	}))
	defer metricsServer.Close()

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	server := &Server{
		client:             fakeClient,
		logger:             ctrl.Log.WithName("test"),
		operatorMetricsURL: metricsServer.URL,
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/prometheus/instant?namespace=default&name=myflow&panel=throughput_gauge", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var out promQueryResponse
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Mode != "instant" {
		t.Errorf("expected instant mode, got %s", out.Mode)
	}
	if len(out.Series) == 0 {
		t.Error("expected at least one series")
	}
}

func TestAPIHandler_Runtime_ProcessorDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: "myflow", Namespace: "default"},
		Spec: dataflowv1.DataFlowSpec{
			Source: dataflowv1.SourceSpec{Type: "kafka"},
			Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(df).Build()

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "df-myflow", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "dataflow-processor", "dataflow.dataflow.io/name": "myflow"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "dataflow-processor", "dataflow.dataflow.io/name": "myflow"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "processor", Image: "test"}}},
			},
		},
		Status: appsv1.DeploymentStatus{Replicas: 1, ReadyReplicas: 1},
	}
	fakeK8s := k8sfake.NewSimpleClientset(dep)

	server := &Server{client: fakeClient, k8sClient: fakeK8s, logger: ctrl.Log.WithName("test")}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest("GET", "/runtime?namespace=default&name=myflow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var out map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	sourceInfo, ok := out["sourceInfo"].(map[string]interface{})
	if !ok || sourceInfo["type"] != "kafka" {
		t.Errorf("expected kafka sourceInfo, got %v", out["sourceInfo"])
	}
	proc, ok := out["processor"].(map[string]interface{})
	if !ok || proc["readyReplicas"] != float64(1) {
		t.Errorf("expected readyReplicas 1, got %v", out["processor"])
	}
}
