/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package gui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
	"github.com/dataflow-operator/dataflow/pkg/k8snames"
)

func setupTestServerWithObjects(objects ...client.Object) (*Server, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

	return &Server{
		client:    fakeClient,
		k8sClient: nil,
		logger:    ctrl.Log.WithName("test"),
	}, nil
}

func TestAPIHandler_ListDataFlowCrons(t *testing.T) {
	dfc := &dataflowv1.DataFlowCron{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cron", Namespace: "default"},
		Spec: dataflowv1.DataFlowCronSpec{
			Schedule: "0 * * * *",
			DataFlowSpec: dataflowv1.DataFlowSpec{
				Source: dataflowv1.SourceSpec{Type: "postgresql"},
				Sink:   dataflowv1.SinkSpec{Type: "clickhouse"},
			},
		},
	}

	server, err := setupTestServerWithObjects(dfc)
	if err != nil {
		t.Fatalf("setup server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest(http.MethodGet, "/dataflowcrons?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", w.Code, w.Body.String())
	}

	var items []dataflowv1.DataFlowCron
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 1 || items[0].Name != "test-cron" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestAPIHandler_CreateDataFlowCron(t *testing.T) {
	server, err := setupTestServerWithObjects()
	if err != nil {
		t.Fatalf("setup server: %v", err)
	}
	handler := NewAPIHandler(server)

	body := dataflowv1.DataFlowCron{
		ObjectMeta: metav1.ObjectMeta{Name: "new-cron"},
		Spec: dataflowv1.DataFlowCronSpec{
			Schedule: "*/10 * * * *",
			DataFlowSpec: dataflowv1.DataFlowSpec{
				Source: dataflowv1.SourceSpec{Type: "postgresql"},
				Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
			},
		},
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/dataflowcrons?namespace=default", bytes.NewReader(raw))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status %d, body: %s", w.Code, w.Body.String())
	}

	var created dataflowv1.DataFlowCron
	if err := server.client.Get(ctx, types.NamespacedName{Namespace: "default", Name: "new-cron"}, &created); err != nil {
		t.Fatalf("get created: %v", err)
	}
}

func TestAPIHandler_TriggerDataFlowCron(t *testing.T) {
	dfc := &dataflowv1.DataFlowCron{
		ObjectMeta: metav1.ObjectMeta{Name: "cron-trigger", Namespace: "default"},
		Spec: dataflowv1.DataFlowCronSpec{
			Schedule: "0 * * * *",
			DataFlowSpec: dataflowv1.DataFlowSpec{
				Source: dataflowv1.SourceSpec{Type: "postgresql"},
				Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
			},
		},
	}
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: k8snames.CronJobName("cron-trigger"), Namespace: "default"},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{{
								Name:  "processor",
								Image: "processor:test",
							}},
						},
					},
				},
			},
		},
	}

	server, err := setupTestServerWithObjects(dfc, cronJob)
	if err != nil {
		t.Fatalf("setup server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest(http.MethodPost, "/dataflowcrons/cron-trigger/trigger?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status %d, body: %s", w.Code, w.Body.String())
	}

	var resp triggerResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.JobName == "" || resp.Namespace != "default" {
		t.Fatalf("unexpected response: %+v", resp)
	}

	var job batchv1.Job
	if err := server.client.Get(ctx, types.NamespacedName{Namespace: "default", Name: resp.JobName}, &job); err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Labels[dataFlowCronOwnerLabel] != "cron-trigger" {
		t.Fatalf("job labels: %+v", job.Labels)
	}
	if job.Annotations[cronJobInstantiateAnnotation] != "manual" {
		t.Fatalf("job annotations: %+v", job.Annotations)
	}
}

func TestAPIHandler_TriggerDataFlowCron_ConflictWhenRunning(t *testing.T) {
	dfc := &dataflowv1.DataFlowCron{
		ObjectMeta: metav1.ObjectMeta{Name: "cron-busy", Namespace: "default"},
		Spec: dataflowv1.DataFlowCronSpec{
			Schedule:          "0 * * * *",
			ConcurrencyPolicy: dataflowv1.DataFlowCronConcurrencyForbid,
			DataFlowSpec: dataflowv1.DataFlowSpec{
				Source: dataflowv1.SourceSpec{Type: "postgresql"},
				Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
			},
		},
	}
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: k8snames.CronJobName("cron-busy"), Namespace: "default"},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers:    []corev1.Container{{Name: "processor", Image: "processor:test"}},
						},
					},
				},
			},
		},
	}
	activeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dfc-cron-busy-running",
			Namespace: "default",
			Labels:    map[string]string{dataFlowCronOwnerLabel: "cron-busy"},
		},
		Status: batchv1.JobStatus{},
	}

	server, err := setupTestServerWithObjects(dfc, cronJob, activeJob)
	if err != nil {
		t.Fatalf("setup server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest(http.MethodPost, "/dataflowcrons/cron-busy/trigger?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409, body: %s", w.Code, w.Body.String())
	}
}

func TestAPIHandler_SuspendDataFlowCron(t *testing.T) {
	dfc := &dataflowv1.DataFlowCron{
		ObjectMeta: metav1.ObjectMeta{Name: "cron-suspend", Namespace: "default"},
		Spec: dataflowv1.DataFlowCronSpec{
			Schedule: "0 * * * *",
			DataFlowSpec: dataflowv1.DataFlowSpec{
				Source: dataflowv1.SourceSpec{Type: "postgresql"},
				Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
			},
		},
	}

	server, err := setupTestServerWithObjects(dfc)
	if err != nil {
		t.Fatalf("setup server: %v", err)
	}
	handler := NewAPIHandler(server)

	req := httptest.NewRequest(http.MethodPost, "/dataflowcrons/cron-suspend/suspend?namespace=default",
		bytes.NewReader([]byte(`{"suspend":true}`)))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", w.Code, w.Body.String())
	}

	var updated dataflowv1.DataFlowCron
	if err := server.client.Get(ctx, types.NamespacedName{Namespace: "default", Name: "cron-suspend"}, &updated); err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if updated.Spec.Suspend == nil || !*updated.Spec.Suspend {
		t.Fatalf("expected suspend true, got %+v", updated.Spec.Suspend)
	}
}

func TestManualCronJobName_TruncatesTo63(t *testing.T) {
	longName := "very-long-dataflow-cron-resource-name-that-exceeds-kubernetes-limits"
	name := manualCronJobName(longName, 1718380800)
	if len(name) > 63 {
		t.Fatalf("name length %d > 63: %q", len(name), name)
	}
}
