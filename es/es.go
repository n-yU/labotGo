// Elasticsearch 関連
package es

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
)

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
	} else {
	}

	util.Logger.Printf("index \"%s\" の作成に成功しました\n", indexName)
	return err
}

// index: 削除
func DeleteIndex(indexName string) (err error) {
	ctx := context.Background()
	if deleteIndex, err := util.EsClient.DeleteIndex(indexName).Do(ctx); err != nil {
		util.Logger.Printf("index \"%s\" の削除に失敗しました\n", indexName)
		return err
	} else if !deleteIndex.Acknowledged {
		util.Logger.Printf("delete %s index - Not acknowledged\n", indexName)
	}

	util.Logger.Printf("index \"%s\" の削除に成功しました\n", indexName)
	return err
}

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

// index: リセット
func ResetIndex(indexName, mappingPath string) (err error) {
	// index 削除
	err = DeleteIndex(indexName)
	if err == nil {
		err = CreateIndex(indexName, mappingPath)
	}

	if err == nil {
		util.Logger.Printf("index \"%s\" のリセットに成功しました\n", indexName)
	} else {
		util.Logger.Printf("index \"%s\" のリセットに失敗しました\n", indexName)
		util.Logger.Println(util.ReferErrorDetail)
		util.Logger.Println(err)
	}
	return err
}

// index: 存在有無
func IsExistIndex(indexName string) (bool, error) {
	isExist, err := util.EsClient.IndexExists(indexName).Do(context.Background())
	return isExist, err
}

// index: 追加
func PutIndex(body interface{}) (err error) {
	var indexName string

	// body の型から index 決定
	switch body.(type) {
	case data.BookSummary:
		// 書籍サマリ -> book
		bookSummary := body.(data.BookSummary)
		indexName = util.EsBookIndexName
		util.Logger.Printf("index \"%s\" への document 追加を試みます（ISBN: %s）\n", indexName, bookSummary.ISBN)

		doc, err := util.EsClient.Index().Index(indexName).Type("doc").Id(bookSummary.ISBN).BodyJson(bookSummary).Do(context.Background())
		if err != nil {
			util.Logger.Printf("index \"%s\" への document 追加に失敗しました\n", indexName)
			util.Logger.Println(util.ReferErrorDetail)
			return err
		}
		util.Logger.Printf("index \"%s\" への document 追加に成功しました（ID: %s）\n", doc.Index, doc.Id)
	default:
		// 未定義構造体
		err = errors.New(fmt.Sprintf("%T 型の body はインデックスへの追加に対応していません", body))
		return err
	}

	return err
}

// index: 登録 document カウント
func CountDoc(indexName string) int64 {
	count, _ := util.EsClient.Count(util.EsBookIndexName).Do(context.Background())
	return count
}
