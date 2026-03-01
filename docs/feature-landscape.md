# Agent Terminal 功能全景

整理兩個既有方案的功能，作為 agent-tmux 設計參考。

---

## A. cc-menu（前身，純 TUI）

技術：zsh 腳本 ~1000 行，依賴 tmux，零依賴安裝。

| 類別 | 功能 |
|------|------|
| Session 管理 | 依工作目錄樹狀分組顯示（`├─` / `└─`） |
| | 二級選單：attach / rename / kill |
| | 快速動作：New Claude、Resume Claude、Shell、Leave |
| Claude 偵測 | 即時偵測 Claude 程序，Braille spinner 動畫 |
| AI 摘要 | Claude Code Stop hook 自動生成，3 分鐘防抖 |
| | 快取於 `~/.cache/cc-menu/`（summary / name / debounce） |
| | 按 `s` 手動觸發 |
| 預覽 | AI 摘要 + pane 最後幾行輸出 |
| 整合 | tmux popup（`prefix + C-s`） |
| | SSH 自動啟動（偵測非 tmux 環境） |
| | tmux 內外皆可運作（switch-client / attach） |

**優勢**：極輕量、SSH 友善、無需 GUI、零資源佔用
**限制**：無檔案瀏覽、無 Git 整合、無遠端連線、擴充性有限

---

## B. Better Agent Terminal（參考對象，GUI）

技術：Electron + React + TypeScript，跨平台桌面應用。

| 類別 | 功能 |
|------|------|
| 工作區管理 | 按專案資料夾分組，拖放排序，分組篩選 |
| | 可分離獨立視窗，雙擊重新命名 |
| | 環境變數管理（全域 + 工作區層級） |
| 終端 | 內建 xterm.js，Google Meet 風格佈局（主面板 + 縮圖欄） |
| | 歷史保留（重啟後淡色顯示舊輸出） |
| | 縮圖實時預覽（最後 8 行） |
| Claude 整合 | 透過 Agent SDK 直接整合（非 CLI） |
| | 訊息串流、可摺疊思考過程 |
| | 權限管理（5 種模式） |
| | 會話恢復、暫停/喚醒 |
| | 圖像附件、Ctrl+P 檔案選擇器 |
| | Token 用量與成本顯示 |
| 多代理 | 預設支援 Claude / Gemini / Codex / Copilot |
| 檔案瀏覽 | 樹狀結構 + 語法高亮預覽（100+ 種檔案） |
| Git 整合 | 分支檢視、提交日誌、Diff 檢視器 |
| 片段管理 | SQLite 儲存，分類/標籤/搜尋 |
| 個人檔案 | 多配置檔快速切換，啟動參數指定 |
| 遠端連線 | WebSocket 伺服器/客戶端，Token 認證 |
| 訊息存檔 | 超過 200 條自動存檔，分頁載入 |

**優勢**：功能齊全、跨平台、GUI 操作直覺
**限制**：Electron 資源佔用高、無法 SSH 使用、安裝複雜

---

## C. 功能交叉比較

| 功能 | cc-menu | Better Agent Terminal |
|------|:-------:|:---------------------:|
| Session/工作區管理 | v | v |
| 樹狀分組 | v | v |
| Claude 程序偵測 | v | - |
| AI 摘要自動生成 | v | - |
| 終端預覽 | v（文字） | v（縮圖） |
| Claude SDK 直接整合 | - | v |
| 多代理支援 | - | v |
| 檔案瀏覽 | - | v |
| Git 整合 | - | v |
| 片段管理 | - | v |
| 遠端連線 | -（靠 SSH） | v |
| 會話持久化 | -（靠 tmux） | v |
| SSH 環境可用 | v | - |
| 資源佔用 | 極低 | 高 |
| 安裝複雜度 | 極低 | 高 |
