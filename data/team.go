// データ管理
package data

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// チームデータ 読み込み
func LoadTeam() (data map[string][]string, err error) {
	f, err := os.Open(TeamDataPath)
	if err != nil {
		return data, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return data, err
	}

	err = yaml.Unmarshal([]byte(bs), &data)
	return data, err
}

// チームデータ 更新
func UpdateTeam(data map[string][]string) (err error) {
	bs, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(TeamDataPath, bs, os.ModePerm)
	return err
}

// 全チームリスト 取得
func GetAllTeams(data map[string][]string) []string {
	teams := []string{}
	for t := range data {
		teams = append(teams, t)
	}
	return teams
}

// Block Kit: チームデータエラー
func GetTeamErrBlocks(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case DataLoadErr:
		text = "チームデータの読み込みに失敗しました"
	case DataUpdateErr:
		text = "チームデータの更新に失敗しました"
	default:
		Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := post.SingleTextSectionBlock(PlainText, post.ErrText(text))
	tipsSection := post.TipsSection(post.TipsDataError(TeamDataPath))
	blocks := []slack.Block{headerSection, tipsSection}

	Logger.Println(text)
	Logger.Println(err)
	return blocks
}

// チームデータによるメンバーデータの同期
func SynchronizeMember(teamData map[string][]string) (err error) {
	memberData := map[string][]string{}
	for teamName, members := range teamData {
		for _, userID := range members {
			if _, ok := memberData[userID]; !ok {
				memberData[userID] = []string{}
			}
			memberData[userID] = append(memberData[userID], teamName)
		}
	}

	err = UpdateMember(memberData)
	return err
}
