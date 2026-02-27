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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

var ctx = context.Background()

func setupTestServer() (*Server, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	server := &Server{
		client:    fakeClient,
		k8sClient: nil, // not needed for dataflow CRUD tests
		logger:    ctrl.Log.WithName("test"),
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

// TestScannerReadsLongLogLine verifies that a scanner with an increased buffer
// reads lines longer than 64 KB without "token too long" error.
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
