// データ管理
package data

import (
	"io/ioutil"
	"os"

	"github.com/n-yU/labotGo/util"
	"gopkg.in/yaml.v3"
)

// 使用例
type ExCmd struct {
	Query string `yaml:"query"`
	Desc  string `yaml:"desc"`
}

// サブコマンド
type SubCmd struct {
	Name string  `yaml:"name"`
	Desc string  `yaml:"desc"`
	Ex   []ExCmd `yaml:"ex"`
}

// ヘルプ
type HelpData struct {
	Desc string   `yaml:"desc"`
	Sub  []SubCmd `yaml:"sub"`
	Ex   []ExCmd  `yaml:"ex"`
}

// ヘルプデータリスト
type HelpsData map[string]*HelpData

// ヘルプデータ 読み込み
func ReadHelps() (hd HelpsData, err error) {
	f, err := os.Open(util.HelpDataPath())
	if err != nil {
		return hd, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return hd, err
	}

	if len(bs) > 0 {
		err = yaml.Unmarshal([]byte(bs), &hd)
	} else {
		util.Logger.Fatalf("ヘルプデータ \"%s\" が存在しません\n", util.HelpDataPath())
	}

	return hd, err
}
