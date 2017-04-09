package main

import (
	"github.com/nsf/termbox-go"
)

var (
	lockerChan chan bool //入力禁止をやり取りするチャンネル
	inputLock  bool      //trueの時は処理を受け付けない
	//キー押下時実行変数
	pushKeyArrowUp    func()
	pushKeyArrowDown  func()
	pushKeyArrowLeft  func()
	pushKeyArrowRight func()
	pushKeyEsc        func()
	pushKeyEnterSpace func()
	pushKeyHome       func()
	pushKeyEnd        func()
)

//inputLoop 入力イベントをループで取得(ich:termboxのキーイベントを受け取る。 endch:trueを送信すると終了する)
func inputLoop() {
	inputLock = false
	lockerChan = make(chan bool, 1)
	pushKeyEsc = func() {}
	pushKeyArrowUp = func() {}
	pushKeyArrowDown = func() {}
	pushKeyArrowLeft = func() {}
	pushKeyArrowRight = func() {}
	pushKeyEsc = func() {}
	pushKeyEnterSpace = func() {}
	pushKeyHome = func() {}
	pushKeyEnd = func() {}
	ich := make(chan termbox.Key, 1)
	termbox.SetInputMode(termbox.InputAlt)

	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				ich <- ev.Key
			default:
			}
		}
	}()

	for {
		select {
		//case inputLock = <-lockerChan: //ロックフラグを通信して変更

		case k := <-ich:
			//キーイベントを受け取ったとき
			switch k {
			case termbox.KeyArrowUp:
				pushKeyArrowUp()
			case termbox.KeyArrowDown:
				pushKeyArrowDown()
			case termbox.KeyArrowLeft:
				pushKeyArrowLeft()
			case termbox.KeyArrowRight:
				pushKeyArrowRight()
			case termbox.KeyEnter, termbox.KeySpace:
				pushKeyEnterSpace() //選択項目を実行
			case termbox.KeyEsc, termbox.KeyBackspace:
				pushKeyEsc()
			case termbox.KeyHome, termbox.KeyF1:
				pushKeyHome()
			case termbox.KeyEnd, termbox.KeyF2:
				pushKeyEnd()
			case termbox.KeyF12:
				appquiet <- true //End
			default:
			}
		default:
		}
	}
}

//SetInputLock 入力を禁止する
func SetInputLock(flag bool) {
	lockerChan <- flag
}

//SetInputFunction キー押下時の実行関数を設定する
func SetInputFunction(arrowUp, arrowDown, arrowLeft, arrowRight, esc, enterSpace, home, end func()) {
	pushKeyArrowUp = arrowUp
	pushKeyArrowDown = arrowDown
	pushKeyArrowLeft = arrowLeft
	pushKeyArrowRight = arrowRight
	pushKeyEsc = esc
	pushKeyEnterSpace = enterSpace
	pushKeyHome = home
	pushKeyEnd = end
}
