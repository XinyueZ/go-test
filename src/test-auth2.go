//Example to post a message on http://www.oschina.net/
package main

import (
	"bytes"
	"fmt"

	"net/http"

	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
)

const (
	AGENT        = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36"
	API_REQTYPE  = "application/x-www-form-urlencoded; charset=UTF-8"
	LOGIN_URL    = "https://www.oschina.net/action/user/hash_login"
	AUTH_URL     = "https://www.oschina.net/action/oauth2/authorize"
	TOKEN_URL    = "https://www.oschina.net/action/openapi/token"
	TOKEN_BODY   = "client_id=%s&client_secret=%s&grant_type=%s&redirect_uri=%s&code=%s&dataType=%s"
	AUTH_REF_URL = "https://www.oschina.net/action/oauth2/authorize?response_type=code&client_id=" + APP_ID + "&redirect_uri=" + REDIRECT_URL
	REDIRECT_URL = "http://wanlingzhao.eu.pn/index.html"
	SCOPE        = "tweet_api"
	GRANT_TYPE   = "authorization_code"
	RET_TYPE     = "json"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func main() {
	chLogin := make(chan *Logined)
	chWriteMsg := make(chan bool)
	defer func() {
		if e := recover(); e != nil {
			close(chLogin)
			close(chWriteMsg)
		}
	}()

	//go writeMessage("yo ho°°°°", ch)
	usr := newUser(ACCOUNT, PWD, APP_ID, APP_SEC)
	go usr.login(chLogin)
	pLogined := <-chLogin
	go writeMessage("放假放假放假", pLogined, chWriteMsg)
	<-chWriteMsg
}

type Token struct {
	UID          int    `json:"uid"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}


func (self *Token) toString() (s string) {
	json, _ := json.Marshal(self)
	s = string(json)
	return
}




type Logined struct {
	Cookie *http.Cookie
	Token  *Token
}

func (self *Logined) toString() (s string) {
	s = self.Token.toString()
	return
}

type User struct {
	Account  string
	Password string
	AppId    string
	AppSec   string
}

func newUser(account, password, appId, appSec string) (usr *User) {
	usr = new(User)
	usr.Account = account
	usr.Password = password
	usr.AppId = appId
	usr.AppSec = appSec
	return
}

func (self *User) buildLoginBody() (body string) {
	body = fmt.Sprintf(`email=%s&pwd=%s`, self.Account, self.Password)
	return
}

func (self *User) buildOAuth2Body() (body string) {
	body = fmt.Sprintf(`client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=""&user_oauth_approval=true&email=%s&pwd=%s`, self.AppId, REDIRECT_URL, SCOPE, self.Account, self.Password)
	return
}

func (self *User) login(ch chan *Logined) {
	defer func() {
		if e := recover(); e != nil {
			close(ch)
		}
	}()

	h := sha1.New()
	io.WriteString(h, self.Password)
	self.Password = hex.EncodeToString(h.Sum(nil))

	client := new(http.Client)
	body := self.buildLoginBody()

	pLogined := new(Logined)
	if r, e := http.NewRequest("POST", LOGIN_URL, bytes.NewBufferString(body)); e == nil {
		r.Header.Add("Accept", "*/*")
		r.Header.Add("Accept-Encoding", "gzip,deflate,sdch")
		r.Header.Add("Accept-Language", "zh-CN,zh;q=0.8")
		r.Header.Add("Connection", "keep-alive")
		r.Header.Add("Content-Type", API_REQTYPE)
		r.Header.Add("Host", "www.oschina.net")
		r.Header.Add("Origin", "https://www.oschina.net")
		r.Header.Add("User-Agent", AGENT)
		r.Header.Add("X-Requested-With", "XMLHttpRequest")
		r.Header.Add("Referer", AUTH_REF_URL)

		if resp, e := client.Do(r); e == nil {
			//Get cookie, and do OAuth2 in order to fetching "code".
			for _, v := range resp.Cookies() {
				if v.Value != "" {
					pLogined.Cookie = v
					code := self.oAuth2(v)
					pLogined.Token = getToken(code)
					break
				}
			}
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}

	ch <- pLogined
}

func (self *User) oAuth2(cookie *http.Cookie) (code string) {
	client := new(http.Client)
	body := self.buildOAuth2Body()

	if r, e := http.NewRequest("POST", AUTH_URL, bytes.NewBufferString(body)); e == nil {
		r.Header.Add("Accept", "*/*")
		r.Header.Add("Accept-Encoding", "gzip,deflate,sdch")
		r.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
		r.Header.Add("Connection", "keep-alive")
		r.Header.Add("Content-Type", API_REQTYPE)
		r.Header.Add("Host", "www.oschina.net")
		r.Header.Add("X-Requested-With", "XMLHttpRequest")
		r.Header.Add("User-Agent", AGENT)
		r.Header.Add("Referer", AUTH_REF_URL)
		r.Header.Add("Pragma", "no-cache")
		r.Header.Add("Cache-Control", "no-cache")
		r.Header.Add("Cache-Control", "no-cache")
		r.Header.Add("Cookie", "oscid="+cookie.Value)

		if resp, e := client.Do(r); e == nil {
			args := resp.Request.URL.Query()
			code = args["code"][0]
			fmt.Printf("code = %s\n", code)
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
	return
}

func getToken(code string) (pToken *Token) {
	client := new(http.Client)
	body := fmt.Sprintf(TOKEN_BODY, APP_ID, APP_SEC, GRANT_TYPE, REDIRECT_URL, code, RET_TYPE)
	if r, e := http.NewRequest("POST", TOKEN_URL, bytes.NewBufferString(body)); e == nil {
		r.Header.Add("Content-Type", API_REQTYPE)
		if resp, e := client.Do(r); e == nil {
			pToken = new(Token)
			if bytes, e := ioutil.ReadAll(resp.Body); e == nil {
				if e = json.Unmarshal(bytes, pToken); e != nil {
					pToken = nil
				}
			} else {
				panic(e)
			}
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
	return
}

func writeMessage(msg string, pLogined *Logined, ch chan bool) {
	client := new(http.Client)
	jsonStr := fmt.Sprintf(`access_token=%s&msg=%s`, pLogined.Token.AccessToken, msg)
	if r, e := http.NewRequest("POST", "https://www.oschina.net/action/openapi/tweet_pub", bytes.NewBufferString(jsonStr)); e == nil {
		r.Header.Add("Content-Type", API_REQTYPE)
		if resp, e := client.Do(r); e == nil {
			printResponse(resp)
			ch <- true
		} else {
			panic(e)
		}
	} else {
		panic(e)
	}
	return
}

func printResponse(resp *http.Response) {
	fmt.Println("Resp status: " + resp.Status)
	fmt.Printf("Resp resp: %v\n", resp)
	fmt.Printf("Resp header: %v\n", resp.Header)

	if bytes, e := ioutil.ReadAll(resp.Body); e == nil {
		fmt.Println("Resp body: " + string(bytes))
	} else {
		panic(e)
	}
}
