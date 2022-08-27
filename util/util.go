// ユーティリティ・変数・定数
package util

import (
	"log"
	"os"
	"reflect"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	Version       = "1.0.0-alpha"
	DebugMode     = false
	DeveloperMode = true // botもメンバー追加を許可（デバッグ用）
	Cmd           = "/lab"

	InChannel = slack.ResponseTypeInChannel
	Ephemeral = slack.ResponseTypeEphemeral
	Markdown  = slack.MarkdownType
	PlainText = slack.PlainTextType

	MemberDataPath = "/go/src/app/data/member.yml"
	TeamDataPath   = "/go/src/app/data/team.yml"

	TipsMemberTeam       = "*labotGo* に追加されたユーザを `メンバー` とし， `メンバー` のグループを `チーム` と呼びます"
	TipsTeamALL          = "チーム `all` は全メンバーが入るチームです．削除しないでください．"
	ErrorSynchronizeData = "メンバーデータとチームデータの同期に失敗しました"

	DataLoadErr   = "dataLoadError"
	DataUpdateErr = "dataUpdateError"
)

var (
	Logger           *log.Logger
	Api              *slack.Client
	SocketModeClient *socketmode.Client
	AllUserIDs       []string
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

// 区切り線ブロック
func Divider() *slack.DividerBlock {
	return slack.NewDividerBlock()
}

// スライス・配列に特定の値が含まれるか判定
func ListContains(list interface{}, elem interface{}) bool {
	// ref.) https://zenn.dev/glassonion1/articles/7c7830a269909c
	listV := reflect.ValueOf(list)

	if listV.Kind() == reflect.Slice {
		for i := 0; i < listV.Len(); i++ {
			item := listV.Index(i).Interface()
			if !reflect.TypeOf(elem).ConvertibleTo(reflect.TypeOf(item)) {
				continue
			}
			target := reflect.ValueOf(elem).Convert(reflect.TypeOf(item)).Interface()
			if ok := reflect.DeepEqual(item, target); ok {
				return true
			}
		}
	}
	return false
}

// 文字列スライスを連結して重複する要素を削除
func UniqueConcatSlice(slice1, slice2 []string) (unique []string) {
	// ref.) https://zenn.dev/orangekame/articles/dad6d0e9382660
	slice1 = append(slice1, slice2...)
	encounter := map[string]bool{}

	for _, v := range slice1 {
		if _, ok := encounter[v]; !ok {
			encounter[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}
