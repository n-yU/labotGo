// labotGo
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/es"
	"github.com/n-yU/labotGo/listen"
	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/olivere/elastic/v7"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// labotGo
func main() {
	util.Logger = log.New(os.Stdout, "[labotGo] ", log.Ldate|log.Ltime)

	// 実行ファイルディレクトリ 取得
	exePath, err := os.Executable()
	if err != nil {
		util.Logger.Println("実行ファイルディレクトリの取得に失敗しました")
		util.Logger.Fatal(err)
	}
	util.Dir = filepath.Dir(exePath)
	util.Logger.Println("実行ファイルディレクトリ:", util.Dir)

	// 環境変数 読み込み
	util.LoadEnv()

	// トークン 確認
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		util.Logger.Fatalln("環境変数 APP_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		util.Logger.Fatalln("環境変数 APP_TOKEN は \"xapp-\" を prefix として持つ必要があります")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		util.Logger.Fatalln("環境変数 BOT_TOKEN が設定されていません")
	}
	if !strings.HasPrefix(botToken, "xoxb-") {
		util.Logger.Fatalln("環境変数 BOT_TOKEN は \"xoxb-\" を prefix として持つ必要があります")
	}

	// クライアント 生成
	util.Api = slack.New(botToken, slack.OptionDebug(util.DebugMode), slack.OptionLog(util.Logger), slack.OptionAppLevelToken(appToken))
	util.SocketModeClient = socketmode.New(util.Api, socketmode.OptionDebug(util.DebugMode), socketmode.OptionLog(util.Logger))

	go func() {
		for evt := range util.SocketModeClient.Events {
			// イベントタイプ別 ハンドリング
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				util.Logger.Println("Socket Mode で Slack に接続しています...")
			case socketmode.EventTypeConnectionError:
				util.Logger.Println("接続に失敗しました．後ほど再試行します...")
			case socketmode.EventTypeConnected:
				util.Logger.Println("Sokect Mode で Slack に接続しました")
			case socketmode.EventTypeHello:
				util.Logger.Println("Hello イベントの受信に成功しました")
			case socketmode.EventTypeEventsAPI:
				// イベントAPI
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					util.Logger.Printf("Ignored %+v \n", evt)
					continue
				}
				if util.DebugMode {
					util.Logger.Printf("Event を受け取りました: %+v\n", eventsAPIEvent)
				}
				util.SocketModeClient.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				default:
					util.SocketModeClient.Debugf("未対応の Events API Event を受け取りました")
				}
			case socketmode.EventTypeInteractive:
				// インタラクション
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					util.Logger.Printf("Ignored %+v\n", evt)
					continue
				}
				if util.DebugMode {
					util.Logger.Println("Interaction を受け取りました")
				}
				util.SocketModeClient.Ack(*evt.Request)

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					if util.DebugMode {
						util.Logger.Printf("Block Action を受け取りました: %+v\n", callback.BlockActionState)
					}
				default:
					util.Logger.Printf("未対応の Block Action を受け取りました")
				}

				// インタラクション 受信処理
				if err := listen.BlockAction(callback); err != nil {
					util.Logger.Fatal(err)
				}
			case socketmode.EventTypeSlashCommand:
				// スラッシュコマンド
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					util.Logger.Printf("Ignored %+v\n", evt)
					continue
				}
				util.Logger.Printf("Slach Command を受け取りました: %+v\n", cmd)
				util.SocketModeClient.Ack(*evt.Request)

				// コマンド 受信処理
				if err := listen.Command(cmd); err != nil {
					util.Logger.Fatal(err)
				}
			default:
				util.Logger.Printf("不明なイベントタイプ %s を受け取りました\n", evt.Type)
			}
		}
	}()

	// 動作チェック
	defaultCh := fmt.Sprintf("#%s", os.Getenv("DEFAULT_CHANNEL"))
	if err := post.Start(defaultCh); err != nil {
		util.Logger.Println("labotGo の起動に失敗しました")
		util.Logger.Printf("Tips: labotGo は起動時に動作チェックのため，デフォルトチャンネル %s にも追加する必要があります\n", defaultCh)
		util.Logger.Println("Tips: デフォルトチャンネルは .env から変更することもできます")
		util.Logger.Fatal(err)
	}

	// 全ユーザIDリスト 取得
	util.AllUserIDs = data.GetAllUserIDs(util.DeveloperMode)

	// マスタユーザID（labotGo ID）取得
	response, err := util.SocketModeClient.AuthTest()
	if err != nil {
		log.Fatal(err)
	}
	util.MasterUserID = response.UserID

	// 登録書籍サマリバッファ 初期化
	data.BookBuffer = map[string]map[string]data.BookSummary{}

	// Elasticsearch: クライアント生成
	if util.EsClient, err = elastic.NewClient(elastic.SetURL(util.EsURL), elastic.SetSniff(false)); err != nil {
		util.Logger.Fatal(err)
	}
	// Elasticsearch: バージョン取得
	if util.EsVersion, err = util.EsClient.ElasticsearchVersion(util.EsURL); err != nil {
		panic(err)
	}
	util.Logger.Printf("Elasticsearch バージョン %s\n", util.EsVersion)

	// 初回起動チェック（データファイル生成）
	if isFirstRun, err := checkFirstRun(); isFirstRun {
		if err != nil {
			util.Logger.Println("初回起動のため，データファイルの生成を試みましたが失敗しました")
			util.Logger.Println("詳しくは次のエラーを確認してください")
			util.Logger.Fatal(err)
		}
		util.Logger.Println("初回起動のため，以下の2つのデータファイルを生成しました")
		util.Logger.Printf("- メンバーデータ: %s\n", util.MemberDataPath())
		util.Logger.Printf("- チームデータ  : %s\n", util.TeamDataPath())
	} else {
		if err != nil {
			util.Logger.Println("データファイルの復元時に以下のエラーが発生しました")
			util.Logger.Fatal(err)
		}
	}
	util.Logger.Println("Tips: ボットが正常に動作しなくなる恐れがあるため，メンバー／チームデータは直接編集しないでください")

	// Elasticsearch: 書籍 index チェック
	if isBookIndex, err := es.InitializeIndex(util.EsBookIndexName, util.EsBookMappingPath()); !isBookIndex {
		if err != nil {
			util.Logger.Printf("index \"%s\" の作成を試みましたが失敗しました\n", util.EsBookIndexName)
			util.Logger.Println("詳しくは次のエラーを確認してください")
			util.Logger.Fatal(err)
		}
		util.Logger.Printf("%s index を作成しました\n", util.EsBookIndexName)
	}

	util.Logger.Printf("labotGo %s を起動しました\n", util.Version)

	// Socket Mode
	util.SocketModeClient.Run()
}

// 初回起動チェック
func checkFirstRun() (isFirstRun bool, err error) {
	// データファイル存在 チェック
	isMemberData, isTeamData := util.FileExists(util.MemberDataPath()), util.FileExists(util.TeamDataPath())
	if isMemberData && isTeamData {
		return isFirstRun, err
	}

	// 一方のデータファイルしか存在しない場合，もう一方のデータファイルを復元
	if isMemberData {
		if md, err := data.LoadMember(); err != nil {
			return isFirstRun, err
		} else {
			util.Logger.Println("チームデータファイルが存在しないため，メンバーデータから復元します")
			return isFirstRun, md.SynchronizeTeam()
		}
	}
	if isTeamData {
		if td, err := data.LoadTeam(); err != nil {
			return isFirstRun, err
		} else {
			util.Logger.Println("メンバーデータファイルが存在しないため，チームデータから復元します")
			return isFirstRun, td.SynchronizeMember()
		}
	}

	// 以下，初回起動処理
	// データファイル 生成
	isFirstRun = true
	if _, err := os.Create(util.MemberDataPath()); err != nil {
		return isFirstRun, err
	}
	if _, err := os.Create(util.TeamDataPath()); err != nil {
		return isFirstRun, err
	}

	// メンバーデータ 初期設定
	md := data.MembersData{}
	md.Add(util.MasterUserID, []string{util.MasterTeamName}, util.MasterUserID)
	if err := md.Reload(); err != nil {
		util.Logger.Fatal(err)
	}

	// チームデータ 初期設定
	td := data.TeamsData{}
	td.Add(util.MasterTeamName, []string{util.MasterUserID}, util.MasterUserID)
	if err := td.Reload(); err != nil {
		util.Logger.Fatal(err)
	}

	return isFirstRun, err
}
