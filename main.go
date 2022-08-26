// labotGo
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/listen"
	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// labotGo
func main() {
	Logger = log.New(os.Stdout, "[labotGo] ", log.Ldate|log.Ltime)
	LoadEnv()

	// トークン 確認
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		Logger.Fatalln("環境変数 APP_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		Logger.Fatalln("環境変数 APP_TOKEN は \"xapp-\" を prefix として持つ必要があります")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		Logger.Fatalln("環境変数 BOT_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(botToken, "xoxb-") {
		Logger.Fatalln("環境変数 BOT_TOKEN は \"xoxb-\" を prefix として持つ必要があります")
	}

	// クライアント 生成
	Api = slack.New(botToken, slack.OptionDebug(DebugMode), slack.OptionLog(Logger), slack.OptionAppLevelToken(appToken))
	SocketModeClient = socketmode.New(Api, socketmode.OptionDebug(DebugMode), socketmode.OptionLog(Logger))

	go func() {
		for evt := range SocketModeClient.Events {
			// イベントタイプ別 ハンドリング
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				Logger.Println("Socket Mode で Slack に接続しています...")
			case socketmode.EventTypeConnectionError:
				Logger.Println("接続に失敗しました．後ほど再試行します...")
			case socketmode.EventTypeConnected:
				Logger.Println("Sokect Mode で Slack に接続しました")
			case socketmode.EventTypeHello:
				Logger.Println("Hello イベントの受信に成功しました")
			case socketmode.EventTypeEventsAPI:
				// イベントAPI
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					Logger.Printf("Ignored %+v \n", evt)
					continue
				}
				Logger.Printf("Event を受け取りました: %+v\n", eventsAPIEvent)
				SocketModeClient.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				default:
					SocketModeClient.Debugf("未対応の Events API Event を受け取りました")
				}
			case socketmode.EventTypeInteractive:
				// インタラクション
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					Logger.Printf("Ignored %+v\n", evt)
					continue
				}
				Logger.Printf("Interaction を受け取りました\n")
				SocketModeClient.Ack(*evt.Request)

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					Logger.Printf("Block Action を受け取りました: %+v\n", callback.BlockActionState)
				default:
					Logger.Printf("未対応の Block Action を受け取りました")
				}

				// インタラクション 受信処理
				if err := listen.BlockAction(callback); err != nil {
					Logger.Fatal(err)
				}
			case socketmode.EventTypeSlashCommand:
				// スラッシュコマンド
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					Logger.Printf("Ignored %+v\n", evt)
					continue
				}
				Logger.Printf("Slach Command を受け取りました: %+v\n", cmd)
				SocketModeClient.Ack(*evt.Request)

				// コマンド 受信処理
				if err := listen.Command(cmd); err != nil {
					Logger.Fatal(err)
				}
			default:
				Logger.Printf("不明なイベントタイプ %s を受け取りました\n", evt.Type)
			}
		}
	}()

	// 動作チェック
	defaultCh := fmt.Sprintf("#%s", os.Getenv("DEFAULT_CHANNEL"))
	if err := post.Start(defaultCh); err != nil {
		Logger.Println("labotGo の起動に失敗しました")
		Logger.Printf("Tips: labotGo は起動時に動作チェックのため，デフォルトチャンネル %s にも追加する必要があります\n", defaultCh)
		Logger.Println("Tips: デフォルトチャンネルは .env から変更することもできます")
		Logger.Fatal(err)
	}
	Logger.Printf("labotGo %s を起動しました\n", Version)

	// 初回起動チェック（データファイル生成）
	isFirstRun, err := checkFirstRun()
	if isFirstRun {
		if err != nil {
			Logger.Println("初回起動のため，データファイルの生成を試みましたが失敗しました")
			Logger.Println("詳しくは次のエラーを確認してください")
			Logger.Fatal(err)
		}
		Logger.Println("初回起動のため，以下のデータファイルを生成しました")
		Logger.Printf("- メンバーデータ: %s\n", MemberDataPath)
		Logger.Printf("- チームデータ  : %s\n", TeamDataPath)
	}
	Logger.Println("Tips: ボットが正常に動作しなくなる恐れがあるため，メンバー／チームデータは直接編集しないでください")

	// Socket Mode
	SocketModeClient.Run()
}

// 初回起動チェック
func checkFirstRun() (isFirstRun bool, err error) {
	// データファイル存在 チェック
	isMemberData, isTeamData := FileExists(MemberDataPath), FileExists(TeamDataPath)
	if isMemberData && isTeamData {
		return isFirstRun, err
	}

	// データファイル 生成
	isFirstRun = true
	if _, err := os.Create(MemberDataPath); err != nil {
		return isFirstRun, err
	}
	if _, err := os.Create(TeamDataPath); err != nil {
		return isFirstRun, err
	}

	// メンバーデータ 初期設定
	memberData := map[string][]string{}
	response, err := SocketModeClient.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	memberData[response.UserID] = []string{"all"}
	if err := data.UpdateMember(memberData); err != nil {
		Logger.Fatal(err)
	}

	// チームデータ 初期設定
	teamData := map[string][]string{}
	teamData["all"] = []string{response.UserID}
	if err := data.UpdateTeam(teamData); err != nil {
		Logger.Fatal(err)
	}

	return isFirstRun, err
}
