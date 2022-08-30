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
