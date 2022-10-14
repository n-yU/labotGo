// データ管理
package data

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/n-yU/labotGo/util"
)

// チーム
type TeamData struct {
	UserIDs []string     `yaml:"members"`
	Created *CreatedInfo `yaml:"created"`
}

// チームリスト
type TeamsData map[string]*TeamData

// チームデータ 読み込み
func ReadTeam() (td TeamsData, err error) {
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

// チームデータ 書き込み
func (td TeamsData) Write() (err error) {
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

// 複合チームメンバー 取得
func (td TeamsData) GetComplexTeamMemberUserIDs(teamNames []string) (uniqueMemberUserIDs []string, err error) {
	var memberUserIDs []string
	for _, teamName := range teamNames {
		if team, ok := td[teamName]; ok {
			for _, userID := range team.UserIDs {
				memberUserIDs = append(memberUserIDs, userID)
			}
		} else {
			err = errors.New(fmt.Sprintf("指定したチーム `%s` は存在しません", teamName))
		}
	}
	uniqueMemberUserIDs = util.UniqueSlice(memberUserIDs)
	return uniqueMemberUserIDs, err
}

// チームデータによるメンバーデータの同期
func (td TeamsData) SynchronizeMember() error {
	if oldMd, err := ReadMember(); err != nil {
		return err
	} else {
		newMd := MembersData{}
		for teamName, team := range td {
			for _, userID := range team.UserIDs {
				if _, ok := newMd[userID]; !ok {
					newMd[userID] = NewMember(
						userID, []string{}, oldMd[userID].Created,
					)
				}
				newMd[userID].TeamNames = append(newMd[userID].TeamNames, teamName)
			}
		}

		err = newMd.Write()
		return err
	}
}

// 新規チーム
func NewTeam(teamUserIDs []string, created *CreatedInfo) *TeamData {
	t := &TeamData{
		UserIDs: teamUserIDs, Created: NewCreatedInfo(created.UserID, created.At),
	}
	return t
}

// チーム追加
func (td TeamsData) Add(teamName string, teamUserIDs []string, actionUserID string) {
	td[teamName] = NewTeam(teamUserIDs, NewCreatedInfo(actionUserID, time.Now()))
}

// チーム更新
func (td TeamsData) Update(oldTeamName, newTeamName string, newUserIDs []string, actionUserID string) []string {
	oldUserIDs, createdUserID, createdAt := td[oldTeamName].UserIDs, td[oldTeamName].Created.UserID, td[oldTeamName].Created.At
	delete(td, oldTeamName)
	td[newTeamName] = NewTeam(newUserIDs, NewCreatedInfo(createdUserID, createdAt))
	return oldUserIDs
}

// チーム削除
func (td TeamsData) Delete(teamName string) {
	delete(td, teamName)
}
