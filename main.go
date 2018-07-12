package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Result struct {
	Word          string
	Transcription string
	Translate     string
}
type Body struct {
	Word string
	Body []byte
}

func main() {
	var (
		clientHTTP = &http.Client{}
		wg         = sync.WaitGroup{}

		line   = make(chan string, 10)
		body   = make(chan Body, 10)
		result = make(chan Result, 10)
	)

	file, _ := os.Open("./eng_words.txt")
	defer file.Close()

	fileResult, err := os.Create("./eng_words_new.txt")
	if err != nil {
		fmt.Printf("failed to open result file, %s", err)
	}
	defer fileResult.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	wg.Add(1)
	go func() {
		for scanner.Scan() {
			line <- scanner.Text()
		}
		close(line)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for l := range line {
			word := strings.Trim(l, " ")

			fmt.Println(word)

			url := fmt.Sprintf("https://wooordhunt.ru/word/%s", word)

			b, err := getPage(clientHTTP, url)
			if err != nil {
				log.Fatalf("failed getting page (%s), %s", url, err)
			}

			body <- Body{
				Body: b,
				Word: word,
			}
		}

		close(body)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for b := range body {
			res, err := parseHTML(b)
			if err != nil {
				fmt.Printf("failed to parse html, %s", err)
			}
			result <- res
		}
		close(result)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for r := range result {
			str := fmt.Sprintf("%s = [%s] = %s\n", r.Word, r.Transcription, r.Translate)
			fileResult.WriteString(str)
		}
		wg.Done()
	}()

	wg.Wait()
}

func getPage(c *http.Client, URL string) ([]byte, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:61.0) Gecko/20100101 Firefox/61.0")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed getting page, returns code of status not OK, %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func parseHTML(b Body) (Result, error) {
	decoder := html.NewTokenizer(bytes.NewBuffer(b.Body))

	r := Result{
		Word: b.Word,
	}

	for {
		switch decoder.Next() {
		case html.ErrorToken:
			if decoder.Err() == io.EOF {
				return r, nil
			}
			return Result{}, fmt.Errorf("token err, %s", decoder.Err())

		case html.StartTagToken:
			t := decoder.Token()

			if t.Data != "span" {
				continue
			}

			if wantedTokenByAttr(t.Attr, "class", "transcription") && wantedTokenByAttr(t.Attr, "title", "британская") {
				decoder.Next()

				l := decoder.Token()

				r.Transcription = strings.Trim(strings.Replace(l.String(), "|", "", 2), " ")
			}

			if wantedTokenByAttr(t.Attr, "class", "t_inline_en") {
				decoder.Next()

				l := decoder.Token()

				r.Translate = l.String()
			}
		}
	}
}

func wantedTokenByAttr(attrs []html.Attribute, key, val string) bool {
	for _, attr := range attrs {
		if attr.Key == key && strings.HasPrefix(attr.Val, val) {
			return true
		}
	}
	return false
}
