package main

import (
	"fmt"
	"log"
	"os"

	service "github.com/hunterx008/zero-manga-downloader/services"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)

	conf := &service.Config{}
	conf.LoadConfig("./config.json")

	cli := service.NewClient()

	for _, entry := range conf.Entries {
		zd := &service.ZeroDownload{
			Cookie:  conf.Cookie,
			OutPath: conf.OutPath,
			Client:  cli,
			Limit:   conf.Limit,
			Pages:   entry.PagesInt, // 使用当前 Entry 的 Pages
		}
		// 下载当前 Entry 的 URL
		zd.DownloadComic([]string{entry.URL})
	}

	log.Printf("DONE! Press any key to exit!/下载完成! 任意按键退出!")
	_, err := fmt.Scanln()
	if err != nil {
		//TODO
	}
}
