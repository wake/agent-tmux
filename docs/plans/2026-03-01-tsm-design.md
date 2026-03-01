# TSM (tmux-session-menu) 設計文件

> 日期：2026-03-01
> 狀態：已確認

## 專案定位

輕量但資訊豐富的 tmux session 選單工具，可在 tmux 內外以統一方式呼叫，專注於快速切換和追蹤 AI 活動。從頭打造，不基於既有 cc-menu。

## 技術棧

| 元件 | 選擇 | 理由 |
|------|------|------|
| 語言 | Go 1.24+ | 單一二進位、跨平台、高效能 |
| TUI | Bubble Tea + Lipgloss + Bubbles | Go 生態最成熟的 TUI 框架，agent-deck 驗證過 |
| 設定檔 | TOML | 清晰可讀 |
| 狀態儲存 | SQLite (WAL mode) | 跨程序安全、輕量 |
| AI 摘要 | 迷你伺服器 + `claude stream -p` + MCP server | 見 AI 摘要架構章節 |

## 專案結構

```
tmux-session-menu/
├── cmd/tsm/                # CLI 入口
│   └── main.go
├── internal/
│   ├── ui/                 # Bubble Tea TUI 層
│   │   ├── app.go          # 主 Model (Elm Architecture)
│   │   ├── sessionlist.go  # Session 列表 + 群組元件
│   │   ├── preview.go      # 底部預覽區塊
│   │   ├── dialog.go       # 操作對話框（更名/建立/刪除確認）
│   │   └── styles.go       # 主題/樣式
│   ├── tmux/               # Tmux 互動層
│   │   ├── session.go      # Session CRUD 操作
│   │   ├── detector.go     # 三層狀態偵測
│   │   └── pipe.go         # Control mode 管道（零子程序）
│   ├── ai/                 # AI 整合
│   │   ├── summary.go      # 摘要服務客戶端
│   │   └── status.go       # AI 模型狀態追蹤
│   ├── config/             # TOML 設定管理
│   └── store/              # SQLite 持久化
├── config.toml             # 預設設定範例
└── docs/                   # 設計文件
```

## UI 設計

### 佈局：單欄 + 群組 + 底部預覽

```
tmux session menu  (↑↓/jk 選擇, Enter 確認, q 離開)

▼ 工作專案
   ► my-project    ● 進行中  3m   claude-sonnet-4-6
     api-server    ◐ 等待中 15m   claude-sonnet-4-6
     frontend      ○ 閒置    2h

▼ 維運
     devops        ○ 閒置    1d
     monitoring    ○ 閒置    3h

  [n] 新建  [g] 新群組  [q] 離開

────────────────────────────────────────────────
Preview: 正在重構 auth 模組的 middleware，已完成 JWT
驗證邏輯，接下來要處理 refresh token 的部分...
```

### Session 列表項目資訊

```
[游標] session名稱  [狀態圖示] [狀態文字]  [距今時間]
                    [AI模型(如果是AI session)]  [摘要(如果有)]
```

**狀態圖示：**
- `●` 進行中 (綠) — AI 正在工作
- `◐` 等待中 (黃) — 需要人類輸入
- `○` 閒置 (灰) — 無活動
- `✗` 錯誤 (紅) — 異常狀態

### 群組功能

- 群組可折疊/展開（`Tab` 或在群組項上按 `Enter`）
- 可將 session 移入/移出群組（`m` 鍵）
- 新建群組（`g` 鍵）
- 未分組的 session 顯示在最上方
- **排序規則**：群組和群組內的項目都可上下移動排序，項目不能跨群組排序（必須先移動到目標群組再排序）
- 群組資訊儲存在 SQLite 中

### 預覽區塊

固定在底部，顯示游標所在 session 的：
- AI 摘要（如果有）
- 或最後幾行終端輸出

### 快捷鍵

| 鍵 | 動作 |
|---|---|
| `j/k` 或 `↑/↓` | 上下選擇 |
| `Enter` | 連線（attach）或展開群組 |
| `r` | 更名 session |
| `d` | 刪除 session（需確認） |
| `n` | 新建 session |
| `g` | 新建群組 |
| `m` | 移動 session 到群組 |
| `J/K` 或 `Shift+↑/↓` | 上下移動排序 |
| `Tab` | 折疊/展開群組 |
| `/` | 搜尋 session |
| `q` 或 `Esc` | 離開選單 |

### 操作對話框

選中 session 後可進行：
- **連線** — attach 到該 session
- **更名** — 輸入新名稱，AI 建議名稱預填
- **刪除** — kill session（確認對話框）

## 呼叫機制

統一二進位 `tsm`，根據 `$TMUX` 環境變數自動判斷行為：

```bash
# 在 tmux 外 — 直接全螢幕啟動
$ tsm

# 在 tmux 內 — 自動透過 tmux popup 彈出
# tmux.conf 綁定:
# bind C-s display-popup -E -w 80% -h 80% "tsm"

# 強制模式
$ tsm --popup    # 強制用 popup
$ tsm --inline   # 強制用全螢幕
```

## 三層狀態偵測

### 第一層：Claude Code Hooks（最快，亞秒級）

注入 hooks 到 Claude Code 的 `settings.json`，監聽事件：
- `SessionStart` — session 啟動
- `UserPromptSubmit` — 使用者送出 prompt
- `Stop` — Claude 停止回應
- `PermissionRequest` — 需要授權確認
- `Notification` — 通知
- `SessionEnd` — session 結束

Hook 觸發時寫入狀態檔案到 `~/.config/tsm/status/<session>`，TUI 透過 FileWatcher 監聽。

### 第二層：Pane Title 偵測（快，不需讀取內容）

透過 `tmux list-panes -a -F "#{pane_title}"` 一次取得所有 pane 標題：
- Braille spinner 字元 (U+2800-28FF) → running
- 完成標記 (checkmark) → done，進入第三層

### 第三層：Terminal Content 解析（最詳細，fallback）

透過 control mode pipe 或 `tmux capture-pane` 抓取終端輸出：

**忙碌指標（AI 正在工作）：**
- `ctrl+c to interrupt` / `esc to interrupt`
- Braille / asterisk spinner 字元
- Claude 思考中關鍵字（Clauding... 等 90+ 個 whimsical words）

**等待指標（需要人類輸入）：**
- 權限提示：`Yes, allow once`、`No, and tell Claude...`
- 輸入提示符：`>` 或 `❯` 出現在最後一行
- 問題提示：`Continue?`、`(Y/n)`

### 偵測優先順序

```
Hook 狀態（2分鐘內有效）→ 直接信任
    ↓ 否則
Pane title 有 spinner → running
    ↓ 否則
Title 有完成標記 → 進入內容解析
    ↓ 否則
Capture-pane → 檢查忙碌指標 → 檢查等待指標
```

### 效能優化

- **Control mode 管道**：開持久的 `tmux -C` 連線，所有指令走 stdin/stdout，零子程序開銷
- **輪替更新**：每個 tick 只檢查部分 session，不全部掃描
- **快取**：session 存在性每 tick 只查一次

## AI 摘要架構

```
tsm (TUI)
 │
 ├── 啟動/連接摘要伺服器 (背景常駐)
 │    └── 使用 claude stream -p 搭配 MCP server
 │         └── prompt 一次提供多組 session 資料
 │              → AI 判斷後透過 MCP tool call 輸出結果
 │
 └── 從摘要伺服器取得結果顯示在 UI
```

1. TUI 啟動時，收集所有 session 的最後 N 行終端輸出
2. 打包成一個 prompt 送給摘要伺服器
3. 伺服器透過 `claude stream -p` + MCP server 處理
4. AI 透過 MCP tool call 回傳各 session 的摘要和建議命名
5. TUI 非同步接收並更新顯示

## 資料儲存

```
~/.config/tsm/
├── config.toml          # 使用者設定
├── state.db             # SQLite (sessions, groups, 排序, 自訂名稱)
├── status/              # Hook 狀態檔案
├── cache/               # AI 摘要快取
└── logs/                # 除錯日誌
```

## 開發階段

### Phase 1 — MVP

1. Session 列表 + 群組 + 排序
2. 連線 / 更名 / 刪除 session
3. 新建 session
4. 底部預覽（terminal capture）
5. 三層狀態偵測
6. AI 模型名稱顯示
7. tmux 內外統一呼叫
8. TOML 設定檔

### Phase 2 — AI 整合

1. AI 摘要伺服器 + MCP 整合
2. AI 建議命名
3. 搜尋功能
4. 更多 AI 工具支援（Gemini CLI 等）

## 參考資料

- [agent-deck](https://github.com/asheshgoplani/agent-deck) — 功能完整的 Go TUI session manager，重要架構參考
- [tmux-menus](https://github.com/jaclu/tmux-menus) — 簡單參考
- cc-menu 截圖和描述 — 使用體驗參考
