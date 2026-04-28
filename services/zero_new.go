package services

import (
	"encoding/json"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// chapterInfo represents a chapter entry from the new page format's embedded JSON.
type chapterInfo struct {
	Zjid   string `json:"zjid"`
	Zjname string `json:"zjname"`
}

// GetComicPageInfoNew parses the new page format (pc/manga_pc.php?kuid=...).
// Chapter list is extracted from the embedded JavaScript variable `const chapters = [...]`.
// Chapter images are extracted via `img.manga-image` selector.
func (zd *ZeroDownload) GetComicPageInfoNew(url string) *Comic {
	res, err := zd.Requert(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("request failure/请求失败: %s", res.Status)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyStr := string(bodyBytes)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		log.Fatal(err)
	}

	// Extract title
	title := doc.Find("title").First().Text()
	reg := regexp.MustCompile(`[ \s]+`)
	til := reg.ReplaceAll([]byte(title), []byte{})

	comic := &Comic{
		Pages: []Page{},
		Title: string(til),
	}

	log.Printf("Preparing to obtain chapter information/准备获取章节信息(新版): %s", comic.Title)

	// Extract chapters from embedded JavaScript: const chapters = [...];
	chaptersRe := regexp.MustCompile(`const\s+chapters\s*=\s*(\[.*?\]);`)
	matches := chaptersRe.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		log.Fatal("Failed to extract chapters data from page/无法从页面提取章节数据")
	}

	var chapters []chapterInfo
	if err := json.Unmarshal([]byte(matches[1]), &chapters); err != nil {
		log.Fatalf("Failed to parse chapters JSON/解析章节JSON失败: %s", err.Error())
	}

	// Determine base URL for chapter pages
	// e.g. https://www.zerobywai.com/pc/manga_pc.php?kuid=21789 -> https://www.zerobywai.com/pc/
	baseUrl := url[:strings.LastIndex(url, "/")+1]

	wg := &sync.WaitGroup{}
	comic.Pages = make([]Page, len(chapters))

	for idx, ch := range chapters {
		// Filter by configured pages
		if len(zd.Pages) > 0 {
			found := false
			for _, p := range zd.Pages {
				if strconv.Itoa(p) == ch.Zjname {
					found = true
					break
				}
			}
			if !found {
				log.Printf("跳过卷号: %s (未在配置的 pages 列表中)", ch.Zjname)
				continue
			}
		}

		chapterUrl := baseUrl + "manga_read_pc.php?zjid=" + ch.Zjid
		page := Page{
			Name:    ch.Zjname,
			PageUrl: chapterUrl,
		}

		wg.Add(1)
		go func(i int, p Page) {
			defer wg.Done()

			res, err := zd.Requert(p.PageUrl)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			imgs := doc.Find("img.manga-image")
			if imgs.Length() == 0 {
				log.Printf(`The chapter image list is empty, possibly VIP-only content.
章节图片列表为空，可能为VIP专属内容，需要充值会员后找到cookie填入config.json里再重新执行: %s`, p.Name)
				comic.Pages[i] = p
				return
			}

			p.Total = imgs.Length()
			p.Urls = make([]string, p.Total)

			imgs.Each(func(j int, s *goquery.Selection) {
				if imageUrl, ok := s.Attr("src"); ok {
					// Handle protocol-relative URLs (e.g. //tupa.zerobywai.com/...)
					if strings.HasPrefix(imageUrl, "//") {
						imageUrl = "https:" + imageUrl
					}
					p.Urls[j] = imageUrl
				}
			})

			log.Printf("%s, chapter/章节: %s, total/总数: %d", comic.Title, p.Name, p.Total)
			comic.Pages[i] = p
		}(idx, page)
	}

	wg.Wait()

	return comic
}
