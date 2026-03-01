package tmux

// StatusInput 封裝三層偵測所需的所有輸入資料。
type StatusInput struct {
	HookStatus  *HookStatus // 第一層：hook 狀態檔案（可為 nil）
	PaneTitle   string      // 第二層：pane title
	PaneContent string      // 第三層：終端內容
}

// ResolveStatus 整合三層偵測，依優先順序回傳最終狀態。
// 優先順序：Hook 狀態（若有效）→ Pane Title → 終端內容。
func ResolveStatus(input StatusInput) SessionStatus {
	// 第一層：Hook 狀態檔案（最高優先）
	if input.HookStatus != nil && input.HookStatus.IsValid() {
		return input.HookStatus.Status
	}

	// 第二層：Pane Title
	titleStatus := DetectTitleStatus(input.PaneTitle)
	switch titleStatus {
	case TitleRunning:
		return StatusRunning
	case TitleDone:
		// Title 顯示完成，降級到第三層進一步判斷
		return DetectStatus(input.PaneContent)
	}

	// 第三層：終端內容
	return DetectStatus(input.PaneContent)
}
