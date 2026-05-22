package services

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Entry struct {
	URL             string   `mapstructure:"url" json:"url"`
	Pages           []string `mapstructure:"pages" json:"pages"`                               // 支持 "1", "3-5" 格式
	PagesInt        []int    `mapstructure:"-" json:"-"`                                       // 解析后的整数集合
	PdfNameTemplate string   `mapstructure:"pdfNameTemplate" json:"pdfNameTemplate,omitempty"` // 非空时覆盖全局 pdfNameTemplate
}

type Config struct {
	Cookie          string  `mapstructure:"cookie" json:"cookie"`
	OutPath         string  `mapstructure:"outPath" json:"outPath"`
	Limit           int     `mapstructure:"limit" json:"limit"`
	Entries         []Entry `mapstructure:"entries" json:"entries"` // 替换原有的 urls 和 page
	ConvertToPdf    bool    `mapstructure:"convertToPdf" json:"convertToPdf"`
	PdfNameTemplate string  `mapstructure:"pdfNameTemplate" json:"pdfNameTemplate,omitempty"` // 可用占位符 {chapter}；空则使用下载下来的卷名
}

func (c *Config) LoadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		log.Fatal(err)
	}

	// Limit max 200
	if c.Limit <= 0 || c.Limit > 200 {
		c.Limit = 200
	}

	// 遍历每个 Entry，解析其 Pages
	for i := range c.Entries {
		entry := &c.Entries[i]
		parsedPages := c.parsePages(entry.Pages)
		entry.PagesInt = parsedPages // 需要新增 PagesInt 字段
	}
}

// config.go 文件中新增方法
func (c *Config) parsePages(rawPages []string) []int {
	var parsedPages []int

	for _, p := range rawPages {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// 处理范围格式 (如 "5-10")
		if strings.Contains(p, "-") {
			parts := strings.Split(p, "-")
			if len(parts) != 2 {
				log.Printf("无效的范围格式: %s", p)
				continue
			}

			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			if err1 != nil || err2 != nil {
				log.Printf("范围值必须为数字: %s", p)
				continue
			}

			if start > end {
				log.Printf("范围起始值不能大于结束值: %s", p)
				continue
			}

			// 生成连续数字
			for i := start; i <= end; i++ {
				parsedPages = append(parsedPages, i)
			}
			continue
		}

		// 处理单卷号
		num, err := strconv.Atoi(p)
		if err != nil {
			log.Printf("无效的卷号格式: %s", p)
			continue
		}
		parsedPages = append(parsedPages, num)
	}

	// 去重并排序
	return processPages(parsedPages)
}

// 辅助函数：去重排序
func processPages(pages []int) []int {
	seen := make(map[int]struct{})
	result := make([]int, 0, len(pages))

	// 去重
	for _, num := range pages {
		if _, exists := seen[num]; exists {
			log.Printf("发现重复卷号: %d", num) // 用 %d 输出整数
			continue                      // 跳过已存在的值
		}
		// 不存在则添加
		seen[num] = struct{}{}
		result = append(result, num)
	}

	// 排序
	sort.Ints(result)
	return result
}
