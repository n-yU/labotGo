// データ管理
package data

import (
	"github.com/n-yU/labotGo/util"
)

// 全ユーザIDリスト 取得
func GetAllUserIDs(isIncludeBot bool) (ids []string) {
	users, err := util.SocketModeClient.GetUsers()
	if err != nil {
		util.Logger.Println("ワークスペースの全ユーザIDリストの取得に失敗しました")
		util.Logger.Fatal(err)
	}

	for _, u := range users {
		if !isIncludeBot || !u.IsBot || u.ID != "USLACKBOT" {
			ids = append(ids, u.ID)
		}
	}

	util.Logger.Println("ワークスペースの全ユーザIDリストの取得に成功しました")
	return ids
}

// 指定ID削除ユーザIDリスト 取得
func GetLimitedUserIDs(excludeUserIDs []string) (ids []string) {
	for _, uID := range util.AllUserIDs {
		if !util.ListContains(excludeUserIDs, uID) {
			ids = append(ids, uID)
		}
	}
	return ids
}

// 非メンバー ユーザIDリスト 取得
func GetAllNonMembers(md MembersData) []string {
	return GetLimitedUserIDs(md.GetAllUserIDs())
}

// メンションフォーマットからのユーザID取得
func RawUserID(value string) string {
	// ref.) https://api.slack.com/reference/surfaces/formatting#mentioning-users
	return value[2 : len(value)-1]
}
