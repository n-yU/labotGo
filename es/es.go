// Elasticsearch 関連
package es

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/n-yU/labotGo/util"
)

// index: 初期化
func InitializeIndex(indexName string, mappingPath string) (isBookIndex bool, err error) {
	// index 存在確認
	if isBookIndex, err = IsExistIndex(indexName); err != nil {
		return isBookIndex, err
	} else if !isBookIndex {
		// index の存在が確認できない場合は作成
		if err = CreateIndex(indexName, mappingPath); err != nil {
			return isBookIndex, err
		}
	} else {
		util.Logger.Printf("index \"%s\" の存在を確認しました\n", indexName)
	}
	return isBookIndex, err
}

// index: 存在有無
func IsExistIndex(indexName string) (bool, error) {
	isExist, err := util.EsClient.IndexExists(indexName).Do(context.Background())
	return isExist, err
}

// index: 作成
func CreateIndex(indexName string, mappingPath string) (err error) {
	f, err := os.Open(mappingPath)
	if err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	mapping := string(bs)

	if createIndex, err := util.EsClient.CreateIndex(indexName).Body(
		mapping).IncludeTypeName(true).Do(context.Background()); err != nil {
		util.Logger.Printf("index \"%s\" の作成に失敗しました\n", indexName)
		return err
	} else if !createIndex.Acknowledged {
		util.Logger.Printf("create %s index - Not acknowledged\n", indexName)
	}

	return err
}
