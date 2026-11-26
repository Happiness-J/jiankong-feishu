package main

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("FEISHU_WEBHOOK", "https://example.invalid/hook")
	cfg := loadConfig()
	if cfg.FeishuWebhook != "https://example.invalid/hook" {
		t.Errorf("FeishuWebhook = %q", cfg.FeishuWebhook)
	}
	if cfg.FeishuSecret != "" {
		t.Errorf("FeishuSecret 默认应为空,实际 %q", cfg.FeishuSecret)
	}
	if cfg.ReportTime != "-" {
		t.Errorf("ReportTime 默认 = %q, want -", cfg.ReportTime)
	}
	if cfg.CPUAlert != 80 || cfg.MemAlert != 80 || cfg.DiskAlert != 80 {
		t.Errorf("阈值默认应为 80/80/80,实际 %d/%d/%d", cfg.CPUAlert, cfg.MemAlert, cfg.DiskAlert)
	}
	if cfg.CustomTitle != "服务器状态日报" {
		t.Errorf("CustomTitle 默认 = %q, want 服务器状态日报", cfg.CustomTitle)
	}
}

func TestLoadConfigCustom(t *testing.T) {
	t.Setenv("FEISHU_WEBHOOK", "https://example.invalid/hook2")
	t.Setenv("FEISHU_SECRET", "s3cret")
	t.Setenv("REPORT_TIME", "08:30")
	t.Setenv("CPU_THRESHOLD", "90")
	t.Setenv("MEM_THRESHOLD", "70")
	t.Setenv("DISK_THRESHOLD", "60")
	t.Setenv("CUSTOM_TITLE", "自定义标题")
	cfg := loadConfig()
	if cfg.FeishuWebhook != "https://example.invalid/hook2" {
		t.Errorf("FeishuWebhook = %q", cfg.FeishuWebhook)
	}
	if cfg.FeishuSecret != "s3cret" {
		t.Errorf("FeishuSecret = %q", cfg.FeishuSecret)
	}
	if cfg.ReportTime != "08:30" {
		t.Errorf("ReportTime = %q", cfg.ReportTime)
	}
	if cfg.CPUAlert != 90 || cfg.MemAlert != 70 || cfg.DiskAlert != 60 {
		t.Errorf("阈值 = %d/%d/%d, want 90/70/60", cfg.CPUAlert, cfg.MemAlert, cfg.DiskAlert)
	}
	if cfg.CustomTitle != "自定义标题" {
		t.Errorf("CustomTitle = %q", cfg.CustomTitle)
	}
}
