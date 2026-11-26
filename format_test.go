package main

import (
	"strings"
	"testing"
)

func TestHumanBoundaries(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1025, "1.0 KB"},
		{1048575, "1024.0 KB"}, // 1 MiB - 1 B,进位到 1024.0 KB
		{10 * 1024 * 1024, "10.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{1024 * 1024 * 1024 * 1024 * 1024, "1.0 PB"},
		{1536, "1.5 KB"},
		{5*1024*1024*1024 + 512*1024*1024, "5.5 GB"},
	}
	for _, c := range cases {
		if got := human(c.in); got != c.want {
			t.Errorf("human(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIconBoundaries(t *testing.T) {
	cases := []struct {
		v         float64
		threshold int
		want      string
	}{
		{0, 0, "✅"},   // 等于阈值不告警
		{-5, 80, "✅"}, // 负值不告警
		{80, 80, "✅"}, // 等于阈值不告警
		{79.9999, 80, "✅"},
		{80.0001, 80, "⚠️"}, // 刚过阈值即告警
		{100, 100, "✅"},
		{100.1, 100, "⚠️"},
		{0, 100, "✅"},
	}
	for _, c := range cases {
		if got := icon(c.v, c.threshold); got != c.want {
			t.Errorf("icon(%v, %d) = %q, want %q", c.v, c.threshold, got, c.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{66, "66"},
		{12.5, "12.5"},
		{12.5000, "12.5"},
		{3.14159, "3.1416"}, // 保留 4 位小数
		{100, "100"},
		{-2.5, "-2.5"},
		{0.0001, "0.0001"},
		{123456.789, "123456.789"},
		{0.1, "0.1"},
		{99.99999, "100"}, // 四舍五入后去零
	}
	for _, c := range cases {
		if got := formatFloat(c.in); got != c.want {
			t.Errorf("formatFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLabelString(t *testing.T) {
	if got := labelString(nil); got != "" {
		t.Errorf("labelString(nil) = %q, want %q", got, "")
	}
	if got := labelString(map[string]string{}); got != "" {
		t.Errorf("labelString(empty) = %q, want %q", got, "")
	}
	if got := labelString(map[string]string{"device": "eth0"}); got != `{device="eth0"}` {
		t.Errorf("labelString single = %q, want %q", got, `{device="eth0"}`)
	}
	// 多标签按键名排序,与 map 写入顺序无关
	if got := labelString(map[string]string{"b": "2", "a": "1"}); got != `{a="1",b="2"}` {
		t.Errorf("labelString multi = %q, want %q", got, `{a="1",b="2"}`)
	}
	if got := labelString(map[string]string{"mountpoint": "/"}); got != `{mountpoint="/"}` {
		t.Errorf("labelString mountpoint = %q, want %q", got, `{mountpoint="/"}`)
	}
}

// 注意:labelString 先用 Replacer 转义、再用 %q 格式化,引号/反斜杠/换行会被
// 转义两次(例如值里的 " 会变成 \\\"),这是当前行为的快照;若以后修成单次
// 转义,本测试需要同步更新期望值。
func TestLabelStringEscaping(t *testing.T) {
	cases := []struct {
		in   map[string]string
		want string
	}{
		{map[string]string{"k": `a"b`}, `{k="a\\\"b"}`},
		{map[string]string{"k": `a\b`}, `{k="a\\\\b"}`},
		{map[string]string{"k": "a\nb"}, `{k="a\\nb"}`},
	}
	for _, c := range cases {
		if got := labelString(c.in); got != c.want {
			t.Errorf("labelString(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRenderPrometheusEdge(t *testing.T) {
	if got := renderPrometheus(nil); got != "" {
		t.Errorf("renderPrometheus(nil) = %q, want empty", got)
	}
	if got := renderPrometheus([]metric{}); got != "" {
		t.Errorf("renderPrometheus(empty) = %q, want empty", got)
	}

	// 样本按指标名排序,与输入顺序无关
	out := renderPrometheus([]metric{
		{name: "z_metric", help: "Z", typ: "gauge", value: 1},
		{name: "a_metric", help: "A", typ: "gauge", value: 2},
	})
	if strings.Index(out, "a_metric") > strings.Index(out, "z_metric") {
		t.Errorf("指标未按名字排序:\n%s", out)
	}
	if !strings.Contains(out, "# HELP a_metric A\n# TYPE a_metric gauge\na_metric 2\n") {
		t.Errorf("单样本输出格式不对:\n%s", out)
	}

	// 同名指标只输出一次 HELP/TYPE,采用排序后第一个样本的 help
	out = renderPrometheus([]metric{
		{name: "dup", help: "HELP-B", typ: "gauge", labels: map[string]string{"mountpoint": "/b"}, value: 2},
		{name: "dup", help: "HELP-A", typ: "gauge", labels: map[string]string{"mountpoint": "/a"}, value: 1},
	})
	if strings.Count(out, "# HELP dup ") != 1 || strings.Count(out, "# TYPE dup ") != 1 {
		t.Errorf("同名指标 HELP/TYPE 应只出现一次:\n%s", out)
	}
	if !strings.Contains(out, "# HELP dup HELP-A\n") {
		t.Errorf("应采用排序后第一个样本的 help:\n%s", out)
	}
	if strings.Contains(out, "HELP-B") {
		t.Errorf("第二个样本的 help 不应出现:\n%s", out)
	}
	if !strings.Contains(out, `dup{mountpoint="/a"} 1`) || !strings.Contains(out, `dup{mountpoint="/b"} 2`) {
		t.Errorf("同名多标签样本缺失:\n%s", out)
	}

	// counter 类型原样输出
	out = renderPrometheus([]metric{
		{name: "c_total", help: "计数", typ: "counter", labels: map[string]string{"device": "eth0"}, value: 12345},
	})
	if !strings.Contains(out, "# TYPE c_total counter\n") || !strings.Contains(out, `c_total{device="eth0"} 12345`) {
		t.Errorf("counter 输出不对:\n%s", out)
	}
}

func TestRenderPrometheusDeterministic(t *testing.T) {
	mk := func() []metric {
		return []metric{
			{name: "m", help: "H", typ: "gauge", labels: map[string]string{"b": "2", "a": "1"}, value: 1.5},
			{name: "m", help: "H", typ: "gauge", labels: map[string]string{"a": "1"}, value: 2},
		}
	}
	a, b := renderPrometheus(mk()), renderPrometheus(mk())
	if a != b {
		t.Errorf("两次渲染结果不一致:\n%s\n---\n%s", a, b)
	}
}
