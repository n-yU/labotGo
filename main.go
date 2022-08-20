package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

const (
	Version = "1.0.0-alpha"
)

var logger *log.Logger

func loadEnv() {
	// 環境変数 読み込み
	// ref.) https://zenn.dev/a_ichi1/articles/c9f3870350c5e2
	err := godotenv.Load(".env")
	if err != nil {
		logger.Println("環境変数の読み込みに失敗しました")
		logger.Fatal(err)
	}
}

func main() {
	logger = log.New(os.Stdout, "[labotGo] ", log.Ldate|log.Ltime)
	loadEnv()

	// クライアント生成
	api := slack.New(os.Getenv("BOT_TOKEN"))
	DefaultCh := fmt.Sprintf("#%s", os.Getenv("DEFAULT_CHANNEL"))

	// 動作チェック
	launch_text := slack.MsgOptionText(
		fmt.Sprintf(":white_check_mark: *<https://github.com/n-yU/labotGo|labotGo> v%s を起動しました*\n", Version), false)
	_, _, err := api.PostMessage(DefaultCh, launch_text)
	if err != nil {
		logger.Println("labotGo の起動に失敗しました")
		logger.Printf("Tips: labotGo は起動時に動作チェックのため，%s にも追加する必要があります\n", DefaultCh)
		logger.Fatal(err)
	}
}
