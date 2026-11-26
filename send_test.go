package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// fake 飞书服务端:只记录请求,不发真实请求。
func TestSendFeishuCardSuccess(t *testing.T) {
	var gotMethod, gotCT string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotCT = r.Method, r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
	}))
	defer srv.Close()

	if err := sendFeishuCard(srv.URL, "", "测试标题", "测试正文"); err != nil {
		t.Fatalf("sendFeishuCard 返回错误: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("请求方法 = %q, want POST", gotMethod)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
	if gotBody["msg_type"] != "interactive" {
		t.Errorf("msg_type = %v, want interactive", gotBody["msg_type"])
	}
	card, ok := gotBody["card"].(map[string]interface{})
	if !ok {
		t.Fatalf("缺少 card 字段: %v", gotBody)
	}
	header, ok := card["header"].(map[string]interface{})
	if !ok {
		t.Fatalf("缺少 card.header 字段: %v", card)
	}
	title, ok := header["title"].(map[string]interface{})
	if !ok || title["content"] != "测试标题" {
		t.Errorf("卡片标题 = %v, want 测试标题", header["title"])
	}
	if title["tag"] != "plain_text" {
		t.Errorf("标题 tag = %v, want plain_text", title["tag"])
	}
	if header["template"] != "blue" {
		t.Errorf("template = %v, want blue", header["template"])
	}
	elements, ok := card["elements"].([]interface{})
	if !ok || len(elements) != 1 {
		t.Fatalf("elements 应恰有 1 个,实际 %v", card["elements"])
	}
	el, ok := elements[0].(map[string]interface{})
	if !ok || el["tag"] != "markdown" || el["content"] != "测试正文" {
		t.Errorf("elements[0] = %v, want markdown/测试正文", elements[0])
	}
	// 无密钥时不应带 timestamp/sign
	if _, has := gotBody["timestamp"]; has {
		t.Error("无密钥时不应有 timestamp 字段")
	}
	if _, has := gotBody["sign"]; has {
		t.Error("无密钥时不应有 sign 字段")
	}
}

func TestSendFeishuCardSigned(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
	}))
	defer srv.Close()

	if err := sendFeishuCard(srv.URL, "testsecret", "T", "B"); err != nil {
		t.Fatalf("sendFeishuCard 返回错误: %v", err)
	}
	tsStr, ok := gotBody["timestamp"].(string)
	if !ok || tsStr == "" {
		t.Fatalf("有密钥时应带 timestamp 字段,实际 body: %v", gotBody)
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || ts <= 0 {
		t.Fatalf("timestamp 字段非法: %q", tsStr)
	}
	sign, ok := gotBody["sign"].(string)
	if !ok || sign == "" {
		t.Fatalf("有密钥时应带 sign 字段,实际 body: %v", gotBody)
	}
	// 用公开的签名算法重算,验证签名与时间戳匹配
	if want := signFeishu("testsecret", ts); sign != want {
		t.Errorf("sign = %q, want %q (按 timestamp %s 重算)", sign, want, tsStr)
	}
}

func TestSendFeishuCardHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	err := sendFeishuCard(srv.URL, "", "T", "B")
	if err == nil {
		t.Fatal("HTTP 500 应返回错误")
	}
	if !strings.Contains(err.Error(), "feishu http 500") || !strings.Contains(err.Error(), "boom") {
		t.Errorf("错误信息应含状态码与响应体,实际: %v", err)
	}
}

func TestSendFeishuCardCodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":19001,"msg":"sign error"}`))
	}))
	defer srv.Close()

	err := sendFeishuCard(srv.URL, "", "T", "B")
	if err == nil {
		t.Fatal("飞书 code 非 0 应返回错误(HTTP 200 也不代表成功)")
	}
	if !strings.Contains(err.Error(), "feishu code 19001") {
		t.Errorf("错误信息应含 code,实际: %v", err)
	}
}

func TestSendFeishuCardInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	// 当前行为:HTTP 200 且响应体不是 JSON 时按成功处理
	if err := sendFeishuCard(srv.URL, "", "T", "B"); err != nil {
		t.Errorf("非 JSON 的 200 响应应按成功处理,实际错误: %v", err)
	}
}

func TestSendFeishuCardUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // 端口已关闭,模拟 webhook 不可达

	if err := sendFeishuCard(url, "", "T", "B"); err == nil {
		t.Error("webhook 不可达应返回错误")
	}
}
