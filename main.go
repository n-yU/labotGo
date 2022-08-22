package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var (
	logger *log.Logger
	client *socketmode.Client
)

func main() {
	// ロガー生成 & 環境変数読み込み
	logger = log.New(os.Stdout, "[labotGo] ", log.Ldate|log.Ltime)
	loadEnv()

	// トークン確認
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		logger.Fatalln("環境変数 APP_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		logger.Fatalln("環境変数 APP_TOKEN は \"xapp-\" を prefix として持つ必要があります")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		logger.Fatalln("環境変数 BOT_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(botToken, "xoxb-") {
		logger.Fatalln("環境変数 BOT_TOKEN は \"xoxb-\" を prefix として持つ必要があります")
	}

	// クライアント生成
	api := slack.New(botToken, slack.OptionDebug(DebugMode), slack.OptionLog(logger), slack.OptionAppLevelToken(appToken))
	client = socketmode.New(api, socketmode.OptionDebug(DebugMode), socketmode.OptionLog(logger))

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.Println("Socket Mode で Slack に接続しています...")
			case socketmode.EventTypeConnectionError:
				logger.Println("接続に失敗しました．後ほど再試行します...")
			case socketmode.EventTypeConnected:
				logger.Println("Sokect Mode で Slack に接続しました")
			case socketmode.EventTypeHello:
				logger.Println("Hello イベントの受信に成功しました")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					logger.Printf("Ignored %+v \n", evt)
					continue
				}
				logger.Printf("Event を受け取りました: %+v\n", eventsAPIEvent)

				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				default:
					client.Debugf("未対応の Events API Event を受け取りました")
				}
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					logger.Printf("Ignored %+v\n", evt)
					continue
				}
				logger.Printf("Interaction を受け取りました: %+v\n", callback)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					logger.Printf("Ignored %+v\n", evt)
					continue
				}
				logger.Printf("Slach Command を受け取りました: %+v\n", cmd)

				payload := listen(cmd)
				client.Ack(*evt.Request, payload)
			default:
				logger.Printf("不明なイベントタイプ %s を受け取りました\n", evt.Type)
			}
		}
	}()

	// 動作チェック
	DefaultCh := fmt.Sprintf("#%s", os.Getenv("DEFAULT_CHANNEL"))
	launch_text := slack.MsgOptionText(
		fmt.Sprintf(":white_check_mark: *<https://github.com/n-yU/labotGo|labotGo> v%s を起動しました*\n", Version), false)
	_, _, err := api.PostMessage(DefaultCh, launch_text)
	if err != nil {
		logger.Println("labotGo の起動に失敗しました")
		logger.Printf("Tips: labotGo は起動時に動作チェックのため，デフォルトチャンネル %s にも追加する必要があります\n", DefaultCh)
		logger.Println("デフォルトチャンネルは .env から変更することもできます")
		logger.Fatal(err)
	}
	logger.Printf("labotGo %s を起動しました\n", Version)

	// Socket Mode
	client.Run()
}
