package main

import (
	"github.com/nsf/termbox-go"
)

//setting
var (
	width    int       //画面横幅
	height   int       //画面縦幅
	appquiet chan bool //このチャンネルに送信を行うとアプリが終了する
)

func run() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.Clear(defaultFg, defaultBg)

	width, height = termbox.Size()
	appquiet = make(chan bool)
	go inputLoop()   //入力待機
	initDraw()       //表示処理初期化
	initChoiceList() //選択肢初期化
	initView()       //画面構成初期化
	<-appquiet
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}
