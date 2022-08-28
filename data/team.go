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

// チーム
type TeamData struct {
	UserIDs []string `yaml:"members"`
}

// チームリスト
type TeamsData map[string]*TeamData

// チームデータ 読み込み
func LoadTeam() (td TeamsData, err error) {
	f, err := os.Open(TeamDataPath())
	if err != nil {
		return td, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return td, err
	}

	err = yaml.Unmarshal([]byte(bs), &td)
	return td, err
}

// チームデータ 更新
func (td TeamsData) Update() (err error) {
	bs, err := yaml.Marshal(&td)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(TeamDataPath(), bs, os.ModePerm)
	return err
}

// 全チームリスト 取得
func (td TeamsData) GetAllNames() (teamNames []string) {
	for t := range td {
		teamNames = append(teamNames, t)
	}
	return teamNames
}

// 全編集可能チームリスト 取得
func (td TeamsData) GetAllEditedNames() (teamName []string) {
	for _, t := range td.GetAllNames() {
		if t != MasterTeamName {
			teamName = append(teamName, t)
		}
	}
	return teamName
}

// チームデータエラー
func (td TeamsData) GetErrBlocks(err error, dataErrType string) []slack.Block {
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
	tipsSection := post.TipsSection(post.TipsDataError(TeamDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	Logger.Println(text)
	Logger.Println(err)
	return blocks
}

// チームデータによるメンバーデータの同期
func (td TeamsData) SynchronizeMember() (err error) {
	md := MembersData{}
	for teamName, team := range td {
		for _, uID := range team.UserIDs {
			if _, ok := md[uID]; !ok {
				md[uID] = new(MemberData)
			}
			md[uID].TeamNames = append(md[uID].TeamNames, teamName)
		}
	}

	err = md.Update()
	return err
}
