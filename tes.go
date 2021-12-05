package main

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	
	"strings"
	"time"
	"fmt"
)

type Request struct {
	httpreq *http.Request
	Header  *http.Header
	Client  *http.Client
	Cookies []*http.Cookie
}

type Response struct {
	R       *http.Response
	content []byte
	text    string
	req     *Request
}

type Header map[string]string

func Session() *Request {
	req := new(Request)
	req.httpreq = &http.Request{
		Method:     "GET",
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	req.Header = &req.httpreq.Header
	req.Client = &http.Client{}
	jar, _ := cookiejar.New(nil)
	req.Client.Jar = jar
	return req
}

func (req *Request) Get(Url string) (resp *Response, err error) {
	req.httpreq.Method = "GET"
	delete(req.httpreq.Header, "Cookie")
	parseUrl, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}
	req.httpreq.URL = parseUrl
	req.Client.Jar.SetCookies(req.httpreq.URL, req.Cookies)
	res, err := req.Client.Do(req.httpreq)
	if err != nil {
		return nil, err
	}
	resp = &Response{}
	resp.R = res
	resp.req = req
	resp.Content()
	defer res.Body.Close()
	return resp, nil
}

func (req *Request) Post(Url string, Body string) (resp *Response, err error) {
	req.httpreq.Method = "POST"
	delete(req.httpreq.Header, "Cookie")
	req.httpreq.Body = ioutil.NopCloser(strings.NewReader(Body))
	URL, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}
	req.httpreq.URL = URL
	req.Client.Jar.SetCookies(req.httpreq.URL, req.Cookies)
	res, err := req.Client.Do(req.httpreq)
	req.httpreq.Body = nil
	req.httpreq.GetBody = nil
	req.httpreq.ContentLength = 0
	if err != nil {
		return nil, err
	}
	resp = &Response{}
	resp.R = res
	resp.req = req
	resp.Content()
	defer res.Body.Close()
	return resp, nil
}

func (resp *Response) Content() []byte {
	var err error
	if len(resp.content) > 0 {
		return resp.content
	}
	var Body = resp.R.Body
	if resp.R.Header.Get("Content-Encoding") == "gzip" && resp.req.Header.Get("Accept-Encoding") != "" {
		reader, err := gzip.NewReader(Body)
		if err != nil {
			return nil
		}
		Body = reader
	}
	resp.content, err = ioutil.ReadAll(Body)
	if err != nil {
		return nil
	}
	return resp.content
}

func (resp *Response) Text() string {
	if resp.content == nil {
		resp.Content()
	}
	resp.text = string(resp.content)
	return resp.text
}

func (resp *Response) Json(v interface{}) error {
	if resp.content == nil {
		resp.Content()
	}
	return json.Unmarshal(resp.content, v)
}

func (resp *Response) Cookies() (cookies []*http.Cookie) {
	httpreq := resp.req.httpreq
	client := resp.req.Client
	cookies = client.Jar.Cookies(httpreq.URL)
	return cookies
}

func (req *Request) SetCookie(cookie *http.Cookie) {
	req.Cookies = append(req.Cookies, cookie)
}

func (req *Request) SetHeader(header Header) {
	for k, v := range header {
		req.Header.Set(k, v)
	}
}

func main() {
    s := Session()
    ts := time.Now().UnixMilli()
    for i:=0;i<20;i++ {
        s.Get("https://shopee.co.id")
    }
    fmt.Println(float32(time.Now().UnixMilli()-ts)/1e3)
}
