package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func isExist(path string)(bool){
	_, err := os.Stat(path)
	if err != nil{
		if os.IsExist(err){
			return true
		}
		if os.IsNotExist(err){
			return false
		}
		return false
	}
	return true
}

func getImag(url string, name string) {
	fmt.Println(name)
	image, err := os.Create(name)
	if err != nil {
		fmt.Printf("name: %s create failed", url)
		return
	}
	defer image.Close()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("url: %s save failed, err: %v", url, err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	_, err = image.Write(data)
	if err != nil {
		fmt.Println(err)
	}
	wg.Done()
	return
}

func MakeSaveDir() string {
	fmt.Printf("保存目录: ")
	reader := bufio.NewReader(os.Stdin)
	saveDir, _, _ := reader.ReadLine()

	if isExist(string(saveDir)) {
		return string(saveDir)
	}

	err := os.MkdirAll(string(saveDir), os.ModePerm)
	if err != nil {
		log.Fatalf("目录不合法: %s, err: %v\n", saveDir, err)
	}

	return string(saveDir)
}



func Download(doc *goquery.Document) (int, error){
	saveDir := MakeSaveDir()
	// 记录存储的图片个数
	count := 0
	doc.Find("enclosure").Each(func(i int, s *goquery.Selection) {
		if content, exist := s.Attr("url"); exist {
			realUrl := func(url []string) string {
				ans := ""
				for _, u := range url[:len(url) - 1] {
					ans += u + "/"
				}
				return strings.TrimRight(ans, "/")
			}(strings.Split(content, "/"))

			name := strconv.Itoa(count) + ".jpg"
			wg.Add(1)
			go getImag(realUrl, path.Join(saveDir, name))
			count++
		}
	})
	wg.Wait()
	fmt.Println("Finish")
	return count, nil
}

func main() {
	fmt.Printf("输入作家主页链接:")
	reader := bufio.NewReader(os.Stdin)
	link, _, _ := reader.ReadLine()

	// 作品均封装在<link>/rss下
	rssLink := string(link) + "/rss"

	res, err := http.Get(rssLink)
	if err != nil {
		log.Fatalf("解析链接异常: %v\n", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Status code: %d; err: %v\n ", res.StatusCode, err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	_, err = Download(doc)
}
