// データ管理
package data

import (
	"time"

	"github.com/n-yU/labotGo/util"
)

// データ作成情報
type CreatedInfo struct {
	UserID string    `yaml:"userID"`
	Image  string    `yaml:"image"`
	At     time.Time `yaml:"at"`
}

// テーブル作成
var (
	booksSqlStmt = `
create table if not exists books (
	ISBN integer not null primary key,
	title text not null,
	owner text not null,
	created_at text not null default (datetime('now', 'localtime')),
	updated_at text not null default (datetime('now', 'localtime'))
)` // 書籍
)

// 新規データ作成情報
func NewCreatedInfo(userID string, At time.Time) (ci *CreatedInfo) {
	if prof, err := util.SocketModeClient.GetUserInfo(userID); err != nil {
		util.Logger.Printf("ユーザ \"%s\" のプロフィール取得に失敗しました", userID)
		util.Logger.Println(err)
	} else {
		ci = &CreatedInfo{UserID: userID, Image: prof.Profile.Image24, At: At}
	}
	return ci
}

// テーブル作成
func CreateTables() (err error) {
	for _, sqlStmt := range []string{booksSqlStmt} {
		if _, err := util.DB.Exec(sqlStmt); err != nil {
			return err
		}
	}
	return err
}
