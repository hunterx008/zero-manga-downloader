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

	comic := zd.GetComicPageInfo("https://www.zerobywrar.com/plugin.php?id=jameson_manhua&c=index&a=bofang&kuid=18581")

	if len(comic.Pages) < 7 {
		t.Fatalf("There should be no less than 7 chapters. Pages: %d", len(comic.Pages))
	}

	pages := []Page{
		{Name: "1", Total: 182},
		{Name: "2", Total: 178},
		{Name: "3", Total: 178},
	}
	for i := 0; i < 3; i++ {
		cur := comic.Pages[i]
		if cur.Name != pages[i].Name || cur.Total != pages[i].Total {
			t.Fatalf("pages unequal, current: %s, %d, should be: %s, %d", cur.Name, cur.Total, pages[i].Name, pages[i].Total)
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
				URL:   "https://www.zerobywrar.com/plugin.php?id=jameson_manhua&c=index&a=bofang&kuid=18581",
				Pages: []string{"1", "3-5", "7", "4-8"},
			},
			{
				URL:   "https://www.zerobywrar.com/plugin.php?id=jameson_manhua&c=index&a=bofang&kuid=18672",
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
