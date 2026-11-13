package view

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// bingEndpoint 是必应每日一图的第三方接口（bing.biturl.top）。
//
// 返回 JSON：start_date / end_date / url / copyright / copyright_link。
// 图片本身在 url 指向的 www.bing.com JPEG 上。
const bingEndpoint = "https://bing.biturl.top/?resolution=1920&format=json&index=0&mkt=zh-CN"

// bingMeta 是接口返回的字段。
type bingMeta struct {
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	URL           string `json:"url"`
	Copyright     string `json:"copyright"`
	CopyrightLink string `json:"copyright_link"`
}

// loadBackground 先读当天缓存，没有再去后台取图。
//
// 网络请求都在 goroutine 里；结果通过 Window.Update 回到主线程写入，
// 这样视图读 app.bg 时没有数据竞争。
func (app *App) loadBackground() {
	if bmp, credit, ok := cachedBackground(); ok {
		app.bg, app.bgCopyright = bmp, credit
		return
	}
	app.bgLoading = true
	go app.fetchBackground()
}

// fetchBackground 在后台拉取必应每日一图。
func (app *App) fetchBackground() {
	bmp, credit, err := fetchBingImage()
	app.publishBackground(bmp, credit, err)
}

// publishBackground 回到主线程写入背景（失败则保留渐变兜底）。
func (app *App) publishBackground(bmp *ui.Bitmap, credit string, err error) {
	apply := func() {
		app.bgLoading = false
		if err != nil || bmp == nil {
			app.bg = nil
			return
		}
		app.bg, app.bgCopyright = bmp, credit
	}
	if app.win != nil {
		app.win.Update(apply)
		return
	}
	apply()
}

// fetchBingImage 取元数据、下载图片、解码并写缓存。
func fetchBingImage() (*ui.Bitmap, string, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	meta, err := getBingMeta(client)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Get(meta.URL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载背景图：HTTP %d", resp.StatusCode)
	}
	// 5 MB 对 1920×1080 的 JPEG 足够，也防止异常响应吃满内存。
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, "", err
	}

	bmp, err := ui.DecodeBitmap(data)
	if err != nil {
		return nil, "", err
	}
	cacheBackground(data, meta.Copyright)
	return bmp, meta.Copyright, nil
}

func getBingMeta(client *http.Client) (bingMeta, error) {
	var meta bingMeta
	resp, err := client.Get(bingEndpoint)
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("必应接口：HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&meta); err != nil {
		return meta, err
	}
	if meta.URL == "" {
		return meta, fmt.Errorf("必应接口没有返回图片地址")
	}
	return meta, nil
}

// backgroundCacheDir 返回背景图缓存目录（用户缓存目录下的 FlowLog/background）。
func backgroundCacheDir() (string, error) {
	base, err := mygo.App.Path(mygo.PathCache)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "background")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// cachedBackground 读当天缓存下来的图片与署名。
func cachedBackground() (*ui.Bitmap, string, bool) {
	dir, err := backgroundCacheDir()
	if err != nil {
		return nil, "", false
	}
	day := model.Today()
	data, err := os.ReadFile(filepath.Join(dir, day+".jpg"))
	if err != nil {
		return nil, "", false
	}
	bmp, err := ui.DecodeBitmap(data)
	if err != nil {
		return nil, "", false
	}
	credit, _ := os.ReadFile(filepath.Join(dir, day+".txt"))
	return bmp, strings.TrimSpace(string(credit)), true
}

// cacheBackground 把图片和署名写进缓存；失败不影响界面。
func cacheBackground(data []byte, credit string) {
	dir, err := backgroundCacheDir()
	if err != nil {
		return
	}
	day := model.Today()
	if err := os.WriteFile(filepath.Join(dir, day+".jpg"), data, 0o644); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, day+".txt"), []byte(credit), 0o644)
	cleanBackgroundCache(dir, 7)
}

// cleanBackgroundCache 只保留最近 keep 天的缓存。
func cleanBackgroundCache(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	type cacheDay struct {
		name  string
		mtime time.Time
	}
	seen := make(map[string]time.Time)
	for _, e := range entries {
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".jpg" && ext != ".txt" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		day := strings.TrimSuffix(name, ext)
		if t, ok := seen[day]; !ok || info.ModTime().After(t) {
			seen[day] = info.ModTime()
		}
	}
	if len(seen) <= keep {
		return
	}
	days := make([]cacheDay, 0, len(seen))
	for day, t := range seen {
		days = append(days, cacheDay{day, t})
	}
	sort.Slice(days, func(i, j int) bool { return days[i].mtime.After(days[j].mtime) })
	for _, d := range days[keep:] {
		_ = os.Remove(filepath.Join(dir, d.name+".jpg"))
		_ = os.Remove(filepath.Join(dir, d.name+".txt"))
	}
}
