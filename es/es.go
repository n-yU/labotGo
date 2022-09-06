// Elasticsearch 関連
package es

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/n-yU/labotGo/data"
	"github.com/n-yU/labotGo/util"
	"github.com/olivere/elastic/v7"
)

var (
	ErrDocAlreadyExist = errors.New("index に指定したIDの document が既に存在します")
	ErrDocNotFound     = errors.New("index にて指定したIDの document が見つかりませんでした")
)

// index: 作成
func CreateIndex(index string, mappingPath string) (err error) {
	f, err := os.Open(mappingPath)
	if err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	mapping := string(bs)

	if createIndex, err := util.EsClient.CreateIndex(index).Body(
		mapping).IncludeTypeName(true).Do(context.Background()); err != nil {
		util.Logger.Printf("index \"%s\" の作成に失敗しました\n", index)
		return err
	} else if !createIndex.Acknowledged {
		util.Logger.Printf("create %s index - Not acknowledged\n", index)
	} else {
	}

	util.Logger.Printf("index \"%s\" の作成に成功しました\n", index)
	return err
}

// index: 削除
func DeleteIndex(index string) (err error) {
	ctx := context.Background()
	if deleteIndex, err := util.EsClient.DeleteIndex(index).Do(ctx); err != nil {
		util.Logger.Printf("index \"%s\" の削除に失敗しました\n", index)
		return err
	} else if !deleteIndex.Acknowledged {
		util.Logger.Printf("delete %s index - Not acknowledged\n", index)
	}

	util.Logger.Printf("index \"%s\" の削除に成功しました\n", index)
	return err
}

// index: 初期化
func InitializeIndex(index string, mappingPath string) (isBookIndex bool, err error) {
	// index 存在確認
	if isBookIndex, err = IsExistIndex(index); err != nil {
		return isBookIndex, err
	} else if !isBookIndex {
		// index の存在が確認できない場合は作成
		if err = CreateIndex(index, mappingPath); err != nil {
			return isBookIndex, err
		}
	} else {
		util.Logger.Printf("index \"%s\" の存在を確認しました\n", index)
	}
	return isBookIndex, err
}

// index: リセット
func ResetIndex(index, mappingPath string) (err error) {
	// index 削除
	err = DeleteIndex(index)
	if err == nil {
		err = CreateIndex(index, mappingPath)
	}

	if err == nil {
		util.Logger.Printf("index \"%s\" のリセットに成功しました\n", index)
	} else {
		util.Logger.Printf("index \"%s\" のリセットに失敗しました\n", index)
		util.Logger.Println(util.ReferErrorDetail)
		util.Logger.Println(err)
	}
	return err
}

// index: 存在有無
func IsExistIndex(index string) (bool, error) {
	isExist, err := util.EsClient.IndexExists(index).Do(context.Background())
	return isExist, err
}

// index: 登録 document カウント
func CountDoc(index string) int64 {
	count, _ := util.EsClient.Count(util.EsBookIndex).Do(context.Background())
	return count
}

// document: 存在有無
func IsExistDoc(index, type_, id string) bool {
	ctx := context.Background()
	_, err := util.EsClient.Get().Index(index).Type(type_).Id(id).Do(ctx)
	isExist := !elastic.IsNotFound(err)
	return isExist
}

// document: 追加
func PutDocument(body interface{}) (err error) {
	var index, type_ string
	ctx := context.Background()

	// body の型から index,type 決定
	switch body.(type) {
	case data.BookSummary:
		// 書籍サマリ -> book
		bookSummary := body.(data.BookSummary)
		index, type_ = util.EsBookIndex, util.EsBookType
		util.Logger.Printf("index \"%s\" - type \"%s\" への document 追加を試みます（ISBN: %s）\n", index, type_, bookSummary.ISBN)

		// 書籍存在確認
		if IsExistDoc(index, type_, bookSummary.ISBN) {
			err = ErrDocAlreadyExist
			util.Logger.Println(err)
			return err
		}

		// 書籍登録
		doc, err := util.EsClient.Index().Index(index).Type(util.EsBookType).Id(bookSummary.ISBN).BodyJson(bookSummary).Do(ctx)
		if err != nil {
			util.Logger.Printf("index \"%s\" - type \"%s\" への document 追加に失敗しました\n", index, type_)
			util.Logger.Println(util.ReferErrorDetail)
			util.Logger.Println(err)
			return err
		}
		util.Logger.Printf("index \"%s\" - type \"%s\" への document 追加に成功しました（ID: %s）\n", doc.Index, doc.Type, doc.Id)
	default:
		// 未定義構造体
		err = errors.New(fmt.Sprintf("%T 型の body はインデックスへの追加に対応していません", body))
		return err
	}

	return err
}

// document: 削除
func DeleteDocument(body interface{}) (err error) {
	var index, type_ string
	ctx := context.Background()

	// body の型から index,type 決定
	switch body.(type) {
	case data.BookSummary:
		// 書籍サマリ -> book
		bookSummary := body.(data.BookSummary)
		index, type_ = util.EsBookIndex, util.EsBookType
		util.Logger.Printf("index \"%s\" - type \"%s\" への document 削除を試みます（ISBN: %s）\n", index, type_, bookSummary.ISBN)

		// 書籍存在確認
		if !IsExistDoc(index, type_, bookSummary.ISBN) {
			err = ErrDocNotFound
			util.Logger.Println(err)
			return err
		}

		// 書籍削除
		doc, err := util.EsClient.Delete().Index(index).Type(type_).Id(bookSummary.ISBN).Do(ctx)
		if err != nil {
			util.Logger.Printf("index \"%s\" - type \"%s\" への document 削除に失敗しました\n", index, type_)
			util.Logger.Println(util.ReferErrorDetail)
			util.Logger.Println(err)
			return err
		}
		util.Logger.Printf("index \"%s\" - type \"%s\" への document 削除に成功しました（ID: %s）\n", doc.Index, doc.Type, doc.Id)
	default:
		// 未定義構造体
		err = errors.New(fmt.Sprintf("%T 型の body はインデックスへの追加に対応していません", body))
		return err
	}

	return err
}

// document: 検索
func SearchDocument(body interface{}, query string, n int) (results interface{}, err error) {
	var (
		index        string
		searchResult *elastic.SearchResult
	)
	ctx := context.Background()

	// body の型から index 決定
	switch body.(type) {
	case data.BookSummary:
		// 書籍サマリ -> book
		var bookSummaryResults []data.BookSummary
		index = util.EsBookIndex
		util.Logger.Printf("index \"%s\" 内の document 検索を試みます\n", index)

		// 書籍検索
		q := elastic.NewMatchQuery("content", query)
		searchResult, err = util.EsClient.Search().Index(index).Query(q).From(0).Size(n).Do(ctx)
		if err != nil {
			util.Logger.Printf("index \"%s\" 内の document 検索に失敗しました\n", index)
			util.Logger.Println(util.ReferErrorDetail)
			util.Logger.Println(err)
			return bookSummaryResults, err
		}

		// 検索結果を書籍サマリ形式で格納
		for _, doc := range searchResult.Each(reflect.TypeOf(body)) {
			bookSummaryResults = append(bookSummaryResults, doc.(data.BookSummary))
		}
		results = bookSummaryResults
	default:
		// 未定義構造体
		err = errors.New(fmt.Sprintf("%T 型の body はインデックス内検索に対応していません", body))
		return results, err
	}

	util.Logger.Printf("ヒット数: %d 件（検索実行時間: %d ms）\n", searchResult.TotalHits(), searchResult.TookInMillis)
	return results, err
}
