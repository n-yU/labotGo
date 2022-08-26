// データ管理
package data

import (
	"io/ioutil"
	"os"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

// メンバーデータ 読み込み
func LoadMember() (data map[string][]string, err error) {
	f, err := os.Open(MemberDataPath)
	if err != nil {
		return data, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return data, err
	}

	if len(bs) > 0 {
		err = yaml.Unmarshal([]byte(bs), &data)
	} else {
		data = map[string][]string{}
	}

	return data, err
}

// メンバーデータ 更新
func UpdateMember(data map[string][]string) (err error) {
	bs, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(MemberDataPath, bs, os.ModePerm)
	return err
}

// 全メンバーリスト 取得
func GetAllMembers(data map[string][]string) []string {
	members := []string{}
	for userId := range data {
		members = append(members, userId)
	}
	return members
}

// Block Kit: メンバーデータエラー
func GetMemberErrBlocks(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case DataLoadErr:
		text = "メンバーデータの読み込みに失敗しました"
	case DataUpdateErr:
		text = "メンバーデータの更新に失敗しました"
	default:
		Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := post.SingleTextSectionBlock(PlainText, post.ErrText(text))
	tipsSection := post.TipsSection(post.TipsDataError(MemberDataPath))
	blocks := []slack.Block{headerSection, tipsSection}

	Logger.Println(text)
	Logger.Println(err)
	return blocks
}

// メンバーデータによるチームデータの同期
func SynchronizeTeam(memberData map[string][]string) (err error) {
	teamData := map[string][]string{}
	for userID, teams := range memberData {
		for _, teamName := range teams {
			if _, ok := teamData[teamName]; !ok {
				teamData[teamName] = []string{}
			}
			teamData[teamName] = append(teamData[teamName], userID)
		}
	}

	err = UpdateTeam(teamData)
	return err
}
