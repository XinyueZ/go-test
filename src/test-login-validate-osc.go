//Example to login oschina.net
package main

import (
	"bytes"
	"fmt"

	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	OSC                 = "www.oschina.net"
	HOST                = "http://www.oschina.net/action/api/"
	API_RESTYPE         = "application/x-www-form-urlencoded"
	POST                = "POST"
	GET                 = "GET"
	KEEP_ALIVE          = "Keep-Alive"
	LOGIN_SCHEME        = `username=%s&pwd=%s&keep_login=1`
	LOGIN_VALIDATE_HTTP = HOST + "login_validate"
	TWEET_LIST          = HOST + "tweet_list?uid=%s&pageIndex=%d&pageSize=25"
	TWEET_PUB           = HOST + "tweet_pub"
	TWEET_PUB_SCHEME    = "uid=%s&msg=%s"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func makeHeader(r *http.Request, cookie string, length int) {
	r.Header.Add("Content-Type", API_RESTYPE)
	r.Header.Add("Content-Length", strconv.Itoa(length))
	r.Header.Add("Host", OSC)
	r.Header.Add("Connection", KEEP_ALIVE)
	r.Header.Add("Cookie", cookie)
	//r.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 4.4.4; Nexus 4 Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/")
}

func printHeader(r *http.Request) {
	header := r.Header
	for k, v := range header {
		fmt.Println("k:", k, "v:", v)
	}
}

type OsChina struct {
	XMLName xml.Name `xml:"oschina"`
	User    User     `xml:"user"`
}

type User struct {
	Uid  string `xml:"uid"`
	Name string `xml:"name"`
}

func main() {
	chUser := make(chan *User)
	chLogin := make(chan *http.Cookie)
	chTweetList := make(chan string)
	chTweetPub := make(chan string)
	defer func() {
		if e := recover(); e != nil {
			close(chUser)
			close(chLogin)
			close(chTweetList)
			close(chTweetPub)
		}
	}()
	go login(ACCOUNT, PWD, chLogin, chUser)
	cookie := <-chLogin //Got user session.
	session := "oscid=" + cookie.Value
	puser := <-chUser
	if cookie != nil {
		fmt.Println("cookie:" + cookie.Value)
		fmt.Println("expires:" + cookie.Expires.String())
		fmt.Println(puser.Uid)
		go printTweetList(puser, session, 1, chTweetList)
		tweetListContent := <-chTweetList
		if tweetListContent != "" {
			fmt.Println(tweetListContent)
			//Just a randem msg
			msgRandem := "冬天了，好冷"
			go pubTweet(puser, session, msgRandem, chTweetPub)
			pubContent := <-chTweetPub
			if pubContent != "" {
				fmt.Println(pubContent)

			}
		}
	}
}

func login(account string, password string, cookieCh chan *http.Cookie, userCh chan *User) {
	fmt.Println("Login.")
	client := new(http.Client)
	body := fmt.Sprintf(LOGIN_SCHEME, account, password)
	url := LOGIN_VALIDATE_HTTP
	fmt.Println(url)
	if r, e := http.NewRequest(POST, url, bytes.NewBufferString(body)); e == nil {
		makeHeader(r, "", len(body))
		if resp, e := client.Do(r); e == nil {
			fmt.Println(resp.Status)
			var cookie *http.Cookie
			if resp != nil {
				defer resp.Body.Close()
			}
			if bytes, err := ioutil.ReadAll(resp.Body); err == nil {
				var posc OsChina
				if err := xml.Unmarshal(bytes, &posc); err == nil {
					for _, v := range resp.Cookies() {
						if v.Value != "" {
							cookie = v
							//break
						}
						fmt.Println(v)
					}
					cookieCh <- cookie
					userCh <- &(posc.User)
				} else {
					panic(err)
				}

			} else {
				panic(err)
			}
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
}

func printTweetList(puser *User, session string, page int, ch chan string) {
	fmt.Println("Get Tweet-List.")
	client := new(http.Client)
	url := fmt.Sprintf(TWEET_LIST, puser.Uid, page)
	fmt.Println(url)
	if r, e := http.NewRequest(GET, url, nil); e == nil {
		makeHeader(r, session, 0)
		if resp, e := client.Do(r); e == nil {
			fmt.Println(resp.Status)
			if resp != nil {
				defer resp.Body.Close()
			}
			if bytes, e := ioutil.ReadAll(resp.Body); e == nil {
				ch <- string(bytes)
			} else {
				panic(e)
			}
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
}

func pubTweet(puser *User, session string, msg string, ch chan string) {
	fmt.Printf("Pub Tweet: %s\n", msg)
	client := new(http.Client)
	url := TWEET_PUB
	fmt.Println(url)
	body := fmt.Sprintf(TWEET_PUB_SCHEME, puser.Uid, msg)
	fmt.Println(body)
	if r, e := http.NewRequest(POST, url, bytes.NewBufferString(body)); e == nil {
		makeHeader(r, session, len(body))
		printHeader(r)
		if resp, e := client.Do(r); e == nil {
			fmt.Println(resp.Status)
			if resp != nil {
				defer resp.Body.Close()
			}
			if bytes, e := ioutil.ReadAll(resp.Body); e == nil {
				ch <- string(bytes)
			} else {
				panic(e)
			}
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
}
