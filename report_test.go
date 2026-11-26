package main

import (
	stdnet "net"
	"strings"
	"testing"
)

func TestBuildReportStructure(t *testing.T) {
	cfg := &Config{
		CustomTitle: "测试标题",
		CPUAlert:    100,
		MemAlert:    100,
		DiskAlert:   100,
		ReportTime:  "-",
	}
	report := buildReport(cfg)
	// 飞书版日报正文不含标题行(标题放在卡片 header 里)
	if strings.Contains(report, "测试标题") {
		t.Errorf("飞书版日报正文不应包含标题,实际:\n%s", report)
	}
	if strings.Contains(report, "##") {
		t.Errorf("飞书版日报正文不应含 markdown 标题,实际:\n%s", report)
	}
	for _, want := range []string{"主机名", "内网IP", "推送时间", "CPU", "内存", "磁盘", "网络", "系统", "运行"} {
		if !strings.Contains(report, want) {
			t.Errorf("日报缺少 %q 字段:\n%s", want, report)
		}
	}
	if strings.Contains(report, "计划时刻") {
		t.Errorf("ReportTime 为 - 时不应出现计划时刻:\n%s", report)
	}
	// 阈值 100 时 CPU 使用率不可能超过阈值,应为 ✅
	if !strings.Contains(report, "✅ **CPU**") {
		t.Errorf("阈值 100 时 CPU 图标应为 ✅:\n%s", report)
	}
}

func TestBuildReportWithReportTime(t *testing.T) {
	cfg := &Config{
		CustomTitle: "测试标题",
		CPUAlert:    80,
		MemAlert:    80,
		DiskAlert:   80,
		ReportTime:  "08:30",
	}
	report := buildReport(cfg)
	if !strings.Contains(report, "计划时刻") || !strings.Contains(report, "08:30") {
		t.Errorf("ReportTime 为 08:30 时应出现计划时刻:\n%s", report)
	}
}

func TestFirstPrivateIPv4(t *testing.T) {
	ip := firstPrivateIPv4()
	parsed := stdnet.ParseIP(ip)
	if parsed == nil || parsed.To4() == nil {
		t.Errorf("firstPrivateIPv4() = %q,不是合法 IPv4 地址", ip)
	}
}

func TestOSVersion(t *testing.T) {
	if got := osVersion(); got == "" {
		t.Error("osVersion() 返回空字符串")
	}
}

func TestDetectVirtualization(t *testing.T) {
	v := detectVirtualization()
	if v == "" {
		t.Error("detectVirtualization() 返回空字符串")
	}
	if strings.Contains(v, "\n") || strings.Contains(v, " ") {
		t.Errorf("detectVirtualization() = %q,应为单个词", v)
	}
}
