package main

import "testing"

func TestHuman(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{8 * 1024 * 1024 * 1024, "8.0 GB"},
	}
	for _, c := range cases {
		if got := human(c.in); got != c.want {
			t.Errorf("human(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIcon(t *testing.T) {
	if icon(50, 80) != "✅" {
		t.Error("icon below threshold should be ✅")
	}
	if icon(80, 80) != "✅" {
		t.Error("icon at threshold should be ✅")
	}
	if icon(81, 80) != "⚠️" {
		t.Error("icon above threshold should be ⚠️")
	}
}

func TestSignFeishu(t *testing.T) {
	a := signFeishu("SECtest", 1700000000)
	b := signFeishu("SECtest", 1700000000)
	if a == "" {
		t.Fatal("signFeishu returned empty")
	}
	if a != b {
		t.Error("signFeishu should be deterministic for same input")
	}
	if c := signFeishu("SECtest", 1700000001); c == a {
		t.Error("signFeishu should change with timestamp")
	}
	if c := signFeishu("SECother", 1700000000); c == a {
		t.Error("signFeishu should change with secret")
	}
}
