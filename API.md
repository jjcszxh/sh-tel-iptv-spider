# 上海电信 IPTV 爬虫 - API 接口文档

> 基础地址：`http://<host>:8888`
>
> 📖 返回 [README](README.md)

---

## 接口总览

| # | 接口 | 说明 |
|---|------|------|
| 1 | `GET /api/health` | 健康检查 |
| 2 | `GET /api/schedule` | 定时任务调度 |
| 3 | `GET /api/requests` | 最近请求记录 |
| 4 | `GET /api/run` | 手动触发任务 |
| 5 | `GET /api/network-check` | 网络连通性检查 |
| 6 | `GET /api/version-check` | 版本更新检查 |
| 7 | `GET /api/self-upgrade` | 远程一键升级 |
| 8 | `GET /api/status.html` | 状态监控页面 |
| 9 | `GET /api/m3u8` | M3U 播放列表 |
| 10 | `GET /api/channel/m3u8` | 单频道 M3U8 |
| 11 | `GET /api/epg` | XMLTV 节目单 |
| 12 | `GET /api/log/stream` | SSE 实时日志流 |

---

## 1. 健康检查

```
GET /api/health
```

**返回示例：**

```json
{
  "status": "ok",
  "db": "connected",
  "session": "valid",
  "uptime": "running",
  "last_fetch": "2026-06-13 13:00:00",
  "last_epg_fetch": "2026-06-13 12:58:30",
  "channel_count": 230,
  "epg_count": 4521
}
```

| 字段 | 说明 | 可能值 |
|------|------|--------|
| `status` | 系统整体状态 | `ok` / `degraded` / `down` |
| `db` | 数据库连接 | `connected` / `disconnected` |
| `session` | 认证会话 | `valid` / `expired` / `not_authenticated` / `not_initialized` |
| `uptime` | 运行状态 | `running` |
| `last_fetch` | 频道列表最后拉取时间 | 时间字符串或空 |
| `last_epg_fetch` | 节目单最后拉取时间 | 时间字符串或空 |
| `channel_count` | 频道总数 | 数字 |
| `epg_count` | 近7天节目单总数 | 数字 |

**HTTP 状态码：**

| 状态码 | 对应 `status` |
|--------|---------------|
| 200 | `ok` 或 `degraded` |
| 503 | `down`（数据库断开） |

---

## 2. 定时任务调度

```
GET /api/schedule
```

**返回示例：**

```json
[
  {
    "ID": 1,
    "PreTime": "2026-06-13T08:00:00Z",
    "NextTime": "2026-06-13T16:00:00Z"
  }
]
```

| 字段 | 说明 |
|------|------|
| `ID` | 任务编号 |
| `PreTime` | 上次执行时间 (ISO 8601) |
| `NextTime` | 下次执行时间 (ISO 8601) |

---

## 3. 最近请求记录

```
GET /api/requests
```

**返回示例：**

```json
[
  {
    "ip": "192.168.0.100:54321",
    "path": "/api/m3u8",
    "time": "2026-06-13T13:30:00+08:00",
    "ua": "TiviMate/4.7.0"
  }
]
```

| 字段 | 说明 |
|------|------|
| `ip` | 请求来源 IP |
| `path` | 请求路径 |
| `time` | 请求时间 |
| `ua` | User-Agent（截取前100字符） |

> 内存环形缓冲区，保留最近 **100** 条记录，重启后清空。

---

## 4. 手动触发任务

```
GET /api/run?task=<任务名>
```

**可用任务：**

| task 值 | 说明 |
|---------|------|
| `update-chi` | 更新频道列表（拉取上海电信IPTV频道数据） |
| `update-epg` | 更新节目单（逐频道拉取EPG数据） |
| `clean-ch` | 清理频道数据 |
| `clean-chi` | 清理频道信息数据 |
| `clean-epg` | 清理节目单数据 |
| `clean` | 清理全部数据（频道 + 频道信息 + 节目单） |
| `upload-m3u` | 生成并上传 M3U 到 OSS |
| `upload-xmltv` | 生成并上传 XMLTV 到 OSS |
| `upload-xmltv7` | 生成并上传7天 XMLTV 到 OSS |

> 任务异步执行，立即返回 `OK`。并发请求会被合并，避免重复执行。

---

## 5. 网络连通性检查

```
GET /api/network-check
```

**返回示例：**

```json
{
  "internet": "ok",
  "internet_ms": 35,
  "iptv": "ok",
  "iptv_ms": 12,
  "message": "外网正常(35ms)，IPTV专网正常(12ms)"
}
```

| 字段 | 说明 | 可能值 |
|------|------|--------|
| `internet` | 外网连通状态 | `ok` / `fail` |
| `internet_ms` | 外网延迟（毫秒） | 数字 |
| `iptv` | IPTV 专网连通状态 | `ok` / `fail` |
| `iptv_ms` | IPTV 专网延迟（毫秒） | 数字 |
| `message` | 汇总描述信息 | 字符串 |

> 外网检测 `www.baidu.com:80`，IPTV 专网检测 `config.yaml` 中 `stb.auth_host` 配置的地址。

---

## 6. 版本更新检查

```
GET /api/version-check
```

**返回示例：**

```json
{
  "current": "V0.0.8",
  "latest": "V0.0.9",
  "has_update": true,
  "url": "https://github.com/jjcszxh/sh-tel-iptv-spider/releases"
}
```

| 字段 | 说明 |
|------|------|
| `current` | 当前程序版本 |
| `latest` | 远端最新版本（从 GitHub 仓库 `version.txt` 读取） |
| `has_update` | 是否有新版本可用 |
| `url` | GitHub Releases 页面地址 |

> 结果缓存 **5 分钟**，避免频繁请求远端。当 `has_update` 为 `true` 时，Web 页面底部会闪烁提示并提供一键升级按钮。

---

## 7. 远程一键升级

```
GET /api/self-upgrade
```

**功能**：从 GitHub Release 下载最新版本二进制文件，校验后替换当前程序并重启。

**返回示例（成功）：**

```json
{
  "success": true,
  "message": "升级成功 V0.0.8 -> V0.0.9，程序将在 1 秒后重启",
  "step": "done",
  "version": "V0.0.9"
}
```

**返回示例（失败 - 已是最新）：**

```json
{
  "success": false,
  "message": "已经是最新版本 V0.0.9",
  "step": "check"
}
```

**返回示例（失败 - 下载404）：**

```json
{
  "success": false,
  "message": "下载失败: HTTP 404",
  "step": "download",
  "version": "V0.0.9"
}
```

| 字段 | 说明 |
|------|------|
| `success` | 是否升级成功 |
| `message` | 详细说明信息 |
| `step` | 当前步骤（`check` / `download` / `verify` / `save` / `backup` / `replace` / `done`） |
| `version` | 目标版本号 |

**升级流程：**

1. 调用 `/api/version-check` 判断是否有新版本
2. 从 GitHub Release 下载对应平台的二进制文件（3次重试，超时5分钟）
3. 校验文件头（Linux: ELF 头 `0x7F E L F`，Windows: PE 头 `MZ`）
4. 备份当前程序 → 替换为新文件
5. 程序退出，由进程管理器（如 procd）自动重启

> ⚠️ 升级过程会退出当前进程，依赖外部进程管理器自动重启。建议配合 procd（OpenWrt）或 systemd 使用。下载限 50MB，文件不完整或校验失败会自动回滚。

---

## 8. 状态监控页面

```
GET /api/status.html
```

返回 HTML 页面，包含：

- 系统健康状态卡片
- 频道列表表格（含搜索、排序、HD/4K 标签）
- 自定义频道高亮显示
- 点击频道名弹出 M3U8 预览弹窗（支持复制）
- 一键下载 M3U8 / EPG
- 网络连通性检查（含实时终端输出）
- 手动触发更新（含实时终端输出）
- SSE 实时日志流
- 最近请求记录
- 暗黑模式切换
- 版本更新提示

---

## 9. M3U 播放列表

```
GET /api/m3u8?<参数>
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `udpxy` | 否 | udpxy 代理地址，如 `http://192.168.0.1:4022` |
| `scheme` | 否 | URL 前缀，如 `rtsp://`、`rtp://`、`igmp://` |
| `xteve` | 否 | 设为 `true` 输出 xteve 兼容格式 |
| `all` | 否 | 设为 `true` 包含所有频道（含不可用） |
| `ref` | 否 | 设为 `true` 跳过缓存强制刷新 |

> `udpxy`、`scheme`、`xteve` 三选一，都不传则使用默认组播地址。

**示例：**

```
/api/m3u8?udpxy=http://192.168.0.1:4022
/api/m3u8?xteve=true
/api/m3u8?scheme=igmp://
```

返回 `iptv.m3u` 文件，浏览器直接下载。

---

## 10. 单频道 M3U8

```
GET /api/channel/m3u8?name=<频道名>
```

**参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 频道名称，精确匹配（如 `CCTV-1`） |

**返回示例：**

```
#EXTM3U url-tvg="..."
#EXTINF:-1 tvg-id="1" tvg-name="CCTV-1" group-title="央视",CCTV-1
igmp://233.18.204.1:5140
```

| HTTP 状态码 | 含义 |
|-------------|------|
| 200 | 正常，返回纯文本 |
| 400 | 缺少 `name` 参数 |
| 404 | 未找到该频道 |

> 数据从 M3U8 缓存中提取，无额外数据库查询。

---

## 11. XMLTV 节目单

```
GET /api/epg?<参数>
```

**参数：**

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `daysAgo` | 否 | `1` | 拉取几天前的数据（1 表示1天前至今） |
| `ref` | 否 | - | 设为 `true` 跳过缓存强制刷新 |

**示例：**

```
/api/epg?daysAgo=3
/api/epg?daysAgo=7&ref=true
```

返回 `application/xml` 格式的 XMLTV 数据。

---

## 12. SSE 实时日志流

```
GET /api/log/stream
```

Server-Sent Events 实时推送日志，连接后立即推送最近 50 行，之后每 300ms 增量推送新日志行。

**事件格式：**

```
data: 2026-06-13 14:30:00 INFO  频道列表拉取完成

data: 2026-06-13 14:30:05 INFO  EPG 更新完成
```

**前端使用示例：**

```javascript
const evt = new EventSource("/api/log/stream");
evt.onmessage = (e) => console.log(e.data);
evt.onerror = () => evt.close();
```

---

## 状态码说明

| HTTP 状态码 | 含义 |
|-------------|------|
| 200 | 正常 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 503 | 系统不可用（数据库断开） |
