// チーム管理
package team

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
		text := post.ErrText(fmt.Sprintf("コマンド %s team *%s* を使用することはできません\n", Cmd, subType))
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

// チームデータ 読み込み
func LoadData() (data map[string][]string, err error) {
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
func UpdateData(data map[string][]string) (err error) {
	bs, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(TeamDataPath, bs, os.ModePerm)
	return err
}
