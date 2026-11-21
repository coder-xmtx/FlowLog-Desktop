package lang

// Separator 是问候语与称呼之间的分隔：中文用全角逗号，英文用逗号加空格。
func Separator() string {
	if IsEnglish() {
		return ", "
	}
	return "，"
}
