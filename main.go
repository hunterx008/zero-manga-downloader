package main

import (
	"fmt"
	"log"
	"os"

	service "github.com/hunterx008/zero-manga-downloader/services"
)

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func main() {
	// Stderr tends to show up immediately in IDE terminals; stdout can look "stuck" in Debug Console.
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)

	conf := &service.Config{}
	conf.LoadConfig("./config.json")

	cli := service.NewClient()

	for _, entry := range conf.Entries {
		pdfTpl := conf.PdfNameTemplate
		if entry.PdfNameTemplate != "" {
			pdfTpl = entry.PdfNameTemplate
		}
		zd := &service.ZeroDownload{
			Cookie:          conf.Cookie,
			OutPath:         conf.OutPath,
			Client:          cli,
			Limit:           conf.Limit,
			Pages:           entry.PagesInt, // 使用当前 Entry 的 Pages
			ConvertToPdf:    conf.ConvertToPdf,
			PdfNameTemplate: pdfTpl,
		}
		// 下载当前 Entry 的 URL
		zd.DownloadComic([]string{entry.URL})
	}

	log.Printf("DONE! Press any key to exit!/下载完成! 任意按键退出!")
	if stdinIsTerminal() {
		_, _ = fmt.Scanln()
	} else {
		log.Printf("Non-interactive stdin: exiting without waiting for a key./标准输入非终端：不等待按键，直接退出。")
	}
}
