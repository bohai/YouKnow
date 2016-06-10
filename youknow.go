package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
)

const (
	STARTURL = "http://cl.miicool.info/thread0806.php?fid=16"
	CAOLIU   = "http://cl.miicool.info/"
	TAIL     = "  草榴社區  - powered by phpwind.net"
)

var (
	mu sync.Mutex
)

func main() {
	fmt.Println("start crawl:" + STARTURL)
	mux := fetchbot.NewMux()
	//first page handler, get topic url
	mux.Response().Path("/thread0806.php").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			reader, err := iconv.NewReader(res.Body, "gb18030", "utf-8")
			if err != nil {
				fmt.Println("iconv", err)
				return
			}
			doc, err := goquery.NewDocumentFromReader(reader)
			if err != nil {
				fmt.Println("goquery", err)
				mu.Lock()
				fmt.Println(res.Status, "retry")
				ctx.Q.SendStringGet(res.Request.URL.String())
				mu.Unlock()
				return
			}
			doc.Find("h3").Each(func(i int, contentSelection *goquery.Selection) {
				href, _ := contentSelection.Find("a").Attr("href")
				txt := contentSelection.Find("a").Text()
				fmt.Println(href, txt)

				if !strings.Contains(href, "htm_data/16/1606") {
					return
				}
				mu.Lock()
				ctx.Q.SendStringGet(CAOLIU + href)
				mu.Unlock()
			})

		}))

	mux.Response().Path("/htm_data/16/1606").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			reader, err := iconv.NewReader(res.Body, "gb18030", "utf-8")
			if err != nil {
				fmt.Println("iconv", err)
				return
			}
			doc, err := goquery.NewDocumentFromReader(reader)
			if err != nil {
				fmt.Println(err)
				return
			}
			var path string
			doc.Find("title").Each(func(i int, contentSelection *goquery.Selection) {
				path = strings.Trim(contentSelection.Text(), TAIL)
				os.Mkdir(path, 0777)
				os.Chdir(path)
				fmt.Println(path)
			})
			doc.Find("div.tpc_content").Eq(0).Find("input").Each(func(i int, contentSelection *goquery.Selection) {
				href, _ := contentSelection.Attr("src")
				fmt.Println(href)
			})
			os.Chdir("../")
		}))
	f := fetchbot.New(mux)
	q := f.Start()
	q.SendStringGet(STARTURL)
	q.Block()
}