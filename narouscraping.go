package main

//小説家になろうを解析
//各話のサブタイトル一覧や、全文などを取得
//著者やあらすじなどの雑多な情報を取得するのはnovelinformationに任せる

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const narouURL string = "http://ncode.syosetu.com" //小説家になろうのURL

type narouNovel struct {
	ncode     string
	rubiStart string //置き換えられるルビのはじめ
	rubiEnd   string //置き換えられるルビの閉じ
}

//小説一話による情報
type storyInformation struct {
	number       int    //何話目
	subTitle     string //サブタイトル
	chapterTitle string //チャプター名
}

//小説構造体を作成
func newNarouNovel() *narouNovel {
	return &narouNovel{}
}

//stories2LinesArray 引数の小説の各話をLines配列に変換
func stories2LinesArray(stories []storyInformation) []Lines {
	linesArr := []Lines{}
	for _, s := range stories {
		lines := Lines{s.subTitle, s.chapterTitle, ""}
		linesArr = append(linesArr, lines)
	}
	return linesArr
}

//初期化
func (novel *narouNovel) init(ncode, rubiSt, rubiEnd string) {
	novel.ncode = ncode
	novel.rubiStart = rubiSt
	novel.rubiEnd = rubiEnd
}

//小説一覧情報を一気に取得
func (novel *narouNovel) getIndexByChapter() []storyInformation {
	doc, err := goquery.NewDocument(narouURL + "/" + novel.ncode + "/")
	if err != nil {
		fmt.Println(err)
	}
	chapterTitle := ""
	storyNumCon := 1                //何話目かのカウンター
	stories := []storyInformation{} //返す値

	//各話の情報を収集
	doc.Find(".index_box").Each(func(_ int, indexbox *goquery.Selection) {
		//各項目を処理
		indexbox.Find("div[class ='chapter_title'], dl[class = 'novel_sublist2']").Each(func(_ int, s *goquery.Selection) {
			attr, _ := s.Attr("class")
			switch attr {
			case "chapter_title":
				//チャプター名
				chapterTitle = s.Text()
			case "novel_sublist2":
				//小説各話
				story := storyInformation{
					storyNumCon,
					s.Find(".subtitle").Text(),
					chapterTitle,
				}
				stories = append(stories, story)
			default:
			}
		})
	})
	return stories
}

//getStory 小説を取得。戻り値は行分けされたString配列
func (novel *narouNovel) getStory(storyNum int) []string {
	doc, err := goquery.NewDocument(narouURL + "/" + novel.ncode + "/" + strconv.Itoa(storyNum) + "/")
	var honbunLinesArray []string
	if err != nil {
		fmt.Println(err)
	}
	//本文を取得
	doc.Find("div[id='novel_honbun']").Each(func(_ int, honbun *goquery.Selection) {
		//ルビに使われる()を置き換えている
		honbun.Find("rp").Each(func(_ int, rubi *goquery.Selection) {
			switch rubi.Text() {
			case "(":
				rubi = rubi.ReplaceWithHtml(novel.rubiStart)
			case ")":
				rubi = rubi.ReplaceWithHtml(novel.rubiEnd)
			default:
			}
		})
		honbunLinesArray = regexp.MustCompile("\r\n|\n\r|\n|\r").Split(honbun.Text(), -1)
	})
	return honbunLinesArray
}
