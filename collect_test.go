package main

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCollectMetricsInvariants(t *testing.T) {
	ms := collectMetrics()
	if len(ms) == 0 {
		t.Fatal("采集结果为空")
	}
	seen := map[string]bool{}
	for _, m := range ms {
		if !strings.HasPrefix(m.name, "jiankong_") {
			t.Errorf("指标名 %q 未以 jiankong_ 开头", m.name)
		}
		if m.typ != "gauge" && m.typ != "counter" {
			t.Errorf("指标 %s 类型 %q 非 gauge/counter", m.name, m.typ)
		}
		if m.help == "" {
			t.Errorf("指标 %s 缺少 help", m.name)
		}
		if math.IsNaN(m.value) || math.IsInf(m.value, 0) {
			t.Errorf("指标 %s 值为非有限数 %v", m.name, m.value)
		}
		key := m.name + labelString(m.labels)
		if seen[key] {
			t.Errorf("重复样本 %s", key)
		}
		seen[key] = true
		// 百分比类指标应在 0~100 之间
		if strings.HasSuffix(m.name, "_percent") && (m.value < 0 || m.value > 100) {
			t.Errorf("指标 %s 百分比 %v 超出 0~100", m.name, m.value)
		}
		// 总量类指标应大于 0
		if strings.HasSuffix(m.name, "_total_bytes") && m.value <= 0 {
			t.Errorf("指标 %s 总量 %v 应大于 0", m.name, m.value)
		}
		// 计数器只增不减
		if m.typ == "counter" && m.value < 0 {
			t.Errorf("计数器 %s 值为负 %v", m.name, m.value)
		}
	}

	// build_info 恒为 1,且 version 标签与构建版本一致(默认 dev)
	found := false
	for _, m := range ms {
		if m.name == "jiankong_build_info" {
			found = true
			if m.value != 1 {
				t.Errorf("jiankong_build_info = %v, want 1", m.value)
			}
			if m.labels["version"] != version {
				t.Errorf("jiankong_build_info version 标签 = %q, want %q", m.labels["version"], version)
			}
		}
	}
	if !found {
		t.Error("缺少 jiankong_build_info")
	}
}

func TestCollectMetricsHasCoreNames(t *testing.T) {
	ms := collectMetrics()
	names := map[string]bool{}
	for _, m := range ms {
		names[m.name] = true
	}
	// 核心指标:与平台无关,在任何机器上都应存在
	required := []string{
		"jiankong_build_info",
		"jiankong_host_info",
		"jiankong_uptime_seconds",
		"jiankong_boot_time_seconds",
		"jiankong_cpu_usage_percent",
		"jiankong_memory_used_bytes",
		"jiankong_memory_total_bytes",
		"jiankong_memory_usage_percent",
		"jiankong_load1",
		"jiankong_load5",
		"jiankong_load15",
		"jiankong_disk_used_bytes",
		"jiankong_disk_total_bytes",
		"jiankong_disk_usage_percent",
		"jiankong_network_receive_bytes_total",
		"jiankong_network_transmit_bytes_total",
	}
	for _, n := range required {
		if !names[n] {
			t.Errorf("缺少核心指标 %s", n)
		}
	}
	// 负载值不应为负
	for _, m := range ms {
		if (m.name == "jiankong_load1" || m.name == "jiankong_load5" || m.name == "jiankong_load15") && m.value < 0 {
			t.Errorf("指标 %s 值为负 %v", m.name, m.value)
		}
	}
}

func TestMetricsHandlerNoAuth(t *testing.T) {
	// token 为空时不要求鉴权
	h := metricsHandler("")
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("无鉴权配置应返回 200,实际 %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; version=0.0.4; charset=utf-8" {
		t.Errorf("Content-Type = %q, want Prometheus 文本类型", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "# HELP jiankong_build_info") {
		t.Error("响应体缺少 jiankong_build_info")
	}
	if !strings.Contains(body, "jiankong_scrape_duration_seconds ") {
		t.Error("响应体缺少 jiankong_scrape_duration_seconds")
	}
}

func TestMetricsHandlerWrongToken(t *testing.T) {
	h := metricsHandler("tok")
	// 错误的 token 应返回 401
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("错误 token 应返回 401,实际 %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Errorf("401 响应体应含 unauthorized,实际 %q", rec.Body.String())
	}
	// 带正确 token 但多余空格也不应通过
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer  tok")
	rec = httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("格式错误的 Authorization 应返回 401,实际 %d", rec.Code)
	}
}
