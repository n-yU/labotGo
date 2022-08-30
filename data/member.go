// データ管理
package data

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/n-yU/labotGo/util"
	"gopkg.in/yaml.v3"
)

// メンバー
type MemberData struct {
	TeamNames []string     `yaml:"teams"`
	Image24   string       `yaml:"image_24"`
	Image32   string       `yaml:"image_32"`
	Image48   string       `yaml:"image_48"`
	Created   *CreatedInfo `yaml:"created"`
}

// メンバーリスト
type MembersData map[string]*MemberData

// メンバーデータ 読み込み
func LoadMember() (md MembersData, err error) {
	f, err := os.Open(util.MemberDataPath())
	if err != nil {
		return md, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return md, err
	}

	if len(bs) > 0 {
		err = yaml.Unmarshal([]byte(bs), &md)
	} else {
		util.Logger.Fatalf("メンバーデータ \"%s\"が存在しません\n", util.MemberDataPath())
	}

	return md, err
}

// メンバーデータ 更新
func (md MembersData) Reload() (err error) {
	bs, err := yaml.Marshal(&md)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.MemberDataPath(), bs, os.ModePerm)
	return err
}

// 全メンバーリスト 取得
func (md MembersData) GetAllUserIDs() (userIDs []string) {
	for uID := range md {
		userIDs = append(userIDs, uID)
	}
	return userIDs
}

// 全編集可能メンバーリスト 取得
func (md MembersData) GetAllEditedUserIDs() (userIDs []string) {
	for _, uID := range md.GetAllUserIDs() {
		if uID != util.MasterUserID {
			userIDs = append(userIDs, uID)
		}
	}
	return userIDs
}

// メンバーデータによるチームデータの同期
func (md MembersData) SynchronizeTeam() error {
	if oldTd, err := LoadTeam(); err != nil {
		return err
	} else {
		// 所属メンバーを持たないチームが存在するため，更新前のチームデータから各チームデータを予めコピー
		newTd := TeamsData{}
		for _, teamName := range oldTd.GetAllNames() {
			newTd[teamName] = NewTeam([]string{}, oldTd[teamName].Created)
		}
		for userID, member := range md {
			for _, teamName := range member.TeamNames {
				newTd[teamName].UserIDs = append(newTd[teamName].UserIDs, userID)
			}
		}

		err := newTd.Reload()
		return err
	}
}

// 新規メンバー
func NewMember(userID string, teamNames []string, created *CreatedInfo) (m *MemberData) {
	if prof, err := util.SocketModeClient.GetUserInfo(userID); err != nil {
		util.Logger.Printf("ユーザ \"%s\" のプロフィール取得に失敗しました", userID)
		util.Logger.Println(err)
	} else {
		m = &MemberData{
			TeamNames: teamNames, Image24: prof.Profile.Image24,
			Image32: prof.Profile.Image32, Image48: prof.Profile.Image48,
			Created: NewCreatedInfo(created.UserID, created.At),
		}
	}
	return m
}

// メンバー追加
func (md MembersData) Add(userID string, teamNames []string, actionUserID string) {
	md[userID] = NewMember(userID, teamNames, NewCreatedInfo(actionUserID, time.Now()))
}

// メンバー更新
func (md MembersData) Update(userID string, newTeamNames []string, actionUserID string) []string {
	oldTeamNames, createdUserID := md[userID].TeamNames, md[userID].Created.UserID
	md[userID] = NewMember(userID, newTeamNames, NewCreatedInfo(createdUserID, md[userID].Created.At))
	return oldTeamNames
}

// メンバー削除
func (md MembersData) Delete(userID string) {
	delete(md, userID)
}

// 指定メンバー プロフィール画像取得
func (md MembersData) GetProfImages(userIDs []string) map[string]string {
	profImages := map[string]string{}
	for _, uID := range userIDs {
		profImages[uID] = md[uID].Image24
	}
	return profImages
}
