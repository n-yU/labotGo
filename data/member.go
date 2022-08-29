// データ管理
package data

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

// メンバー
type MemberData struct {
	TeamNames     []string  `yaml:"teams"`
	Image24       string    `yaml:"image_24"`
	Image32       string    `yaml:"image_32"`
	Image48       string    `yaml:"image_48"`
	CreatedUserID string    `yaml:"created_user"`
	CreatedAt     time.Time `yaml:"created_at"`
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

// メンバーデータエラー
func (md MembersData) GetErrBlocks(err error, dataErrType string) []slack.Block {
	var text string
	switch dataErrType {
	case util.DataLoadErr:
		text = "メンバーデータの読み込みに失敗しました"
	case util.DataReloadErr:
		text = "メンバーデータの更新に失敗しました"
	default:
		util.Logger.Fatalf("データエラータイプ %s は未定義です\n", dataErrType)
	}

	headerSection := post.SingleTextSectionBlock(util.PlainText, post.ErrText(text))
	tipsSection := post.TipsSection(post.TipsDataError(util.MemberDataPath()))
	blocks := []slack.Block{headerSection, tipsSection}

	util.Logger.Println(text)
	util.Logger.Println(err)
	return blocks
}

// メンバーデータによるチームデータの同期
func (md MembersData) SynchronizeTeam() error {
	if oldTd, err := LoadTeam(); err != nil {
		return err
	} else {
		newTd := TeamsData{}
		for userID, member := range md {
			for _, teamName := range member.TeamNames {
				if _, ok := newTd[teamName]; !ok {
					newTd[teamName] = NewTeam([]string{}, oldTd[teamName].CreatedUserID)
				}
				newTd[teamName].UserIDs = append(newTd[teamName].UserIDs, userID)
			}
		}

		err := newTd.Reload()
		return err
	}
}

// 新規メンバー
func NewMember(userID string, teamNames []string, createdUserID string, createdAt time.Time) (m *MemberData) {
	if prof, err := util.SocketModeClient.GetUserInfo(userID); err != nil {
		util.Logger.Printf("ユーザ \"%s\" のプロフィール取得に失敗しました", userID)
		util.Logger.Println(err)
	} else {
		m = &MemberData{
			TeamNames: teamNames, Image24: prof.Profile.Image24,
			Image32: prof.Profile.Image32, Image48: prof.Profile.Image48,
			CreatedUserID: createdUserID, CreatedAt: createdAt,
		}
	}
	return m
}

// メンバー追加
func (md MembersData) Add(userID string, teamNames []string, actionUserID string) {
	md[userID] = NewMember(userID, teamNames, actionUserID, time.Now())
}

// メンバー更新
func (md MembersData) Update(userID string, newTeamNames []string, actionUserID string) []string {
	oldTeamNames, createdUserID := md[userID].TeamNames, md[userID].CreatedUserID
	md[userID] = NewMember(userID, newTeamNames, createdUserID, md[userID].CreatedAt)
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
