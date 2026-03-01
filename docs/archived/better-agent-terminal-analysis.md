# Better Agent Terminal 功能分析報告

> 來源：https://github.com/tony1223/better-agent-terminal
> 分析日期：2026-02-28

## 專案概述

Better Agent Terminal 是一個基於 **Electron + React 18 + TypeScript** 的跨平台桌面應用程式（Windows / macOS / Linux），定位為**終端聚合器 + AI 代理整合平台**。核心賣點是將傳統終端、檔案瀏覽、Git 操作、Claude Code AI 助手無縫整合在同一個視窗中。

| 項目 | 資訊 |
|------|------|
| 版本 | 1.47.0 |
| 授權 | MIT |
| 技術棧 | Electron 28 / React 18 / TypeScript / xterm.js 5.5 / node-pty |
| AI SDK | @anthropic-ai/claude-agent-sdk ^0.2.47 / @anthropic-ai/claude-code ^2.1.49 |
| 資料庫 | better-sqlite3（片段管理） |
| 建置工具 | Vite 5 + electron-builder |

---

## 功能清單

### 1. 工作區管理

以「工作區」為單位組織專案，每個工作區對應一個資料夾路徑，內含獨立的終端和 Claude 會話。

| 功能 | 說明 |
|------|------|
| 多工作區 | 按專案資料夾分組，獨立配置 |
| 拖放排序 | 自由重排工作區順序 |
| 工作區分組 | 下拉篩選器分類工作區 |
| 可分離視窗 | 將工作區彈出為獨立視窗，重啟後自動重新連接 |
| 環境變數 | 全域和工作區層級的環境變數，可啟用/停用 |
| 活動指示器 | 視覺化顯示哪些工作區有終端活動 |
| 雙擊重新命名 | 快速編輯工作區別名 |
| 右鍵選單 | 新增終端、環境變數、分離等快速操作 |

### 2. 終端功能

採用 Google Meet 風格佈局：主面板 70% + 縮圖欄 30%。

**佈局結構：**
```
┌──────────────────────────┬────────────┐
│                          │ 縮圖 1     │
│     主面板（70%）         │ 縮圖 2     │
│   當前聚焦的終端或 Claude  │ 縮圖 3     │
│                          │ [+ 新增]   │
└──────────────────────────┴────────────┘
```

| 功能 | 說明 |
|------|------|
| 多終端/工作區 | 每個工作區可開多個 xterm.js 終端 |
| Unicode/CJK 支援 | 完整中日韓文字支援 |
| 歷史保留 | 終端重啟時保留舊輸出（淡色顯示，可摺疊） |
| 標籤導航 | Terminal / Files / Git 三個檢視切換 |
| 縮圖預覽 | 實時顯示最後 8 行輸出（自動清除 ANSI） |
| Shell 自動偵測 | Windows: pwsh > powershell > cmd；macOS/Linux: $SHELL 或 zsh |
| 重啟功能 | 記住 cwd，保留歷史，顯示分隔線 |

### 3. Claude Code 代理整合

透過 `@anthropic-ai/claude-agent-sdk` 直接整合 Claude Code，無需開獨立終端。

**核心能力：**

| 功能 | 說明 |
|------|------|
| 訊息串流 | 實時流式傳輸 Claude 回應 |
| 延伸思考 | 可摺疊的思考過程區塊 |
| 權限管理 | bypass / allModerations / toolModerations / denial / sandbox 五種模式 |
| 活動工作列 | 顯示執行中操作及經過時間 |
| 會話恢復 | 跨重啟持久化，自動恢復對話 |
| 會話暫停/喚醒 | Rest / Wake 功能暫時釋放資源 |
| 提示歷史 | 檢視並複製所有使用者提示 |
| 圖像附件 | 拖放或按鈕上傳（最多 5 張），base64 編碼 |
| 可點擊 URL | Markdown 連結和裸 URL 在瀏覽器中開啟 |
| 可點擊檔案路徑 | 檔案路徑預覽，帶語法高亮 |
| Ctrl+P 檔案選擇器 | 搜尋並附加檔案到上下文 |
| 多模型支援 | 可切換 Claude 模型 |
| 努力等級 | low / medium / high / max 四段 |
| 1M Token 上下文 | 為 Sonnet 4/4.5 啟用百萬 token 上下文窗口 |

**狀態列資訊：**
- Token 使用量與成本估計
- 上下文窗口使用百分比
- 當前模型名稱
- Git 分支
- 會話持續時間
- 輪次計數

**斜線命令：**
- `/resume` — 恢復先前的 Claude 會話
- `/model` — 切換可用模型

### 4. 多代理預設

不限於 Claude Code，支援多種 CLI 代理：

| 代理 | 圖示 | 啟動命令 |
|------|------|----------|
| Claude Code | ✦（橙） | `claude --continue` |
| Gemini CLI | ◇（藍） | `gemini` |
| Codex CLI | ⬡（綠） | `codex` |
| GitHub Copilot | ⬢（紫） | `gh copilot` |
| Terminal | ⌘（灰） | 無代理，純終端 |

可在全域或工作區層級設定預設代理。

### 5. 檔案瀏覽器（Files Tab）

| 功能 | 說明 |
|------|------|
| 樹狀結構 | 展開/摺疊資料夾 |
| 檔案預覽 | 100+ 檔案類型的語法高亮（highlight.js） |
| 搜尋 | 遞迴搜尋檔案（自動忽略 .git / node_modules 等） |
| 圖像預覽 | 內聯顯示 PNG / JPG / GIF |
| 檔案大小限制 | 預覽限制 512KB |

### 6. Git 整合（Git Tab）

| 功能 | 說明 |
|------|------|
| 分支檢視 | 顯示當前分支 |
| 提交日誌 | 完整歷史，相對時間顯示 |
| Diff 檢視器 | 按檔案展開差異，語法高亮 +/- 行 |
| 狀態指示 | 未追蹤、已修改、已新增等 |
| GitHub 連結偵測 | 自動解析 SSH/HTTPS remote 並生成 GitHub URL |

### 7. 程式碼片段管理器

基於 SQLite 的片段儲存系統。

| 功能 | 說明 |
|------|------|
| CRUD | 建立、讀取、更新、刪除片段 |
| 分類與標籤 | 組織片段 |
| 最愛標記 | 標星常用片段 |
| 搜尋 | 按名稱、內容或標籤搜尋 |
| 格式支援 | plaintext / markdown |
| 雙擊操作 | 複製到剪貼簿 / 貼到終端（可選自動 Enter）/ 編輯 |

### 8. 個人檔案管理（Profiles）

| 功能 | 說明 |
|------|------|
| 多配置檔 | 儲存和載入不同的工作區配置 |
| 快速切換 | 從選單切換配置 |
| 複製 | 從現有配置複製新的 |
| 啟動參數 | `--profile=<id>` 指定啟動配置 |

每個配置包含：所有工作區狀態、終端狀態、活動工作區、活動分組。

### 9. 遠端連接（Remote Multi-Client）

伺服器/客戶端架構，用於分散式終端管理。

**伺服器端：**
- WebSocket 伺服器（預設埠 9876）
- Token 認證
- 心跳機制（30 秒 ping）
- 已連接客戶端清單

**客戶端：**
- 連接到遠端主機
- 代理所有 PTY / Claude / Git / 檔案系統操作
- 雙向事件廣播
- 自動重連

### 10. 訊息存檔

| 功能 | 說明 |
|------|------|
| 自動存檔 | Claude 訊息超過 200 條後自動存檔至磁碟 |
| 分頁載入 | 按需載入舊訊息（50 條/批） |
| JSONL 格式 | 每行一條訊息 |

### 11. 設定面板

| 類別 | 選項 |
|------|------|
| 外觀 | 14+ 種字體、字體大小、7 種色彩預設、自訂顏色 |
| Shell | Shell 類型選擇、自訂路徑 |
| 環境變數 | 全域環境變數管理 |
| 代理 | 預設代理、自動命令、權限模式 |
| 進階 | 1M token 上下文啟用 |
| 遠端 | 伺服器啟停、埠設定、客戶端監測 |

### 12. 鍵盤快捷鍵

| 快捷鍵 | 功能 |
|--------|------|
| `Ctrl+P` / `Cmd+P` | 檔案選擇器 |
| `Shift+Tab` | 終端/代理模式切換 |
| `Enter` | 發送訊息 |
| `Shift+Enter` | 插入換行 |
| `Escape` | 停止串流/關閉對話框 |
| `Ctrl+Shift+C` | 複製選中文字 |
| `Ctrl+Shift+V` | 貼上 |
| 右鍵 | 複製（已選中時）或貼上 |

---

## 專案結構

```
better-agent-terminal/
├── electron/                          # Electron 主進程
│   ├── main.ts                        # 視窗管理、IPC 路由、選單（27KB）
│   ├── preload.ts                     # IPC 安全橋接（15KB）
│   ├── claude-agent-manager.ts        # Claude SDK 會話管理（37KB）
│   ├── pty-manager.ts                 # PTY 進程管理（8.5KB）
│   ├── profile-manager.ts            # 個人檔案管理（9KB）
│   ├── snippet-db.ts                  # SQLite 片段儲存（4.7KB）
│   ├── update-checker.ts             # 版本檢查（3.1KB）
│   └── remote/                        # 遠端伺服器/客戶端
│       ├── protocol.ts               # 協議定義
│       ├── remote-server.ts           # WebSocket 伺服器（5.8KB）
│       ├── remote-client.ts           # 遠端客戶端（6.3KB）
│       ├── handler-registry.ts        # 處理器註冊
│       └── broadcast-hub.ts           # 事件廣播
├── src/                               # React 前端
│   ├── App.tsx                        # 主應用元件（16.7KB）
│   ├── main.tsx                       # Entry point
│   ├── components/
│   │   ├── ClaudeAgentPanel.tsx       # Claude AI 代理 UI（91KB，最大元件）
│   │   ├── Sidebar.tsx                # 工作區清單（19KB）
│   │   ├── ProfilePanel.tsx           # 個人檔案管理（19.6KB）
│   │   ├── SettingsPanel.tsx          # 設定面板（17.6KB）
│   │   ├── TerminalPanel.tsx          # xterm.js 終端（14.8KB）
│   │   ├── PathLinker.tsx             # 路徑/URL 預覽（14.3KB）
│   │   ├── WorkspaceView.tsx          # 工作區容器（13.8KB）
│   │   ├── FileTree.tsx               # 檔案瀏覽器（13.3KB）
│   │   ├── SnippetPanel.tsx           # 片段管理（13.2KB）
│   │   ├── GitPanel.tsx               # Git 面板（10.5KB）
│   │   ├── EnvVarEditor.tsx           # 環境變數編輯器（5.2KB）
│   │   ├── PromptBox.tsx              # 提示輸入框（5.2KB）
│   │   ├── ThumbnailBar.tsx           # 縮圖列（3.8KB）
│   │   └── ActivityIndicator.tsx      # 活動指示器（1.9KB）
│   ├── stores/
│   │   ├── workspace-store.ts         # 工作區狀態管理（12KB）
│   │   └── settings-store.ts          # 設定狀態管理（6.8KB）
│   ├── types/
│   │   ├── index.ts                   # 核心型別定義
│   │   ├── agent-presets.ts           # 代理預設配置
│   │   ├── claude-agent.ts            # Claude 訊息型別
│   │   └── electron.d.ts             # IPC 型別定義
│   └── styles/
│       ├── base.css
│       └── claude-agent.css           # Claude UI 樣式（33KB）
└── package.json / tsconfig.json / vite.config.ts
```

---

## IPC 通道總覽

應用透過 Electron IPC 在主進程與渲染進程之間通訊。

| 類別 | 通道 | 方向 |
|------|------|------|
| **PTY** | `pty:create` / `write` / `resize` / `kill` / `restart` / `get-cwd` | 雙向 |
| **PTY 事件** | `pty:output` / `pty:exit` | 主 → 渲染 |
| **Claude** | `claude:start-session` / `send-message` / `stop-session` / `set-*` / `resume-session` / `rest-session` / `wake-session` | 渲染 → 主 |
| **Claude 事件** | `claude:message` / `stream` / `tool-use` / `result` / `error` / `status` / `permission-request` / `ask-user` | 主 → 渲染 |
| **工作區** | `workspace:save` / `load` / `detach` / `reattach` | 雙向 |
| **設定** | `settings:save` / `load` / `get-shell-path` | 雙向 |
| **Git** | `git:branch` / `log` / `diff` / `status` / `get-github-url` | 渲染 → 主 |
| **檔案** | `fs:readdir` / `readFile` / `search` | 渲染 → 主 |
| **片段** | `snippet:getAll` / `create` / `update` / `delete` / `search` / `toggleFavorite` | 雙向 |
| **個人檔案** | `profile:list` / `create` / `load` / `delete` / `rename` / `duplicate` | 雙向 |
| **遠端** | `remote:start-server` / `stop-server` / `connect` / `disconnect` / `status` | 雙向 |
| **訊息存檔** | `archiveMessages` / `loadArchived` / `clearArchive` | 渲染 → 主 |

---

## 資料儲存

| 平台 | 路徑 |
|------|------|
| Windows | `%APPDATA%/better-agent-terminal/` |
| macOS | `~/Library/Application Support/better-agent-terminal/` |
| Linux | `~/.config/better-agent-terminal/` |

| 檔案 | 用途 |
|------|------|
| `workspaces.json` | 工作區狀態 |
| `settings.json` | 應用設定 |
| `snippets.json` | 程式碼片段 |
| `profiles/` | 配置檔快照 |
| `message-archives/` | Claude 訊息存檔（JSONL） |

---

## 自動設定的環境變數

應用啟動終端時會自動注入以下環境變數：

```
LANG=en_US.UTF-8
LC_ALL=en_US.UTF-8
TERM=xterm-256color
COLORTERM=truecolor
FORCE_COLOR=3
TERM_PROGRAM=better-terminal
TERM_PROGRAM_VERSION=1.0
CI=                          # 清空，避免被誤判為 CI 環境
```

---

## 發佈與安裝

| 平台 | 格式 |
|------|------|
| Windows | NSIS 安裝程式 + ZIP 便攜版 |
| macOS | DMG（Universal Binary，支援 Intel + Apple Silicon） |
| Linux | AppImage |

版本格式：`1.YY.MMDDHHmmss`，GitHub Actions 自動構建。

---

## 與 claude-tmux (cc-menu) 的比較

| 面向 | cc-menu | Better Agent Terminal |
|------|---------|----------------------|
| 技術 | 純 zsh 腳本（~1000 行） | Electron + React + TypeScript |
| 終端 | 依賴 tmux | 內建 xterm.js |
| AI 整合 | 讀取 JSONL + 呼叫 `claude -p` | 直接整合 Claude Agent SDK |
| 佈局 | 文字 TUI，單欄 | GUI，Google Meet 風格雙欄 |
| 部署 | symlink 一個檔案 | 需安裝 Electron 應用 |
| 平台 | macOS/Linux（需 tmux + zsh） | Windows / macOS / Linux |
| 資源佔用 | 極低（shell 程序） | 較高（Electron 進程） |
| 檔案瀏覽 | 無 | 內建樹狀瀏覽 + 語法高亮 |
| Git | 無 | 內建分支/提交/Diff 檢視 |
| 遠端 | 無（依賴 SSH + tmux） | WebSocket 伺服器/客戶端 |
| 擴充性 | 有限 | 多代理預設、外掛架構 |
| 適合場景 | 輕量 SSH 遠端、tmux 重度使用者 | 本機開發、需要 GUI 的使用者 |
