// データ管理
package data

import (
	"fmt"
	"strings"

	"github.com/n-yU/labotGo/util"
)

// JPRO-onix 準拠項目
type BookOnix struct {
	CollateralDetail struct {
		TextContent []struct {
			Text string `json:"Text"`
		} `json:"TextContent"`
	} `json:"CollateralDetail"`
}

// 書籍サマリ
type BookSummary struct {
	ISBN       string `json:"isbn"`
	Title      string `json:"title"`
	Publisher  string `json:"publisher"`
	Pubdate    string `json:"pubdate"`
	Cover      string `json:"cover"`
	Authors    string `json:"author"`
	PubdateYMD string `json:"pubdateYMD"`
	Content    string `json:"content"`
}

// 書籍
type Book struct {
	BookOnix    `json:"onix"`
	BookSummary `json:"summary"`
}

// 書籍リスト
type Books []*Book

// 登録書籍サマリ バッファ
var BookBuffer map[string]map[string]BookSummary

// yyyy-MM-dd 形式 書籍出版日 設定
func (b *BookSummary) SetPubdateYMD() {
	if len(b.Pubdate) == 8 {
		b.PubdateYMD = strings.Join([]string{b.Pubdate[:4], b.Pubdate[4:6], b.Pubdate[6:]}, "-")
	} else {
		b.PubdateYMD = fmt.Sprintf("%s-01", b.Pubdate)
	}
}

// 書籍内容（タイトル+出版社+著者+内容紹介）設定
func (b *Book) SetContent() {
	var contentList []string

	textContent := b.BookOnix.CollateralDetail.TextContent
	if len(textContent) > 0 {
		for _, tc := range textContent {
			contentList = append(contentList, strings.Replace(tc.Text, "\n", " ", -1))
		}
	}

	b.BookSummary.Content = fmt.Sprintf(
		"%s %s %s %s", b.BookSummary.Title, b.BookSummary.Publisher,
		b.BookSummary.Authors, strings.Join(contentList, " "),
	)
}

// DB: 書籍テーブル 追加
func (b *BookSummary) AddDB() error {
	stmt, err := util.DB.Prepare("insert into books (ISBN, title, owner) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.ISBN, b.Title, util.DefaultBookOwner())
	return err
}

// DB: 書籍オーナー 取得
func (b *BookSummary) GetOwner() (owner string) {
	stmt, _ := util.DB.Prepare("select owner from books where ISBN = ?")
	defer stmt.Close()

	stmt.QueryRow(b.ISBN).Scan(&owner)
	return owner
}

// DB: 書籍オーナー 変更
func (b *BookSummary) ChangeOwner(newOwner string) (err error) {
	stmt, _ := util.DB.Prepare("update books set owner = ? where ISBN = ?")
	defer stmt.Close()

	_, err = stmt.Exec(newOwner, b.ISBN)
	return err
}

// DB: オーナー書籍 取得
func GetOwnerBooks(owner string) (ISBNs []string) {
	stmt, _ := util.DB.Prepare("select ISBN from books where owner = ?")
	defer stmt.Close()

	rows, _ := stmt.Query(owner)
	defer rows.Close()

	for rows.Next() {
		var isbn string
		rows.Scan(&isbn)
		ISBNs = append(ISBNs, isbn)
	}
	return ISBNs
}
