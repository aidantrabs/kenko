package prommetrics

import (
	"testing"

	"github.com/aidantrabs/kenko"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestReporter(t *testing.T) *Reporter {
	t.Helper()
	reg := prometheus.NewPedanticRegistry()
	return New(WithRegistry(reg))
}

func TestReportCheck_Healthy(t *testing.T) {
	r := newTestReporter(t)
	r.ReportCheck("api", kenko.StatusHealthy, 0.042)

	reg := r.registerer.(*prometheus.Registry)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	if len(families) == 0 {
		t.Fatal("expected at least one metric family")
	}

	found := false
	for _, f := range families {
		if f.GetName() == "kenko_target_up" {
			found = true
			for _, m := range f.GetMetric() {
				if m.GetGauge().GetValue() != 1 {
					t.Errorf("target_up = %v, want 1", m.GetGauge().GetValue())
				}
			}
		}
	}
	if !found {
		t.Error("kenko_target_up metric not found")
	}
}

func TestReportCheck_Unhealthy(t *testing.T) {
	r := newTestReporter(t)
	r.ReportCheck("api", kenko.StatusUnhealthy, 5.0)

	reg := r.registerer.(*prometheus.Registry)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, f := range families {
		if f.GetName() == "kenko_target_up" {
			for _, m := range f.GetMetric() {
				if m.GetGauge().GetValue() != 0 {
					t.Errorf("target_up = %v, want 0", m.GetGauge().GetValue())
				}
			}
		}
	}
}

func TestNew_WithNamespace(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	r := New(WithRegistry(reg), WithNamespace("myapp"))
	r.ReportCheck("api", kenko.StatusHealthy, 0.1)

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	found := false
	for _, f := range families {
		if f.GetName() == "myapp_kenko_target_up" {
			found = true
		}
	}
	if !found {
		names := make([]string, 0, len(families))
		for _, f := range families {
			names = append(names, f.GetName())
		}
		t.Errorf("myapp_kenko_target_up not found, got: %v", names)
	}
}
