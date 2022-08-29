// データ管理
package data

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
)

// チーム
type TeamData struct {
	UserIDs       []string  `yaml:"members"`
	CreatedUserID string    `yaml:"created_user"`
	CreatedAt     time.Time `yaml:"created_at"`
}

// チームリスト
type TeamsData map[string]*TeamData

// チームデータ 読み込み
func LoadTeam() (td TeamsData, err error) {
	f, err := os.Open(util.TeamDataPath())
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
func (td TeamsData) Reload() (err error) {
	bs, err := yaml.Marshal(&td)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.TeamDataPath(), bs, os.ModePerm)
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
		if t != util.MasterTeamName {
			teamName = append(teamName, t)
		}
	}
	return teamName
}

// チームデータエラー
func (td TeamsData) GetErrBlocks(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case util.DataLoadErr:
		text = "チームデータの読み込みに失敗しました"
	case util.DataReloadErr:
		text = "チームデータの更新に失敗しました"
	default:
		util.Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := post.SingleTextSectionBlock(util.PlainText, post.ErrText(text))
	tipsSection := post.TipsSection(post.TipsDataError(util.TeamDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	util.Logger.Println(text)
	util.Logger.Println(err)
	return blocks
}

// チームデータによるメンバーデータの同期
func (td TeamsData) SynchronizeMember() error {
	if oldMd, err := LoadMember(); err != nil {
		return err
	} else {
		newMd := MembersData{}
		for teamName, team := range td {
			for _, userID := range team.UserIDs {
				if _, ok := newMd[userID]; !ok {
					newMd[userID] = NewMember(
						userID, []string{}, oldMd[userID].CreatedUserID, oldMd[userID].CreatedAt,
					)
				}
				newMd[userID].TeamNames = append(newMd[userID].TeamNames, teamName)
			}
		}

		err = newMd.Reload()
		return err
	}
}

// 新規チーム
func NewTeam(teamUserIDs []string, createdUserID string) *TeamData {
	t := &TeamData{
		UserIDs: teamUserIDs, CreatedUserID: createdUserID, CreatedAt: time.Now(),
	}
	return t
}

// チーム追加
func (td TeamsData) Add(teamName string, teamUserIDs []string, actionUserID string) {
	td[teamName] = NewTeam(teamUserIDs, actionUserID)
}

// チーム更新
func (td TeamsData) Update(oldTeamName, newTeamName string, newUserIDs []string, actionUserID string) []string {
	oldUserIDs, createdUserID := td[oldTeamName].UserIDs, td[oldTeamName].CreatedUserID
	delete(td, oldTeamName)
	td[newTeamName] = NewTeam(newUserIDs, createdUserID)
	return oldUserIDs
}

// チーム削除
func (td TeamsData) Delete(teamName string) {
	delete(td, teamName)
}
