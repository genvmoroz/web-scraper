package scraper

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const (
	pathDelimiter      = "/"
	digits             = "1234567890"
	notAllowedSymbols  = "!@#$%^&*()_+-={}!\"№;'<>/\\~`:?*"
	tagRegexPattern    = "^[A-Za-z]+(\\d+)?(\\[\\d+]{1})?$"
	openSquareBracket  = '['
	closeSquareBracket = ']'
)

type (
	HTTPClient interface {
		Get(*url.URL) (*http.Response, error)
	}

	httpClientWithRetry struct {
		client       http.Client
		retries      uint
		retryTimeout time.Duration
	}

	Scraper struct {
		doc *html.Node
	}
)

// DefaultHTTPClient is a HTTPClient with configured retry: retries = 3, retryTimeout = 30s
var DefaultHTTPClient = defaultHTTPClientWithRetry()

func New(webAddress string, client HTTPClient) (*Scraper, error) {
	if !utf8.ValidString(webAddress) {
		return nil, errors.New("webAddress is not valid utf8 string")
	}
	if client == nil {
		return nil, errors.New("client should be not nil")
	}
	webAddress = strings.TrimSpace(webAddress)
	if webAddress == "" {
		return nil, errors.New("webAddress should be not empty")
	}

	parsedURL, err := url.Parse(webAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url [%s]: %w", webAddress, err)
	}

	resp, err := client.Get(parsedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to do GET request to url [%s]: %w", webAddress, err)
	}
	defer func() {
		if resp == nil || resp.Body == nil {
			return
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("resp body close error: %s", closeErr.Error())
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is not 200: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content as html: %s", err)
	}

	return &Scraper{
		doc: doc,
	}, nil
}

func (s *Scraper) GetValue(fullXPath string) (string, error) {
	if !utf8.ValidString(fullXPath) {
		return "", errors.New("fullXPath is not valid utf8 string")
	}

	if !strings.HasPrefix(fullXPath, pathDelimiter) {
		return "", fmt.Errorf("should have a prefix \"/\"")
	}

	return parseValue(strings.Split(fullXPath[1:], pathDelimiter), s.doc)
}

func parseValue(path []string, rootNode *html.Node) (string, error) {
	if len(path) == 0 {
		if rootNode.Type == html.TextNode {
			return rootNode.Data, nil
		} else {
			return "", fmt.Errorf("failed to get string value from node, NodeType: %v", rootNode.Type)
		}
	}

	var (
		targetTagName      = path[0]
		tagsCount     uint = 1
	)

	tagNum, err := parseElement(targetTagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse element number: %w", err)
	}

	if strings.ContainsRune(targetTagName, '[') {
		targetTagName = targetTagName[:strings.IndexByte(targetTagName, '[')]
	}

	for n := rootNode; n != nil; n = n.NextSibling {
		switch n.Type {
		case html.ErrorNode:
			return "", errors.New("node processing error")
		case html.DocumentNode:
			return parseValue(path, n.FirstChild)
		case html.DoctypeNode:
			return parseValue(path, n.NextSibling)
		case html.ElementNode:
			if n.Data == targetTagName {
				if tagsCount == tagNum {
					return parseValue(path[1:], n.FirstChild)
				} else {
					tagsCount++
				}
			}
		}
	}

	return "", errors.New("element not found")
}

// parseElement parses html element by path, returns its number or error if occurred
func parseElement(path string) (uint, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, errors.New("empty string")
	}

	m := strings.IndexAny(path, notAllowedSymbols)
	if m != -1 {
		return 0, errors.New("the tag contains a not allowed symbol")
	}

	o, c := getSquareBracketsIndexes(path)
	if (o == -1) != (c == -1) || c < o || (c != -1 && c != len(path)-1) || o == 0 {
		return 0, errors.New("brackets are arranged incorrectly")
	}

	d := strings.IndexAny(path, digits)
	if d > c && c != -1 {
		return 0, errors.New("the tag number is out of brackets")
	}

	if o == -1 && c == -1 {
		return 1, nil
	}

	n, err := strconv.Atoi(path[o+1 : c])
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int: %w", err)
	}

	return uint(n), nil
}

func parseElementWithRegex(s string) (uint, error) {
	match, err := regexp.MatchString(tagRegexPattern, s)
	if err != nil {
		return 0, fmt.Errorf("failed to match string with pattern %s: %w", tagRegexPattern, err)
	}
	if !match {
		return 0, fmt.Errorf("%s does not match with regex pattern %s", s, tagRegexPattern)
	}

	o, c := getSquareBracketsIndexes(s)

	if o == -1 && c == -1 {
		return 1, nil
	}

	n, err := strconv.Atoi(s[o+1 : c])
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int: %w", err)
	}

	return uint(n), nil
}

// getSquareBracketsIndexes returns indexes of square brackets
//	- first returning value is index of open square bracket
//	- second returning value is index of close square bracket
func getSquareBracketsIndexes(s string) (o int, c int) {
	return strings.IndexByte(s, openSquareBracket), strings.IndexByte(s, closeSquareBracket)
}

func NewHTTPClientWithRetry(retries uint, retryTimeout time.Duration) (HTTPClient, error) {
	if retries < 1 {
		return nil, errors.New("retries should not be less than 1")
	}
	if retryTimeout < 0 {
		return nil, errors.New("retryTimeout should not be negative")
	}

	return &httpClientWithRetry{
		client: http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				MaxIdleConns:      10,
				IdleConnTimeout:   30 * time.Second},
		},
		retries:      retries,
		retryTimeout: retryTimeout,
	}, nil
}

func defaultHTTPClientWithRetry() HTTPClient {
	return &httpClientWithRetry{
		client: http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				MaxIdleConns:      10,
				IdleConnTimeout:   30 * time.Second,
			},
		},
		retries:      3,
		retryTimeout: 30 * time.Second,
	}
}

func (c *httpClientWithRetry) Get(url *url.URL) (*http.Response, error) {
	if url == nil {
		return nil, errors.New("url cannot be nil")
	}
	if c.retries < 1 {
		return nil, errors.New("retries should not be less than 1")
	}
	if c.retryTimeout < 0 {
		return nil, errors.New("retryTimeout should not be negative")
	}

	req := &http.Request{Method: http.MethodGet, URL: url, Header: make(map[string][]string)}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")

	var (
		err  error
		resp *http.Response
	)
	for retry := c.retries; retry > 0; retry-- {
		resp, err = c.client.Do(req)
		if err == nil {
			return resp, nil
		}
		log.Printf("failed to do GET request: %s. Retrying", err.Error())
		time.Sleep(c.retryTimeout)
	}

	return nil, fmt.Errorf("execution request timeout: %w", err)
}
