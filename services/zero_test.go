package services

import (
	"testing"
)

func TestGetComicPageInfo(t *testing.T) {
	cli := NewClient()
	zd := &ZeroDownload{
		Cookie:  "",
		OutPath: "download",
		Client:  cli,
		Limit:   12,
		Pages:   []int{1, 2, 3},
	}

	// 使用新版URL格式，同时测试 GetComicPageInfo 的路由分发逻辑
	comic := zd.GetComicPageInfo("https://www.zerobywai.com/pc/manga_pc.php?kuid=21789")

	if len(comic.Pages) < 3 {
		t.Fatalf("There should be no less than 3 chapters. Pages: %d", len(comic.Pages))
	}

	for _, page := range comic.Pages {
		if page.Name != "" && page.Total == 0 {
			t.Fatalf("Page %s has 0 images", page.Name)
		}
	}
}

func TestGetComicPageInfo_Sp(t *testing.T) {
	// 测试配置解析
	cli := NewClient()
	c := &Config{
		OutPath: "download",
		Entries: []Entry{
			{
				URL:   "https://www.zerobywai.com/pc/manga_pc.php?kuid=21789",
				Pages: []string{"1", "3-5", "7", "4-8"},
			},
			{
				URL:   "https://www.zerobywai.com/pc/manga_pc.php?kuid=21789",
				Pages: []string{"2-5", "4-8"},
			},
		},
	}

	for _, entry := range c.Entries {
		parsedPages := c.parsePages(entry.Pages)
		entry.PagesInt = parsedPages // 需要新增 PagesInt 字段

		zd := &ZeroDownload{
			Cookie:  c.Cookie,
			OutPath: c.OutPath,
			Client:  cli,
			Limit:   c.Limit,
			Pages:   entry.PagesInt, // 使用当前 Entry 的 Pages
		}
		// 下载当前 Entry 的 URL
		zd.GetComicPageInfo(entry.URL)
	}
}

func TestGetComicPageInfoNew(t *testing.T) {
	cli := NewClient()
	zd := &ZeroDownload{
		Cookie:  "",
		OutPath: "download",
		Client:  cli,
		Limit:   12,
		Pages:   []int{1, 2, 3},
	}

	comic := zd.GetComicPageInfoNew("https://www.zerobywai.com/pc/manga_pc.php?kuid=21789")

	if comic.Title == "" {
		t.Fatal("Title should not be empty")
	}

	// 只下载了 pages 1,2,3，检查这些 page 是否存在
	foundPages := 0
	for _, page := range comic.Pages {
		if page.Name != "" {
			foundPages++
		}
	}
	if foundPages < 3 {
		t.Fatalf("Expected at least 3 chapters, got: %d", foundPages)
	}

	t.Logf("Title: %s, Total pages: %d", comic.Title, foundPages)
}
