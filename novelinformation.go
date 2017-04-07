package main

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const narouAPI = "http://api.syosetu.com/novelapi/api/" //なろうAPIのURL
const narouAPITimeLayout = "2006-01-02 15:04:05"        //なろうAPIにおける日付のフォーマット

//ジャンルの定義構造体
type genre struct {
	id        int    //ジャンルのID
	genreName string //ジャンル名
}

//String stringでジャンル名を返す
func (g genre) String() string {
	return g.genreName
}

//ジャンルをジャンルIDでソートするための配列の型宣言
type genres []genre

func (g genres) Len() int {
	return len(g)
}

func (g genres) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func (g genres) Less(i, j int) bool {
	return g[i].id < g[j].id
}

//FindName ジャンル名で検索
func (g genres) FindName(str string) (resp genre, ok bool) {
	resp = genre{}
	ok = false
	for _, v := range g {
		if v.genreName == str {
			resp = v
			ok = true
		}
	}
	return
}

//FindID IDで検索
func (g genres) FindID(id int) (resp genre, ok bool) {
	resp = genre{}
	ok = false
	for _, v := range g {
		if v.id == id {
			resp = v
			ok = true
		}
	}
	return
}

//getGenreStringArray ジャンル構造体ををStringの配列で返す
func getGenreStringArray(ggs genres) []string {
	garray := []string{}
	sort.Sort(ggs) //ソート
	for _, v := range ggs {
		garray = append(garray, v.String())
	}
	return garray
}

var (
	biggenres genres = []genre{
		{1, "恋愛"},
		{2, "ファンタジー"},
		{3, "文芸"},
		{4, "SF"},
		{99, "その他"},
		{98, "ノンジャンル"},
	}

	smallgenres genres = []genre{
		{101, "異世界[恋愛]"},
		{102, "現実世界[恋愛]"},
		{201, "ハイファンタジー[ファンタジー]"},
		{202, "ローファンタジー[ファンタジー]"},
		{301, "純文学[文芸]"},
		{302, "ヒューマンドラマ[文芸]"},
		{303, "歴史[文芸]"},
		{304, "推理[文芸]"},
		{305, "ホラー[文芸]"},
		{306, "アクション[文芸]"},
		{307, "コメディ[文芸]"},
		{401, "VRゲーム[SF]"},
		{402, "宇宙[SF]"},
		{403, "空想科学[SF]"},
		{404, "パニック[SF]"},
		{9901, "童話[その他]"},
		{9902, "詩[その他]"},
		{9903, "エッセイ[その他]"},
		{9904, "リプレイ[その他]"},
		{9999, "その他[その他]"},
		{9801, "ノンジャンル[ノンジャンル]"},
	}
)

//novelinformation 小説家になろうの小説情報
type novelinformation struct {
	//基本情報
	ncode            string    //小説のID
	title            string    //小説のタイトル
	author           string    //著者
	allcount         int       //小説の話数
	firstpostingdate time.Time //初回掲載日
	lastpostingdate  time.Time //最終掲載日
	novelupdatedat   time.Time //小説の更新日時
	isrensai         bool      //連載作品なら真(短編は偽)
	isend            bool      //完結済みなら真
	isr15            bool      //R15作品なら真
	isbl             bool      //BL作品なら真
	isgl             bool      //ガールズラブ作品なら真
	iszankoku        bool      //「残酷な描写あり」なら真
	istensei         bool      //「異世界転生」なら真
	istenni          bool      //「異世界転移」なら真
	synopsis         string    //あらすじ
	keyword          string    //キーワード
	biggenre         genre     //大ジャンル
	smallgenre       genre     //ジャンル
	//使用者による情報
	currentcount int  //現在読んでいる話数
	islock       bool //更新を行わないならTrue
}

//なろうAPIで取得した構造体を小説情報へ挿入するための中間構造体
type narouAPIjson struct {
	//タイプ１
	Allcount int `json:"allcount"` //infoデータの総数
	//タイプ２
	Title          string `json:"title"`
	Writer         string `json:"writer"`
	Story          string `json:"story"`
	Biggenre       int    `json:"biggenre"`
	Genre          int    `json:"genre"`
	Keyword        string `json:"keyword"`
	GeneralFirstup string `json:"general_firstup"`
	GeneralLastup  string `json:"general_lastup"`
	NovelType      int    `json:"noveltype"`
	End            int    `json:"end"`
	GeneralAllNo   int    `json:"general_all_no"`
	Isr15          int    `json:"isr15"`
	Isbl           int    `json:"isbl"`
	Isgl           int    `json:"isgl"`
	Iszankoku      int    `json:"iszankoku"`
	Istensei       int    `json:"istensei"`
	Istenni        int    `json:"istenni"`
	NovelupdatedAt string `json:"novelupdated_at"`
}

//なろうAPIで取得した検索結果を代入する構造体
type narouAPISearchResultjson struct {
	//タイプ1
	Allcount int `json:"allcount"`
	//タイプ2
	Title  string `json:"title"`
	Writer string `json:"writer"`
	Story  string `json:"story"`
	Ncode  string `json:"ncode"`
}

//ResultListStringArray 項目の一覧を取得する
func ResultListStringArray(v []narouAPISearchResultjson) []Lines {
	sa := []Lines{}
	for _, v := range v {
		sa = append(sa, Lines([]string{v.Title, "作者    　：" + v.Writer, "あらすじ　：" + v.Story, ""}))
	}
	return sa
}

func newNovelinformation() *novelinformation {
	return &novelinformation{}
}

//init 小説家になろうの小説情報を引数のNコードから入手する。
func (info *novelinformation) init(ncode string) (*novelinformation, error) {
	values := url.Values{}
	var intermediateinfo []narouAPIjson //変換するための中間情報構造体
	values.Add("gzip", "5")             //gzipで圧縮レベルを5を指定
	values.Add("out", "json")           //jsonで出力
	values.Add("ncode", ncode)          //出力するNcodeを指定
	values.Add("of", "t-w-s-bg-g-k-gf-gl-nt-e-ga-ir-ibl-igl-izk-its-iti-nu")

	resp, err := http.Get(narouAPI + "?" + values.Encode()) //なろうAPIから情報を取得

	if err != nil {
		return &novelinformation{}, err
	}
	defer resp.Body.Close() //終了処理

	//gzipで圧縮されているので解凍する
	decompbody, err := gzip.NewReader(resp.Body) //解凍のために読み込ませる
	if err != nil {
		return &novelinformation{}, err
	}
	defer decompbody.Close() //終了処理

	dec := json.NewDecoder(decompbody) //jsonをデコード
	err = dec.Decode(&intermediateinfo)
	if err != nil {
		return &novelinformation{}, err
	}

	//jsonから読み込まれた中間構造体を扱う形に移植する
	info.ncode = ncode                                                                              //Nコード
	info.title = intermediateinfo[1].Title                                                          //タイトル
	info.author = intermediateinfo[1].Writer                                                        //著者
	info.allcount = intermediateinfo[1].GeneralAllNo                                                //総話数
	info.firstpostingdate, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].GeneralFirstup) //初回掲載日時
	if err != nil {
		info.firstpostingdate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.lastpostingdate, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].GeneralLastup) //最終掲載日時
	if err != nil {
		info.lastpostingdate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.novelupdatedat, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].NovelupdatedAt) //小説の更新日時
	if err != nil {
		info.novelupdatedat = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.isrensai = (intermediateinfo[1].NovelType == 1)
	info.isend = (intermediateinfo[1].End == 0)
	info.isr15 = (intermediateinfo[1].Isr15 == 1)
	info.isbl = (intermediateinfo[1].Isbl == 1)
	info.isgl = (intermediateinfo[1].Isgl == 1)
	info.iszankoku = (intermediateinfo[1].Iszankoku == 1)
	info.istensei = (intermediateinfo[1].Istensei == 1)
	info.istenni = (intermediateinfo[1].Istenni == 1)
	info.synopsis = intermediateinfo[1].Story  //あらすじ
	info.keyword = intermediateinfo[1].Keyword //キーワード
	bg, _ := biggenres.FindID(intermediateinfo[1].Biggenre)
	info.biggenre = bg
	sg, _ := smallgenres.FindID(intermediateinfo[1].Genre)
	info.smallgenre = sg
	info.currentcount = 0 //まだしおりを挟んでいない状態
	info.islock = false   //更新ロックをかけていない状態

	return info, nil //無事に処理が終了したs
}

//update 小説家になろうの小説情報を更新する
func (info *novelinformation) update() {
	values := url.Values{}
	var intermediateinfo []narouAPIjson //変換するための中間情報構造体
	values.Add("gzip", "5")             //gzipで圧縮レベルを5を指定
	values.Add("out", "json")           //jsonで出力
	values.Add("ncode", info.ncode)     //出力するNcodeを指定
	values.Add("of", "t-w-s-bg-g-k-gf-gl-nt-e-ga-ir-ibl-igl-izk-its-iti-nu")

	resp, err := http.Get(narouAPI + "?" + values.Encode()) //なろうAPIから情報を取得

	if err != nil {
		return
	}
	defer resp.Body.Close() //終了処理

	//gzipで圧縮されているので解凍する
	decompbody, err := gzip.NewReader(resp.Body) //解凍のために読み込ませる
	if err != nil {
		return
	}
	defer decompbody.Close() //終了処理

	dec := json.NewDecoder(decompbody) //jsonをデコード
	err = dec.Decode(&intermediateinfo)
	if err != nil {
		return
	}

	//jsonから読み込まれた中間構造体を扱う形に移植する
	info.title = intermediateinfo[1].Title                                                          //タイトル
	info.author = intermediateinfo[1].Writer                                                        //著者
	info.allcount = intermediateinfo[1].GeneralAllNo                                                //総話数
	info.firstpostingdate, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].GeneralFirstup) //初回掲載日時
	if err != nil {
		info.firstpostingdate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.lastpostingdate, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].GeneralLastup) //最終掲載日時
	if err != nil {
		info.lastpostingdate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.novelupdatedat, err = time.Parse(narouAPITimeLayout, intermediateinfo[1].NovelupdatedAt) //小説の更新日時
	if err != nil {
		info.novelupdatedat = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local) //取得失敗の場合2000/1/1/0/0/00の日付が代入される
		err = nil
	}
	info.isrensai = (intermediateinfo[1].NovelType == 1)
	info.isend = (intermediateinfo[1].End == 0)
	info.isr15 = (intermediateinfo[1].Isr15 == 1)
	info.isbl = (intermediateinfo[1].Isbl == 1)
	info.isgl = (intermediateinfo[1].Isgl == 1)
	info.iszankoku = (intermediateinfo[1].Iszankoku == 1)
	info.istensei = (intermediateinfo[1].Istensei == 1)
	info.istenni = (intermediateinfo[1].Istenni == 1)
	info.synopsis = intermediateinfo[1].Story //あらすじ
	bg, _ := biggenres.FindID(intermediateinfo[1].Biggenre)
	info.biggenre = bg
	sg, _ := smallgenres.FindID(intermediateinfo[1].Genre)
	info.smallgenre = sg
	info.currentcount = 0 //まだしおりを挟んでいない状態
	info.islock = false   //更新ロックをかけていない状態
}
