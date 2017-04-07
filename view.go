package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nsf/termbox-go"
)

var (
	defaultFg   termbox.Attribute //選択肢の文字
	defaultBg   termbox.Attribute //選択肢のバックグラウンド
	currentmode *viewer           //現在表示中の構造体

	topView                   *topview
	managementdlView          *managementdlview
	searchmenuView            *searchmenuview            //検索条件を指定
	searchmenufiltergenreView *searchmenufiltergenreview //検索ジャンルを指定
	searchresultView          *searchresultview
	noveltopView              *noveltopview
	novelviewerView           *novelview
)

//ScreenType 画面のタイプ
type ScreenType int

//各画面の名称
const (
	Top            ScreenType = iota
	ManagementOfDL ScreenType = iota
	SearchMenu     ScreenType = iota
	SearchResult   ScreenType = iota
	NovelTop       ScreenType = iota
	NovelView      ScreenType = iota
)

func (s ScreenType) String() string {
	switch s {
	case Top:
		return "Top"
	case ManagementOfDL:
		return "ManagementOfDL"
	case SearchMenu:
		return "SearchMenu"
	case SearchResult:
		return "SearchResult"
	case NovelTop:
		return "NovelTop"
	case NovelView:
		return "NovelView"
	default:
		return "UnKnown"
	}
}

//トップ画面構造体
type topview struct {
	downloading bool //ダウンロード中ならオン
}

//DL作品管理画面構造体
type managementdlview struct {
}

//検索画面構造体
type searchmenuview struct {
	searchFilter url.Values //検索条件を指定するクエリを保存、追加していく
	searchString string     //何についてを検索条件として指定するか記述して、表示する
}

//検索ジャンル指定画面構造体
type searchmenufiltergenreview struct {
	searchFilter url.Values //検索条件を指定するクエリを保存追加する
	searchString string
}

//検索結果画面構造体
type searchresultview struct {
	searchFilter url.Values                 //なろうAPIの検索文字列
	searchString string                     //検索条件について記述した文字列
	resultList   []narouAPISearchResultjson //検索結果
	updateResult bool                       //trueの時情報を更新
}

//小説トップ画面構造体
type noveltopview struct {
	ncode        string //入手するNCode
	title        string
	novelInfo    *novelinformation //表示する小説の情報
	novelStories *narouNovel
	storiesIndex []storyInformation
}

//小説表示画面構造体
type novelview struct {
	novelInfo    *novelinformation
	novelStories *narouNovel
	storyInfo    *storyInformation
	ncode        string
	currentnum   int //現在話数
}

//画面表示インターフェース
type viewer interface {
	turnview()
}

//initView 画面初期化
func initView() {
	//各画面を初期化して作成
	defaultFg = termbox.ColorGreen
	defaultBg = termbox.ColorDefault
	topView = &topview{
		false,
	}
	managementdlView = &managementdlview{}
	searchmenuView = &searchmenuview{
		url.Values{},
		"検索条件を決めてください。",
	}
	searchmenufiltergenreView = &searchmenufiltergenreview{
		url.Values{},
		"",
	}
	searchresultView = &searchresultview{
		url.Values{},
		"",
		[]narouAPISearchResultjson{},
		true,
	}
	noveltopView = &noveltopview{
		"",
		"",
		&novelinformation{},
		&narouNovel{},
		[]storyInformation{},
	}
	novelviewerView = &novelview{
		&novelinformation{},
		&narouNovel{},
		&storyInformation{},
		"",
		0,
	}

	SetView(topView) //トップ画面を設定
}

//SetView 引数の画面に切り替える
func SetView(set viewer) {
	set.turnview()
}

//トップ画面
func (view *topview) turnview() {
	//画面構成定義
	initDraw()
	initChoiceList()

	//トップ画面における選択肢の処理
	topmenu := func(num int) {
		switch num {
		case 0:
			//小説を探す
			SetView(searchmenuView) //検索条件1へ
		case 1:
			//入手した小説を読む
			SetView(managementdlView)
		default:
			//その他
		}
	}

	cancelSelection := func() {
		//トップ画面なのでアプリを終了する
		appquiet <- true
	}

	setStrings([]string{
		"小説を探す",
		"入手した小説を読む",
	})
	setExecute(topmenu)
	cancelSetting(true, "終了", cancelSelection)
	setPattern(pat3)
	setSection(5, height-5)

	//描画
	drawLine("なろうが読みたい！", 0, 0, defaultFg, defaultBg)
	drawRow("=", 1, defaultFg, defaultBg)
	drawLine("なろうが読みたい！は「小説家になろう」を閲覧、保存する非公式コンソールビュワーです。", 0, 2, defaultFg, defaultBg)
	drawLine("このソフトを使用して生じた損害や責任の一切を製作者は保証できませんのでご注意ください。", 0, 3, defaultFg, defaultBg)
	drawChoiceList()
	var dlFinishStr string
	if view.downloading {
		//ダウンロード中
		dlFinishStr = "ダウンロード中です"
	} else {
		//ダウンロード完了
		dlFinishStr = "ダウンロードが完了しました"
	}
	drawLine(dlFinishStr, 0, height, defaultFg, defaultBg)
}

//DL管理画面
func (view *managementdlview) turnview() {

}

//検索メニュー
func (view *searchmenuview) turnview() {
	//画面構成定義
	initDraw()
	initChoiceList()

	//検索画面のメニューを定義
	searchMenu := func(num int) {
		switch num {
		case 0:
			//総合評価の高い順
			view.searchFilter.Add("order", "hyoka")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "総合評価の多い順"
			SetView(searchmenufiltergenreView) //ジャンル指定画面へ

		case 1:
			//ブックマーク数の多い順
			view.searchFilter.Add("order", "favnovelcnt")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "ブックマーク数の多い順"
			SetView(searchmenufiltergenreView)

		case 2:
			//レビューの数の多い順
			view.searchFilter.Add("order", "reviewcnt")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "レビューの数の多い順"
			SetView(searchmenufiltergenreView)

		case 3:
			//感想の多い順
			view.searchFilter.Add("order", "impressioncnt")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "感想の多い順"
			SetView(searchmenufiltergenreView)

		case 4:
			//評価者数の多い順
			view.searchFilter.Add("order", "hyokacnt")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "評価者数の多い順"
			SetView(searchmenufiltergenreView)

		case 5:
			//週間ユニークユーザーの多い順
			view.searchFilter.Add("order", "weekly")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "週間ユニークユーザーの多い順"
			SetView(searchmenufiltergenreView)

		case 6:
			//新着順
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "新着順"

		case 7:
			//古い順
			view.searchFilter.Add("order", "old")
			searchmenufiltergenreView.searchFilter = view.searchFilter
			searchmenufiltergenreView.searchString = "古い順"

		case 8:
		//タイトルで検索
		case 9:
			//作者名で検索
		default:
		}
	}
	cancelSelection := func() {
		view.searchFilter = url.Values{} //検索条件を初期化
		SetView(topView)
	}
	setStrings([]string{
		"総合評価の高い順",
		"ブックマーク数の多い順",
		"レビュー数の多い順",
		"感想の多い順",
		"評価者数の多い順",
		"週間ユニークユーザーの多い順",
		"新着順",
		"古い順",
		"タイトルで検索",
		"作者名で検索",
	})
	setExecute(searchMenu)
	cancelSetting(true, "トップ画面に戻る", cancelSelection)
	setPattern(pat3)
	setSection(4, height-4) //画面一番下までを描画範囲

	//描画
	drawLine("なろうが読みたい！", 0, 0, defaultFg, defaultBg)
	drawLine("小説を探す。", 0, 1, defaultFg, defaultBg)
	drawLine(view.searchString, 0, 2, defaultFg, defaultBg)
	drawRow("=", 3, defaultFg, defaultBg)
	drawChoiceList()

}

//検索ジャンル指定画面
func (view *searchmenufiltergenreview) turnview() {
	//画面構成定義
	initDraw()
	initChoiceList()
	var AllGenresStringArray []string

	//検索画面のジャンル指定メニューを定義
	selectMenu := func(num int) {
		//numから配列の文字列を取得し、検索してジャンルを取得する
		str := AllGenresStringArray[num]
		res1, ok1 := biggenres.FindName(str)
		res2, ok2 := smallgenres.FindName(str)

		if str == "全てのジャンル" {
			//全てのジャンルで検索
			searchresultView.searchFilter = view.searchFilter
			searchresultView.searchString = view.searchString + "/" + str
			searchresultView.updateResult = true
			SetView(searchresultView) //検索結果へ
		} else if ok1 {
			//大ジャンルで見つかった時
			//ジャンル指定を追加
			view.searchFilter.Add("biggenre", strconv.Itoa(res1.id))
			searchresultView.searchFilter = view.searchFilter
			searchresultView.searchString = view.searchString + "/" + res1.genreName
			searchresultView.updateResult = true
			SetView(searchresultView)
		} else if ok2 {
			//少ジャンルで見つかった時
			//ジャンル指定を追加
			view.searchFilter.Add("genre", strconv.Itoa(res2.id))
			searchresultView.searchFilter = view.searchFilter
			searchresultView.searchString = view.searchString + "/" + res2.genreName
			searchresultView.updateResult = true
			SetView(searchresultView)
		} else {
			//大小ジャンルでも検索が見つからないならエラーを表示
			drawLine("ジャンル指定ができませんでした。", 0, height, defaultFg, defaultBg)
		}
	}

	cancelSelection := func() {
		view.searchFilter.Del("order")
		view.searchString = ""
		searchmenuView.searchFilter.Del("order")
		SetView(searchmenuView)
	}
	AllGenresStringArray = append([]string{"全てのジャンル"}, getGenreStringArray(biggenres)...)
	AllGenresStringArray = append(AllGenresStringArray, getGenreStringArray(smallgenres)...)
	setStrings(AllGenresStringArray)
	setExecute(selectMenu)
	cancelSetting(true, "戻る", cancelSelection)
	setPattern(pat2)
	setSection(4, height-4)

	//描画
	drawLine("なろうが読みたい！", 0, 0, defaultFg, defaultBg)
	drawLine("小説を探す", 0, 1, defaultFg, defaultBg)
	drawLine(view.searchString, 0, 2, defaultFg, defaultBg)
	drawRow("=", 3, defaultFg, defaultBg)
	drawChoiceList()
}

//検索結果
func (view *searchresultview) turnview() {
	//画面構成定義
	initChoiceList()
	initDraw()

	//フラグがオンのとき検索更新が行われる
	if view.updateResult {
		drawLine(view.searchString+"を取得中。", 0, 0, defaultFg, defaultBg)
		//一覧を取得
		values := view.searchFilter
		view.resultList = []narouAPISearchResultjson{}   //検索結果を代入する構造体
		resultInfoSimple := []narouAPISearchResultjson{} //検索結果単体
		var pos int                                      //現在位置
		maxNum := 50                                     //取得数
		values.Add("gzip", "5")
		values.Add("out", "json")
		values.Add("of", "t-n-w-s") //タイトル、Nコード、作者を取得
		values.Add("lim", "50")     //50件出力

		for pos = 1; pos <= maxNum; pos++ {
			//一つ一つ取得していき、ランキングの順番を保証する
			values.Add("st", strconv.Itoa(pos)) //ランキングの順位

			//なろうAPIから情報を取得
			resp, err := http.Get(narouAPI + "?" + values.Encode())
			if err != nil {
				fmt.Println(err)
			}
			defer resp.Body.Close() //終了処理

			//gzipで圧縮されているので解凍する
			decompbody, err := gzip.NewReader(resp.Body) //解凍のために読み込ませる
			if err != nil {
				fmt.Println(err)
			}
			defer decompbody.Close() //終了処理

			dec := json.NewDecoder(decompbody)  //jsonをデコード
			err = dec.Decode(&resultInfoSimple) //jsonを構造体に代入
			if err != nil {
				fmt.Println(err)
			}
			view.resultList = append(view.resultList, resultInfoSimple[1]) //Allcountのみの構造体を除外し、情報を追加していく
			values.Del("st")

			//情報ロード画面の進捗状況を描画
			Clear()
			drawLineNoStatic("ロード中…"+strconv.Itoa(pos)+"/"+strconv.Itoa(maxNum), 0, 1, defaultFg, defaultBg)
			drawScreen()
		}

		initDraw() //ロード画面消去
	}

	selectNovels := func(num int) {
		selectedNovel := view.resultList[num]    //小説情報を取得
		noveltopView.ncode = selectedNovel.Ncode //小説情報を代入
		noveltopView.title = selectedNovel.Title
		SetView(noveltopView)
	}

	cancelSelection := func() {
		view.searchFilter.Del("genre")
		view.searchFilter.Del("biggenre")
		view.searchString = ""
		view.updateResult = true
		searchmenufiltergenreView.searchFilter.Del("genre")
		searchmenufiltergenreView.searchFilter.Del("biggenre")
		SetView(searchmenufiltergenreView)
	}
	setMultipleLines(ResultListStringArray(view.resultList)) //小説を表示
	setExecute(selectNovels)                                 //表示関数
	cancelSetting(true, "戻る", cancelSelection)
	setPattern(pat2)
	setSection(4, height-4)

	//描画
	drawLine("なろうが読みたい！", 0, 0, defaultFg, defaultBg)
	drawLine(view.searchString+" 検索結果", 0, 1, defaultFg, defaultBg)
	drawLine(strconv.Itoa(len(view.resultList))+"件表示", 0, 2, defaultFg, defaultBg)
	drawRow("=", 3, defaultFg, defaultBg)
	drawChoiceList()
}

//小説詳細トップ
func (view *noveltopview) turnview() {
	//画面構成定義
	initChoiceList()
	initDraw()
	drawLine(view.title+"を取得中。", 0, 0, defaultFg, defaultBg)
	//情報取得
	view.novelInfo = newNovelinformation()
	view.novelInfo.init(view.ncode)
	view.novelStories = newNarouNovel()
	view.novelStories.init(view.ncode, "《", "》") //Nコードとルビを設定
	view.storiesIndex = view.novelStories.getIndexByChapter()

	//読込終了
	initDraw()

	selectStories := func(num int) {
		novelviewerView.ncode = view.novelInfo.ncode
		novelviewerView.currentnum = num + 1 //閲覧話数をセット
		novelviewerView.novelInfo = view.novelInfo
		novelviewerView.novelStories = view.novelStories
		novelviewerView.storyInfo = &view.storiesIndex[num+1]
		SetView(novelviewerView)
	}

	selectCancel := func() {
		view.ncode = ""
		view.title = ""
		view.novelInfo = newNovelinformation()
		view.novelStories = newNarouNovel()
		searchresultView.updateResult = false
		SetView(searchresultView)
	}

	setExecute(selectStories)
	cancelSetting(true, "小説一覧に戻る", selectCancel)
	setPattern(pat2)
	setMultipleLines(stories2LinesArray(view.storiesIndex))
	setSection(6, height-6)

	//描画
	drawLine("なろうが読みたい！", 0, 0, defaultFg, defaultBg)
	drawRow("=", 1, defaultFg, defaultBg)
	drawLine(view.novelInfo.title, 0, 2, defaultFg, defaultBg)
	drawLine("作者："+view.novelInfo.author, 0, 3, defaultFg, defaultBg)
	drawLine(view.novelInfo.keyword, 0, 4, defaultFg, defaultBg)
	drawRow("=", 5, defaultFg, defaultBg)
	drawChoiceList()
}

//小説各話
func (view *novelview) turnview() {
	//画面構成定義
	initChoiceList()
	initDraw()

	//Escキーを押したときの動作
	doCancel := func() {
		view.ncode = ""
		view.currentnum = 0
		SetView(noveltopView)
	}

	nextPage := func() {
		nextViewer := novelview{}
		nextViewer.ncode = view.novelInfo.ncode
		nextViewer.currentnum = view.currentnum + 1 //閲覧話数をセット
		nextViewer.novelInfo = view.novelInfo
		nextViewer.novelStories = view.novelStories
		nextViewer.storyInfo = &noveltopView.storiesIndex[nextViewer.currentnum]
		novelviewerView = &nextViewer
		SetView(novelviewerView)
	}

	previousPage := func() {
		previousViewer := novelview{}
		previousViewer.ncode = view.novelInfo.ncode
		previousViewer.currentnum = view.currentnum - 1 //閲覧話数をセット
		previousViewer.novelInfo = view.novelInfo
		previousViewer.novelStories = view.novelStories
		previousViewer.storyInfo = &noveltopView.storiesIndex[previousViewer.currentnum]
		novelviewerView = &previousViewer
		SetView(novelviewerView)
	}

	viewer := NewMultiLineViewer()
	var header []string //頭に追加する次のページとかの指示
	viewerScreen := []string{
		view.novelInfo.title,
		view.storyInfo.chapterTitle,
		"作者：" + view.novelInfo.author,
		strconv.Itoa(view.currentnum) + "/" + strconv.Itoa(view.novelInfo.allcount),
		stringJoinRow("=", width-8),
		view.storyInfo.subTitle,
		stringJoinRow("=", width-8),
	}
	viewerScreen = append(viewerScreen, view.novelStories.getStory(view.currentnum)...)
	viewer.Init()
	viewer.CancelSetting(doCancel)
	if view.currentnum == 1 && view.novelInfo.allcount == 1 {
		//全一話のときは使うことが出来ない
		viewer.SetLeftRightFunc(func() {}, func() {})
		header = []string{""} //ヘッダーに指示なし

	} else if view.currentnum == 1 {
		//第一話目の時は前のページに戻れない
		viewer.SetLeftRightFunc(func() {}, nextPage)
		header = []string{"次のページへ→"}

	} else if view.novelInfo.allcount == view.currentnum {
		//最終話の時は次のページへ進めない
		viewer.SetLeftRightFunc(previousPage, func() {})
		header = []string{"←前のページへ"}

	} else {
		//前後のページへの遷移
		viewer.SetLeftRightFunc(previousPage, nextPage)
		header = []string{"次のページへ→", "←前のページへ"}
	}
	viewerScreen = append(header, viewerScreen...)
	viewer.SetStrings(viewerScreen)
	viewer.Draw()
}
