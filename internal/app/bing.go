package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
//
// 加一个随机的 _ 参数，避开接口与 CDN 的缓存：0 点刚过时缓存里可能还是
// 前一天的图，直接读缓存就会把旧图当成新图。
const bingEndpoint = "https://bing.biturl.top/?resolution=1920&format=json&index=0&mkt=zh-CN"

// bingMeta 是接口返回的字段。
type bingMeta struct {
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	URL           string `json:"url"`
	Copyright     string `json:"copyright"`
	CopyrightLink string `json:"copyright_link"`
}

// imageDay 是图片所属日期的本地日期字符串（2006-01-02）。
//
// 接口的 end_date 是 "20261120" 这样的格式；解析失败时退回今天。
func (m bingMeta) imageDay() string {
	if len(m.EndDate) == 8 {
		if t, err := time.Parse("20060102", m.EndDate); err == nil {
			return t.Format(model.DateLayout)
		}
	}
	return model.Today()
}

// 背景图的重试节奏：0 点刚过时必应可能还没换图，隔一会儿再问一次。
const (
	bgRetryInterval = 15 * time.Minute
	bgRetryLimit    = 8
)

// LoadBackground 先读今天的缓存，没有再去后台取图。
//
// 取图时如果必应还没有换到新一天，会先把旧图显示出来，并定时重试，
// 直到拿到今天的图为止——否则 0 点后打开应用会一整天停在旧图上。
func (a *App) LoadBackground() {
	day := model.Today()
	if bmp, credit, ok := cachedBackground(day); ok {
		a.BG, a.BGCopyright = bmp, credit
		return
	}
	// 先显示最近一张缓存，避免一片空白；新图到了再替换。
	if bmp, credit, ok := latestCachedBackground(); ok {
		a.BG, a.BGCopyright = bmp, credit
	}
	a.BGLoading = true
	go a.fetchBackgroundLoop()
}

// fetchBackgroundLoop 拉取今天的图；必应还没换图时隔一会儿再试。
func (a *App) fetchBackgroundLoop() {
	for attempt := 0; attempt <= bgRetryLimit; attempt++ {
		bmp, credit, day, err := fetchBingImage()
		a.publishBackground(bmp, credit, day, err)
		if err == nil && day == model.Today() {
			return
		}
		if attempt == bgRetryLimit {
			return
		}
		time.Sleep(bgRetryInterval)
	}
}

// publishBackground 回到主线程写入背景（失败则保留原有图片）。
func (a *App) publishBackground(bmp *ui.Bitmap, credit, day string, err error) {
	apply := func() {
		a.BGLoading = false
		if err != nil || bmp == nil {
			return
		}
		a.BG, a.BGCopyright = bmp, credit
		a.BGDay = day
	}
	if a.Win != nil {
		a.Win.Update(apply)
		return
	}
	apply()
}

// fetchBingImage 取元数据、下载图片、解码并写缓存，同时返回图片所属日期。
func fetchBingImage() (*ui.Bitmap, string, string, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	meta, err := getBingMeta(client)
	if err != nil {
		return nil, "", "", err
	}
	day := meta.imageDay()

	// 这张图已经缓存过：直接读，省一次下载。
	if data, credit, ok := readCachedImage(day); ok {
		if bmp, err := ui.DecodeBitmap(data); err == nil {
			return bmp, credit, day, nil
		}
	}

	resp, err := client.Get(meta.URL)
	if err != nil {
		return nil, "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("下载背景图：HTTP %d", resp.StatusCode)
	}
	// 5 MB 对 1920×1080 的 JPEG 足够，也防止异常响应吃满内存。
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, "", "", err
	}

	bmp, err := ui.DecodeBitmap(data)
	if err != nil {
		return nil, "", "", err
	}
	cacheBackground(day, data, meta.Copyright)
	return bmp, meta.Copyright, day, nil
}

// getBingMeta 请求必应每日一图的接口，返回图片地址与版权信息。
//
// 参数 client 用来复用连接与超时设置。
func getBingMeta(client *http.Client) (bingMeta, error) {
	var meta bingMeta
	// 拼一个一次性参数，绕开接口与 CDN 对同一 URL 的缓存。
	resp, err := client.Get(bingEndpoint + "&_=" + strconv.FormatInt(time.Now().UnixNano(), 36))
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("必应接口：HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&meta); err != nil {
		return meta, fmt.Errorf("解析必应接口: %w", err)
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

// cachedBackground 读某一天缓存下来的图片与署名。
func cachedBackground(day string) (*ui.Bitmap, string, bool) {
	data, credit, ok := readCachedImage(day)
	if !ok {
		return nil, "", false
	}
	bmp, err := ui.DecodeBitmap(data)
	if err != nil {
		return nil, "", false
	}
	return bmp, credit, true
}

// readCachedImage 读某一天的图片字节与署名。
func readCachedImage(day string) ([]byte, string, bool) {
	dir, err := backgroundCacheDir()
	if err != nil {
		return nil, "", false
	}
	data, err := os.ReadFile(filepath.Join(dir, day+".jpg"))
	if err != nil {
		return nil, "", false
	}
	credit, _ := os.ReadFile(filepath.Join(dir, day+".txt"))
	return data, strings.TrimSpace(string(credit)), true
}

// latestCachedBackground 读缓存里最近一天的图片，用作新图到位前的占位。
func latestCachedBackground() (*ui.Bitmap, string, bool) {
	dir, err := backgroundCacheDir()
	if err != nil {
		return nil, "", false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", false
	}
	newest := ""
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".jpg") {
			continue
		}
		day := strings.TrimSuffix(name, ".jpg")
		if day > newest {
			newest = day
		}
	}
	if newest == "" {
		return nil, "", false
	}
	return cachedBackground(newest)
}

// cacheBackground 把图片和署名写进缓存；失败不影响界面。
func cacheBackground(day string, data []byte, credit string) {
	dir, err := backgroundCacheDir()
	if err != nil {
		return
	}
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
