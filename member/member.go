// メンバー管理
package member

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/n-yU/labotGo/post"
	. "github.com/n-yU/labotGo/util"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

// コマンド応答ブロック 取得
func GetBlocks(cmdValues []string) (blocks []slack.Block, responseType string, ok bool) {
	switch subType := cmdValues[0]; subType {
	case "add":
		blocks, responseType, ok = getBlockAdd(), Ephemeral, true
	case "edit":

	case "delete":

	case "list":

	default:
		text := fmt.Sprintf("コマンド %s member *%s* を使用することはできません\n", Cmd, subType)
		blocks, responseType, ok = post.CreateSingleTextBlock(text), Ephemeral, true
	}

	return blocks, responseType, ok
}

// 指定アクション 実行
func Action(actionId string, callback slack.InteractionCallback) (err error) {
	switch {
	case strings.HasSuffix(actionId, "Add"):
		blocks := AddMember(callback.BlockActionState.Values)
		err = post.PostMessage(callback, blocks, Ephemeral)
	}

	return err
}

// メンバーデータ 読み込み
func LoadData() (data map[string]interface{}, err error) {
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
		data = map[string]interface{}{}
	}

	return data, err
}

// メンバーデータ 更新
func UpdateData(data map[string]interface{}) (err error) {
	bs, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(MemberDataPath, bs, os.ModePerm)
	return err
}
