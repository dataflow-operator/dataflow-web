/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
*/

package gui

import (
	"context"
	"encoding/json"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
	"github.com/dataflow-operator/dataflow/pkg/k8snames"
)

const checkpointJSONKey = "checkpoint.json"

func (h *APIHandler) handleRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
		h.server.logger.Error(err, "Failed to get DataFlow for runtime")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := buildRuntimeResponse(r.Context(), h, &df)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.server.logger.Error(err, "Failed to encode runtime response")
	}
}

func buildRuntimeResponse(ctx context.Context, h *APIHandler, df *dataflowv1.DataFlow) map[string]interface{} {
	namespace := df.Namespace
	name := df.Name

	persistence := checkpointPersistenceEnabled(df.Spec.CheckpointPersistence)
	cmName := k8snames.ProcessorCheckpointConfigMap(name)

	out := map[string]interface{}{
		"namespace":              namespace,
		"name":                   name,
		"sourceType":             df.Spec.Source.Type,
		"checkpointPersistence":  persistence,
		"checkpointConfigMap":    cmName,
		"checkpoints":            map[string]interface{}{},
		"checkpointPersisted":    false,
		"checkpointMessage":      "",
		"sourceInfo":             buildSourceInfo(&df.Spec.Source),
		"conditions":             df.Status.Conditions,
		"processor":              map[string]interface{}{},
	}

	if !persistence {
		out["checkpointMessage"] = "Checkpoint persistence is disabled in spec"
	} else if h.server.k8sClient == nil {
		out["checkpointMessage"] = "Kubernetes client unavailable"
	} else {
		cm, err := h.server.k8sClient.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			out["checkpointMessage"] = "Checkpoint ConfigMap not found yet"
		} else if err != nil {
			out["checkpointMessage"] = err.Error()
		} else {
			raw, ok := cm.Data[checkpointJSONKey]
			if !ok || raw == "" {
				out["checkpointMessage"] = "No checkpoint data written yet"
			} else {
				var checkpoints map[string]interface{}
				if err := json.Unmarshal([]byte(raw), &checkpoints); err != nil {
					out["checkpointMessage"] = "Failed to parse checkpoint.json"
					out["checkpointRaw"] = raw
				} else {
					out["checkpoints"] = checkpoints
					out["checkpointPersisted"] = true
					out["checkpointMessage"] = ""
				}
			}
		}
	}

	out["processor"] = loadProcessorInfo(ctx, h, namespace, name)
	return out
}

func checkpointPersistenceEnabled(v *bool) bool {
	return v == nil || *v
}

func buildSourceInfo(src *dataflowv1.SourceSpec) map[string]interface{} {
	info := map[string]interface{}{
		"type": src.Type,
	}
	switch src.Type {
	case "kafka":
		if cfg, err := src.GetKafkaConfig(); err == nil && cfg != nil {
			info["topic"] = cfg.Topic
			info["consumerGroup"] = cfg.ConsumerGroup
			info["brokers"] = cfg.Brokers
			info["checkpointNote"] = "Kafka offsets are managed by the consumer group, not the checkpoint ConfigMap"
		}
	case "nessie":
		if cfg, err := src.GetNessieConfig(); err == nil && cfg != nil {
			info["branch"] = cfg.Branch
			info["namespace"] = cfg.Namespace
			info["table"] = cfg.Table
			if cfg.IncrementalBySnapshot != nil {
				info["incrementalBySnapshot"] = *cfg.IncrementalBySnapshot
			}
			if cfg.SnapshotCheckpoints != nil {
				info["snapshotCheckpoints"] = *cfg.SnapshotCheckpoints
			}
			info["startSnapshotID"] = cfg.StartSnapshotID
		}
	default:
		info["checkpointNote"] = "Polling source; position stored in checkpoint ConfigMap when persistence is enabled"
	}
	return info
}

func loadProcessorInfo(ctx context.Context, h *APIHandler, namespace, dataflowName string) map[string]interface{} {
	result := map[string]interface{}{
		"deploymentName": k8snames.ProcessorDeployment(dataflowName),
		"replicas":       int32(0),
		"readyReplicas":  int32(0),
		"pods":           []map[string]interface{}{},
	}

	if h.server.k8sClient == nil {
		return result
	}

	depName := k8snames.ProcessorDeployment(dataflowName)
	dep, err := h.server.k8sClient.AppsV1().Deployments(namespace).Get(ctx, depName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return result
	}
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	fillDeploymentInfo(result, dep)

	labelSelector := "dataflow.dataflow.io/name=" + dataflowName + ",app=dataflow-processor"
	pods, err := h.server.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		result["podsError"] = err.Error()
		return result
	}

	podList := make([]map[string]interface{}, 0, len(pods.Items))
	for _, p := range pods.Items {
		restarts := int32(0)
		ready := false
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
			if cs.Ready {
				ready = true
			}
		}
		podList = append(podList, map[string]interface{}{
			"name":      p.Name,
			"phase":     p.Status.Phase,
			"ready":     ready,
			"restarts":  restarts,
			"node":      p.Spec.NodeName,
			"startTime": p.Status.StartTime,
		})
	}
	result["pods"] = podList
	return result
}

func fillDeploymentInfo(result map[string]interface{}, dep *appsv1.Deployment) {
	result["replicas"] = dep.Status.Replicas
	result["readyReplicas"] = dep.Status.ReadyReplicas
	result["availableReplicas"] = dep.Status.AvailableReplicas
	result["updatedReplicas"] = dep.Status.UpdatedReplicas
	conds := make([]map[string]interface{}, 0, len(dep.Status.Conditions))
	for _, c := range dep.Status.Conditions {
		conds = append(conds, map[string]interface{}{
			"type":    c.Type,
			"status":  c.Status,
			"reason":  c.Reason,
			"message": c.Message,
		})
	}
	result["conditions"] = conds
}
