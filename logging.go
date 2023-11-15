package main

/*логирование в несколько потоков на своих каналах*/

import (
	"fmt"
	"time"
)

type loggingElem struct {
	ltype string    // respomse or request or error
	u_id  string    // session id
	head  []byte    // headers
	body  []byte    // body
	t     time.Time //timestamp
}

var (
	counts   int                = 1
	channels []chan loggingElem = nil
)

func logginginit(countsize int, size int) {
	counts = countsize
	// на всякий
	if counts <= 0 {
		counts = 1
	}
	if size <= 0 {
		size = 1
	}
	channels = []chan loggingElem{make(chan loggingElem, size)}
	for i := 2; i <= counts; i++ {
		channels = append(channels, make(chan loggingElem, size))
	}
}

/***********************выбор каналов *******************************/

func loggingsend(elem loggingElem) {

	selected := 0
	for i := 0; i < counts; i++ {
		b := loggingselector(elem, channels[i])
		if b {
			selected = i + 1
			break
		}
	}
	if selected > 0 {
		fmt.Printf("selected channel %d\n", selected)
	} else {
		fmt.Println("no selected channel")
	}
}

func loggingselector(elem loggingElem, cnl chan loggingElem) bool {
	// non blocking send
	select {
	case cnl <- elem:
		return true
	default:
		return false
	}
}

/***********************консьюмеры и отправители *******************************/

func loggingrun() {

	for i := 0; i < counts; i++ {

		go func(n int, chn chan loggingElem) {
			for {
				elem, runned := <-chn
				fmt.Printf("consume channel %d\n", n)
				bytestoKafka(elem)
				if !runned {
					break
				}
			}
		}(i+1, channels[i])

	}

}
