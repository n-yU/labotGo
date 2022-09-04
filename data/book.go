// データ管理
package data

import (
	"fmt"
	"strings"
)

// 商品付帯項目
type TextContent []struct {
	Text            string `json:"Text"`
	TextType        string `json:"TextType"`
	ContentAudience string `json:"ContentAudience"`
}

// JPRO-onix準拠項目
type BookOnix struct {
	CollateralDetail struct {
		TextContent `json:"TextContent"`
	} `json:"CollateralDetail"`
}

// 書籍サマリ
type BookSummary struct {
	Title      string `json:"title"`
	ISBN       string `json:"isbn"`
	Publisher  string `json:"publisher"`
	Pubdate    string `json:"pubdate"`
	Cover      string `json:"cover"`
	Authors    string `json:"author"`
	PubdateYMD string
	Content    string
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
	if textContent != nil {
		for _, tc := range textContent {
			contentList = append(contentList, strings.Replace(tc.Text, "\n", " ", -1))
		}
	}

	b.BookSummary.Content = fmt.Sprintf(
		"%s %s %s %s", b.BookSummary.Title, b.BookSummary.Publisher,
		b.BookSummary.Authors, strings.Join(contentList, " "),
	)
}
