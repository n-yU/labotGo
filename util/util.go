// ユーティリティ・変数・定数
package util

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/joho/godotenv"
	"github.com/olivere/elastic/v7"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

const (
	Version        = "1.0.0-alpha"
	Cmd            = "/lab"
	MasterTeamName = "all"

	InChannel = slack.ResponseTypeInChannel
	Ephemeral = slack.ResponseTypeEphemeral
	Markdown  = slack.MarkdownType
	PlainText = slack.PlainTextType

	TipsMemberTeam       = "*labotGo* に追加されたユーザを `メンバー` とし， `メンバー` のグループを `チーム` と呼びます"
	TipsMasterTeam       = "チーム `all` は全メンバーが入るチームです．編集・削除などの各操作はできません．"
	ErrorSynchronizeData = "メンバーデータとチームデータの同期に失敗しました"
	ReferErrorDetail     = "詳しくは次のエラーを確認してください"

	DataReadErr  = "dataReadError"
	DataWriteErr = "dataWriteError"

	GroupTypeOptionNum  = "グループ数"
	GroupTypeOptionSize = "グループサイズ"

	OpenBD = "https://api.openbd.jp/v1/get"

	EsURL       = "http://elasticsearch:9200"
	EsBookIndex = "book"
	EsBookType  = "doc"

	MaxBlocks = 50
)

var (
	Debug            = false
	NoElastic        = false
	Logger           *log.Logger
	Api              *slack.Client
	SocketModeClient *socketmode.Client

	EsClient  *elastic.Client // Elasticsearch クライアント
	EsVersion string          // Elasticsearch バージョン

	Dir          string   // 実行ファイルパスディレクトリ
	AllUserIDs   []string // ワークスペース全ユーザID
	MasterUserID string   // マスターユーザID（labotGo ID）

	DB *sql.DB // データベース
)

// getter: メンバーデータパス
func MemberDataPath() string {
	return fmt.Sprintf("%s/data/member.yml", Dir)
}

// getter: チームデータパス
func TeamDataPath() string {
	return fmt.Sprintf("%s/data/team.yml", Dir)
}

// getter: マスターユーザ Tips
func TipsMasterUser() string {
	return fmt.Sprintf("ユーザ <@%s> はマスターメンバーです．編集・削除などの各操作はできません．", MasterUserID)
}

// getter: チームリスト Tips
func TipsTeamList() string {
	return fmt.Sprintf("チームの一覧を確認するには `%s team list` を実行してください", Cmd)
}

// getter: index:"book" マッピングパス
func EsBookMappingPath() string {
	return fmt.Sprintf("%s/es/mapping.json", Dir)
}

// getter: 環境変数パス
func EnvPath() string {
	return fmt.Sprintf("%s/.env", Dir)
}

// getter: DBパス
func DBPath() string {
	return fmt.Sprintf("%s/data/data.db", Dir)
}

// getter: デフォルト書籍オーナー
func DefaultBookOwner() string {
	return os.Getenv("DEFAULT_BOOK_OWNER")
}

// 環境変数 読み込み
func LoadEnv() {
	// ref.) https://zenn.dev/a_ichi1/articles/c9f3870350c5e2
	if err := godotenv.Load(EnvPath()); err != nil {
		Logger.Println("環境変数の読み込みに失敗しました")
		Logger.Fatal(err)
	}
}

// ファイル存在 チェック
func FileExists(filePath string) bool {
	// ref.) https://qiita.com/suin/items/b9c0f92851454dc6d461
	_, err := os.Stat(filePath)
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

// 文字列スライスから重複する要素を削除
func UniqueSlice(slice []string) (unique []string) {
	// ref.) https://zenn.dev/orangekame/articles/dad6d0e9382660
	encounter := map[string]bool{}
	for _, v := range slice {
		if _, ok := encounter[v]; !ok {
			encounter[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}

// 文字列スライスを連結して重複する要素を削除
func UniqueConcatSlice(slice1, slice2 []string) []string {
	return UniqueSlice(append(slice1, slice2...))
}

// 時刻フォーマット
func FormatTime(t time.Time) string {
	tz, _ := time.LoadLocation("Asia/Tokyo")
	return t.In(tz).Format("2006/01/02 15:04:05")
}

// リセットコード 設定有無
func IsSetResetCode() bool {
	return os.Getenv("RESET_CODE") != "rc"
}
