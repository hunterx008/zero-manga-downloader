package services

import (
	"os"
	"testing"
)

// testZerobyDetailsURL 用于联网集成测试：PC 详情页（内嵌 mangaDownloadChapters），当前站点上可访问。
// 若站点再次改版导致 404，请替换为新的详情页 URL（浏览器地址栏复制）。
const testZerobyDetailsURL = "https://www.zerobywai.com/pc/details/?kuid=21936"

func TestGetComicPageInfo(t *testing.T) {
	cli := NewClient()
	zd := &ZeroDownload{
		Cookie:  "",
		OutPath: "download",
		Client:  cli,
		Limit:   12,
		Pages:   []int{1, 2, 3},
	}

	// 详情页：走 GetComicPageInfoDetails，覆盖路由分发
	comic := zd.GetComicPageInfo(testZerobyDetailsURL)

	if len(comic.Pages) < 3 {
		t.Fatalf("There should be no less than 3 chapters. Pages: %d", len(comic.Pages))
	}

	// 章节列表与阅读页 URL 应能解析出来（不依赖 Cookie）。
	withName := 0
	for _, page := range comic.Pages {
		if page.Name == "" {
			continue
		}
		withName++
		if page.PageUrl == "" {
			t.Fatalf("Page %s has empty PageUrl", page.Name)
		}
	}
	if withName < 3 {
		t.Fatalf("Expected at least 3 chapters with names, got: %d", withName)
	}

	// 无 Cookie 时（如 GitHub Actions）站点可能返回空图列表，默认不强制要求 Total>0。
	// 本地有有效 Cookie 且需严格校验图片数时：TEST_REQUIRE_CHAPTER_IMAGES=1 go test ./services -run TestGetComicPageInfo
	if os.Getenv("TEST_REQUIRE_CHAPTER_IMAGES") == "1" {
		for _, page := range comic.Pages {
			if page.Name != "" && page.Total == 0 {
				t.Fatalf("Page %s has 0 images (set cookie or unset TEST_REQUIRE_CHAPTER_IMAGES)", page.Name)
			}
		}
	}
}

func TestGetComicPageInfo_Sp(t *testing.T) {
	cli := NewClient()
	c := &Config{
		OutPath: "download",
		Entries: []Entry{
			{
				URL:   testZerobyDetailsURL,
				Pages: []string{"1", "3-5", "7", "4-8"},
			},
			{
				URL:   testZerobyDetailsURL,
				Pages: []string{"2-5", "4-8"},
			},
		},
	}

	for _, entry := range c.Entries {
		parsedPages := c.parsePages(entry.Pages)
		entry.PagesInt = parsedPages

		zd := &ZeroDownload{
			Cookie:  c.Cookie,
			OutPath: c.OutPath,
			Client:  cli,
			Limit:   c.Limit,
			Pages:   entry.PagesInt,
		}
		zd.GetComicPageInfo(entry.URL)
	}
}

func TestGetComicPageInfoDetails(t *testing.T) {
	cli := NewClient()
	zd := &ZeroDownload{
		Cookie:  "",
		OutPath: "download",
		Client:  cli,
		Limit:   12,
		Pages:   []int{1, 2, 3},
	}

	comic := zd.GetComicPageInfoDetails(testZerobyDetailsURL)

	if comic.Title == "" {
		t.Fatal("Title should not be empty")
	}

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
