// ユーティリティ・変数・定数
package util

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	Version   = "1.0.0-alpha"
	DebugMode = false
	Cmd       = "/lab"

	InChannel = slack.ResponseTypeInChannel
	Ephemeral = slack.ResponseTypeEphemeral
	Markdown  = slack.MarkdownType
	PlainText = slack.PlainTextType

	MemberDataPath = "/go/src/app/member/data.yml"
	TeamDataPath   = "/go/src/app/team/data.yml"

	TipsMemberTeam = "*labotGo* に追加されたユーザを `メンバー` とし， `メンバー` のグループを `チーム` と呼びます"
)

var (
	Logger           *log.Logger
	Api              *slack.Client
	SocketModeClient *socketmode.Client
)

// 環境変数 読み込み
func LoadEnv() {
	// ref.) https://zenn.dev/a_ichi1/articles/c9f3870350c5e2
	if err := godotenv.Load(".env"); err != nil {
		Logger.Println("環境変数の読み込みに失敗しました")
		Logger.Fatal(err)
	}
}

// ファイル存在 チェック
func FileExists(filename string) bool {
	// ref.) https://qiita.com/suin/items/b9c0f92851454dc6d461
	_, err := os.Stat(filename)
	return err == nil
}
