/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
*/

package promparse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Sample is a single Prometheus metric sample parsed from text exposition format.
type Sample struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	Type   string            `json:"type,omitempty"`
}

// ParseText parses Prometheus text exposition format into samples (counters and gauges only).
func ParseText(r io.Reader) ([]Sample, error) {
	scanner := bufio.NewScanner(r)
	const maxLineSize = 64 * 1024
	buf := make([]byte, 0, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	var currentType string
	var samples []Sample

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# TYPE ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentType = parts[3]
			}
			continue
		}
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if !strings.HasPrefix(line, "dataflow_") {
			continue
		}

		name, labels, value, err := parseMetricLine(line)
		if err != nil {
			continue
		}
		samples = append(samples, Sample{
			Name:   name,
			Labels: labels,
			Value:  value,
			Type:   currentType,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan metrics: %w", err)
	}
	return samples, nil
}

func parseMetricLine(line string) (string, map[string]string, float64, error) {
	open := strings.Index(line, "{")
	if open < 0 {
		space := strings.LastIndex(line, " ")
		if space < 0 {
			return "", nil, 0, fmt.Errorf("invalid line")
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(line[space+1:]), 64)
		return line[:space], map[string]string{}, val, err
	}
	close := strings.Index(line[open:], "}")
	if close < 0 {
		return "", nil, 0, fmt.Errorf("unclosed labels")
	}
	close += open
	name := line[:open]
	labels := parseLabels(line[open+1 : close])
	val, err := strconv.ParseFloat(strings.TrimSpace(line[close+1:]), 64)
	return name, labels, val, err
}

func parseLabels(s string) map[string]string {
	labels := make(map[string]string)
	if s == "" {
		return labels
	}
	for _, part := range splitLabels(s) {
		eq := strings.Index(part, "=")
		if eq < 0 {
			continue
		}
		key := part[:eq]
		val := strings.Trim(part[eq+1:], `"`)
		labels[key] = val
	}
	return labels
}

func splitLabels(s string) []string {
	var parts []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
			cur.WriteByte(c)
			continue
		}
		if c == ',' && !inQuotes {
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// FilterByDataFlow keeps samples whose namespace and name labels match.
func FilterByDataFlow(samples []Sample, namespace, name string) []Sample {
	out := make([]Sample, 0, len(samples))
	for _, s := range samples {
		if s.Labels["namespace"] == namespace && s.Labels["name"] == name {
			out = append(out, s)
		}
	}
	return out
}
