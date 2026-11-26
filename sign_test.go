package main

import "testing"

// 已知向量:HMAC-SHA256(key="秒级时间戳\nsecret", msg="")后 base64,
// 独立算出,与实现交叉验证(而不是只测两次调用结果相同)。
func TestSignFeishuKnownVector(t *testing.T) {
	cases := []struct {
		secret string
		ts     int64
		want   string
	}{
		{"SECtest", 1700000000, "G7XpBpG8NgG02fJOAhX6FRAObIljmFoxVReo8I62pEk="},
		{"", 0, "53Dh/MqIJzmbXR1ky1eoAPGAmY2mY/DJucsfzM60KVk="},
		{"abc", 1, "4lhRCmKS2AbLyOZEpn2glNhQwnUTSFYbNrSRQtG1/08="},
	}
	for _, c := range cases {
		if got := signFeishu(c.secret, c.ts); got != c.want {
			t.Errorf("signFeishu(%q, %d) = %q, want %q", c.secret, c.ts, got, c.want)
		}
	}
	// 密钥不同结果也不同
	if a, b := signFeishu("SECtest", 1700000000), signFeishu("SECother", 1700000000); a == b {
		t.Error("signFeishu should change with secret")
	}
}
