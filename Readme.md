# jiankong-feishu

把一台 Linux 服务器的状态做成日报，通过飞书自定义机器人以消息卡片形式推送到群里。单文件 Go 程序、无常驻进程，配合 systemd 定时器每天定时发送一次。

推送内容：主机名、内网 IP、CPU / 内存 / 磁盘使用率（可设阈值，超限显示 ⚠️）、网络流量、系统版本与虚拟化类型、已运行时长。

同系列：钉钉版 jiankong-dingtalk、企业微信版 jiankong-wecom。

## 飞书机器人准备

1. 在目标群里：群设置 → 群机器人 → 添加机器人 → 自定义机器人
2. 安全设置建议勾选「签名校验」，保存好 Webhook 地址和签名密钥
3. 把 Webhook 填到 `FEISHU_WEBHOOK`，密钥填到 `FEISHU_SECRET`（没开签名校验可不填）

## 配置（环境变量）

| 变量 | 必填 | 默认 | 说明 |
| --- | --- | --- | --- |
| `FEISHU_WEBHOOK` | 是 | - | 飞书自定义机器人 Webhook 地址 |
| `FEISHU_SECRET` | 否 | - | 机器人签名校验密钥 |
| `CUSTOM_TITLE` | 否 | 服务器状态日报 | 卡片标题 |
| `CPU_THRESHOLD` | 否 | 80 | CPU 使用率告警阈值（%） |
| `MEM_THRESHOLD` | 否 | 80 | 内存使用率告警阈值（%） |
| `DISK_THRESHOLD` | 否 | 80 | 磁盘使用率告警阈值（%） |
| `REPORT_TIME` | 否 | - | 计划推送时刻，仅在日报中展示，实际调度由 systemd 定时器决定 |

## 使用

### 进项目根目录
```
cd jiankong-feishu
```
### 拉依赖
```
go mod tidy
```
### 编译
```
CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=v1.0.0" -o jiankong-feishu .
```
### 本地运行
```
export FEISHU_WEBHOOK=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
export FEISHU_SECRET=xxxxxxxxxxxxxxxx
./jiankong-feishu
```

### systemd 管理
#### 二进制放置
```
sudo cp jiankong-feishu /usr/local/bin/
sudo chmod +x /usr/local/bin/jiankong-feishu
```

#### systemd 单元文件
```
sudo tee /etc/systemd/system/jiankong-feishu.service >/dev/null <<'EOF'
[Unit]
Description=Push server status to Feishu once
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
Environment=FEISHU_WEBHOOK=https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN
Environment=FEISHU_SECRET=YOUR_SECRET
Environment=CPU_THRESHOLD=85
Environment=MEM_THRESHOLD=85
Environment=DISK_THRESHOLD=85
Environment=REPORT_TIME=09:30
ExecStart=/usr/local/bin/jiankong-feishu
EOF
```
#### systemd 定时器
```
sudo tee /etc/systemd/system/jiankong-feishu.timer >/dev/null <<'EOF'
[Unit]
Description=Daily jiankong to Feishu

[Timer]
OnCalendar=*-*-* 09:30:00
Persistent=true

[Install]
WantedBy=timers.target
EOF
```

### 启用并立即测试
```
sudo systemctl daemon-reload
sudo systemctl enable --now jiankong-feishu.timer
sudo systemctl start jiankong-feishu.service
journalctl -u jiankong-feishu --no-pager -e
```

## 更新记录

### v1.0.0 (2026-11-15)
- 首个版本：服务器状态日报以飞书消息卡片推送（CPU / 内存 / 磁盘 / 网络 / 主机信息）
- 支持签名校验；推送失败（飞书 code 非 0）会正确报错退出，不会误报成功
