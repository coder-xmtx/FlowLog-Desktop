// Package design 是流记的界面设计系统。
//
// 它只关心「长什么样」：颜色角色、字号、圆角、动效与一组基础组件，
// 不含任何业务状态。页面（internal/pages）与外壳（internal/view）都从这里取样式。
//
// 配色遵循 Material Design 3 的角色划分：primary / secondary / tertiary 三组强调色，
// surface 由低到高五个 container 层级，再加上 outline 与 on-* 文字色。
// 浅色是淡蓝底 + 白色卡片，深色是深靛蓝底 + 略亮的容器层，避免纯黑与高对比刺眼。
package design

import "github.com/egoist/mygo/ui"

// Palette 是一套完整的颜色角色。
//
// 浅色与深色各一份，字段一一对应；渲染时只认角色，不认具体色值。
type Palette struct {
	// Dark 表示这是深色板。
	Dark bool

	// Background 是窗口底色，SurfaceContainer* 是叠在上面的容器层，
	// 数字越大越靠近前景（卡片、输入框、悬浮条）。
	Background              ui.Color
	SurfaceContainerLowest  ui.Color
	SurfaceContainerLow     ui.Color
	SurfaceContainer        ui.Color
	SurfaceContainerHigh    ui.Color
	SurfaceContainerHighest ui.Color

	// OnSurface 是主要文字色，OnSurfaceVariant 是次要文字与图标色。
	OnSurface        ui.Color
	OnSurfaceVariant ui.Color

	// Outline 是强调的描边，OutlineVariant 是分隔线与弱描边。
	Outline        ui.Color
	OutlineVariant ui.Color

	// 主色：主要按钮、选中态、焦点环。
	Primary            ui.Color
	OnPrimary          ui.Color
	PrimaryContainer   ui.Color
	OnPrimaryContainer ui.Color

	// 辅色：次级按钮、筛选芯片、侧边栏选中指示。
	Secondary            ui.Color
	OnSecondary          ui.Color
	SecondaryContainer   ui.Color
	OnSecondaryContainer ui.Color

	// 第三色：成功、已复习等「正向」强调。
	Tertiary            ui.Color
	OnTertiary          ui.Color
	TertiaryContainer   ui.Color
	OnTertiaryContainer ui.Color

	// 错误色：删除、校验失败。
	Error            ui.Color
	OnError          ui.Color
	ErrorContainer   ui.Color
	OnErrorContainer ui.Color

	// 搜索命中的高亮：暖琥珀色，和蓝紫色界面拉开对比。
	Highlight   ui.Color
	OnHighlight ui.Color

	// 提示气泡与 Toast 用的反色，保证在背景上足够醒目。
	Inverse     ui.Color
	InverseText ui.Color
}

// Light 返回浅色板：淡蓝底、近白卡片、靛蓝主色。
func Light() Palette {
	return Palette{
		Dark:                    false,
		Background:              ui.Hex("#F2F5FC"),
		SurfaceContainerLowest:  ui.Hex("#FFFFFF"),
		SurfaceContainerLow:     ui.Hex("#FAFBFF"),
		SurfaceContainer:        ui.Hex("#EFF2FA"),
		SurfaceContainerHigh:    ui.Hex("#E8EDF8"),
		SurfaceContainerHighest: ui.Hex("#E1E7F4"),

		OnSurface:        ui.Hex("#191C2B"),
		OnSurfaceVariant: ui.Hex("#565C75"),

		Outline:        ui.Hex("#767D99"),
		OutlineVariant: ui.Hex("#D3D9EA"),

		Primary:            ui.Hex("#4C5BD4"),
		OnPrimary:          ui.Hex("#FFFFFF"),
		PrimaryContainer:   ui.Hex("#DFE2FF"),
		OnPrimaryContainer: ui.Hex("#12184C"),

		Secondary:            ui.Hex("#5A607D"),
		OnSecondary:          ui.Hex("#FFFFFF"),
		SecondaryContainer:   ui.Hex("#DFE2F5"),
		OnSecondaryContainer: ui.Hex("#171A2B"),

		Tertiary:            ui.Hex("#00686F"),
		OnTertiary:          ui.Hex("#FFFFFF"),
		TertiaryContainer:   ui.Hex("#9CF0F8"),
		OnTertiaryContainer: ui.Hex("#002022"),

		Error:            ui.Hex("#BA1A1A"),
		OnError:          ui.Hex("#FFFFFF"),
		ErrorContainer:   ui.Hex("#FFDAD6"),
		OnErrorContainer: ui.Hex("#410002"),

		Highlight:   ui.Hex("#FFE08A"),
		OnHighlight: ui.Hex("#3D2E00"),

		Inverse:     ui.Hex("#2C2F3D"),
		InverseText: ui.Hex("#F2F3FB"),
	}
}

// Dark 返回深色板：深靛蓝底、偏紫的容器层、浅靛蓝主色。
func Dark() Palette {
	return Palette{
		Dark:                    true,
		Background:              ui.Hex("#0F1120"),
		SurfaceContainerLowest:  ui.Hex("#0A0C17"),
		SurfaceContainerLow:     ui.Hex("#151829"),
		SurfaceContainer:        ui.Hex("#1B1F33"),
		SurfaceContainerHigh:    ui.Hex("#22263D"),
		SurfaceContainerHighest: ui.Hex("#292D47"),

		OnSurface:        ui.Hex("#E4E6F4"),
		OnSurfaceVariant: ui.Hex("#B0B5CC"),

		Outline:        ui.Hex("#7C8199"),
		OutlineVariant: ui.Hex("#393E57"),

		Primary:            ui.Hex("#BAC3FF"),
		OnPrimary:          ui.Hex("#16206B"),
		PrimaryContainer:   ui.Hex("#303C92"),
		OnPrimaryContainer: ui.Hex("#DEE0FF"),

		Secondary:            ui.Hex("#C2C5DD"),
		OnSecondary:          ui.Hex("#2C3047"),
		SecondaryContainer:   ui.Hex("#3B3F55"),
		OnSecondaryContainer: ui.Hex("#DEE1F9"),

		Tertiary:            ui.Hex("#4CD9E3"),
		OnTertiary:          ui.Hex("#00363B"),
		TertiaryContainer:   ui.Hex("#004F55"),
		OnTertiaryContainer: ui.Hex("#9CF0F8"),

		Error:            ui.Hex("#FFB4AB"),
		OnError:          ui.Hex("#690005"),
		ErrorContainer:   ui.Hex("#93000A"),
		OnErrorContainer: ui.Hex("#FFDAD6"),

		Highlight:   ui.Hex("#6B4E00"),
		OnHighlight: ui.Hex("#FFE08A"),

		Inverse:     ui.Hex("#E4E6F4"),
		InverseText: ui.Hex("#1B1E2C"),
	}
}

// For 按深浅返回对应的颜色板。
func For(dark bool) Palette {
	if dark {
		return Dark()
	}
	return Light()
}
