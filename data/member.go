// データ管理
package data

import (
	"io/ioutil"
	"os"

	"github.com/n-yU/labotGo/post"
	"github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

// メンバー
type MemberData struct {
	TeamNames []string `yaml:"teams"`
	Image     string   `yaml:"image"`
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
func (md MembersData) Update() (err error) {
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
	case util.DataUpdateErr:
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
func (md MembersData) SynchronizeTeam() (err error) {
	td := TeamsData{}
	for userID, member := range md {
		for _, teamName := range member.TeamNames {
			if _, ok := td[teamName]; !ok {
				td[teamName] = new(TeamData)
			}
			td[teamName].UserIDs = append(td[teamName].UserIDs, userID)
		}
	}

	err = td.Update()
	return err
}
