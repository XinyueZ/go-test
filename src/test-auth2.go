//Example to post a message on http://www.oschina.net/
pckage main

import (
	"bytes"
	"fmt"

	"net/http"
	"strconv"
)

const (
	API_RESTYPE = "application/x-www-form-urlencoded"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func main() {
	ch := make(chan bool)
	defer func() {
		if e := recover(); e != nil {
			close(ch)
		}
	}()
	go writeMessage("hello,osc", ch)
	<-ch
}

func writeMessage(msg string, ch chan bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(Error)
		}
	}()

	client := new(http.Client)
	jsonStr := fmt.Sprintf(`access_token=c092f5e3-aea7-49a7-9c84-396980173353&msg=%s`, msg)
	if r, e := http.NewRequest("POST", "https://www.oschina.net/action/openapi/tweet_pub", bytes.NewBufferString(jsonStr)); e == nil {
		r.Header.Add("Content-Type", API_RESTYPE)
		r.Header.Add("Content-Length", strconv.Itoa(len(jsonStr)))
		if resp, e := client.Do(r); e == nil {
			fmt.Println(resp.Status)
			ch <- true
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
	return
}
