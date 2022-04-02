// nolint: funlen
package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"golang.org/x/net/html"
)

type (
	httpClientWithoutError        struct{}
	httpClientWithParsingError    struct{}
	httpClientWithError           struct{}
	httpClientWithNonOKStatusCode struct{}
)

func (*httpClientWithoutError) Get(_ *url.URL) (*http.Response, error) {
	b, _ := os.ReadFile("./test-data/correct.html.txt")
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(b)),
	}, nil
}

func (*httpClientWithParsingError) Get(_ *url.URL) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(iotest.ErrReader(io.ErrUnexpectedEOF)),
	}, nil
}

func (*httpClientWithNonOKStatusCode) Get(_ *url.URL) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusBadGateway,
	}, nil
}

func (*httpClientWithError) Get(_ *url.URL) (*http.Response, error) {
	return nil, errors.New("error occurred")
}

func TestNew(t *testing.T) {
	r, _ := (&httpClientWithoutError{}).Get(nil)
	correctNode, _ := html.Parse(r.Body)
	_ = r.Body.Close()

	type args struct {
		webAddress string
		client     HTTPClient
	}
	tests := []struct {
		name    string
		args    args
		want    *Scraper
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				webAddress: "https://someAddress",
				client:     &httpClientWithoutError{},
			},
			want: &Scraper{doc: correctNode},
		},
		{
			name: "nullable client",
			args: args{
				webAddress: "https://someAddress",
				client:     nil,
			},
			wantErr: true,
		},
		{
			name: "incorrect webAddress",
			args: args{
				webAddress: "ht12s-o.ate@#$%^&*()_@stvaml",
				client:     &httpClientWithoutError{},
			},
			wantErr: true,
		},
		{
			name: "empty webAddress",
			args: args{
				webAddress: "",
				client:     &httpClientWithoutError{},
			},
			wantErr: true,
		},
		{
			name: "error during going GET request",
			args: args{
				webAddress: "https://someAddress",
				client:     &httpClientWithError{},
			},
			wantErr: true,
		},
		{
			name: "status code is not 200",
			args: args{
				webAddress: "https://someAddress",
				client:     &httpClientWithNonOKStatusCode{},
			},
			wantErr: true,
		},
		{
			name: "invalid html content",
			args: args{
				webAddress: "https://someAddress",
				client:     &httpClientWithParsingError{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.webAddress, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type parseElementNumberTestCase struct {
	name    string
	argStr  string
	want    uint
	wantErr bool
}

func TestParseElementNumber(t *testing.T) {
	tests := []parseElementNumberTestCase{
		{
			name:    "correct",
			argStr:  "someTag[9]",
			want:    9,
			wantErr: false,
		},
		{
			name:    "correct",
			argStr:  "someTag1[9]",
			want:    9,
			wantErr: false,
		},
		{
			name:    "without number_1",
			argStr:  "someTag",
			want:    1,
			wantErr: false,
		},
		{
			name:    "without number_2",
			argStr:  "someTag2",
			want:    1,
			wantErr: false,
		},
		{
			name:    "empty string",
			argStr:  "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_1",
			argStr:  "someTag[1",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_2",
			argStr:  "someTag1]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_3",
			argStr:  "someTag]1[",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_4",
			argStr:  "[someTag",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_5",
			argStr:  "]someTag",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_6",
			argStr:  "someTag[a]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_7",
			argStr:  "someT[ag3626]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_8",
			argStr:  "[999]someTag",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_9",
			argStr:  "[999]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_10",
			argStr:  "someTag[[999]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_11",
			argStr:  "someTag[999]]]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_12",
			argStr:  "someTag[888][999]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_13",
			argStr:  "[124]someTag[999]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_15",
			argStr:  "someTag[999]0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_16",
			argStr:  "someT^%$ag[999]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "incorrect_17",
			argStr:  ";:(№:[999]",
			want:    0,
			wantErr: true,
		},
	}
	testParseNumberWithFunc(t, parseElement, tests, "without-regex")
	testParseNumberWithFunc(t, parseElementWithRegex, tests, "with-regex")
}

func testParseNumberWithFunc(t *testing.T, parseFunc func(string) (uint, error), tests []parseElementNumberTestCase, postfix string) {
	for _, tt := range tests {
		n := fmt.Sprintf("%s-%s", tt.name, postfix)
		t.Run(n, func(t *testing.T) {
			got, err := parseFunc(tt.argStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("[%s] parseFunc() error = %v, wantErr %v", postfix, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("[%s] parseFunc() got = %v, want %v", postfix, got, tt.want)
			}
		})
	}
}

func TestNewHTTPClientWithRetry(t *testing.T) {
	type args struct {
		retries      uint
		retryTimeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    HTTPClient
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				retries:      10,
				retryTimeout: 30 * time.Second,
			},
			want: &httpClientWithRetry{
				client: http.Client{Transport: &http.Transport{
					DisableKeepAlives: true,
					MaxIdleConns:      10,
					IdleConnTimeout:   30 * time.Second,
				}},
				retries:      10,
				retryTimeout: 30 * time.Second,
			},
		},
		{
			name: "negative retryTimeout",
			args: args{
				retries:      30,
				retryTimeout: -30 * time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTTPClientWithRetry(tt.args.retries, tt.args.retryTimeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHTTPClientWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHTTPClientWithRetry() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSquareBracketsIndexes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		wantO int
		wantC int
	}{
		{
			name:  "with brackets",
			args:  args{s: "some[text]"},
			wantO: 4,
			wantC: 9,
		},
		{
			name:  "without brackets",
			args:  args{s: "someText"},
			wantO: -1,
			wantC: -1,
		},
		{
			name:  "empty string",
			args:  args{s: ""},
			wantO: -1,
			wantC: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotO, gotC := getSquareBracketsIndexes(tt.args.s)
			if gotO != tt.wantO {
				t.Errorf("getSquareBracketsIndexes() gotO = %v, want %v", gotO, tt.wantO)
			}
			if gotC != tt.wantC {
				t.Errorf("getSquareBracketsIndexes() gotC = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}

func BenchmarkParseElementNumber(b *testing.B) {
	v := strings.Split(readTestTagPaths(), "\n")
	correct := v[0]
	invalid := v[1]
	for i := 0; i < b.N; i++ {
		_, _ = parseElement(correct)
		_, _ = parseElement(invalid)
	}
}

func BenchmarkParseElementNumberWithRegex(b *testing.B) {
	v := strings.Split(readTestTagPaths(), "\n")
	correct := v[0]
	invalid := v[1]
	for i := 0; i < b.N; i++ {
		_, _ = parseElementWithRegex(correct)
		_, _ = parseElementWithRegex(invalid)
	}
}

func TestScraperGetValue(t *testing.T) {
	r, _ := (&httpClientWithoutError{}).Get(nil)
	correctDoc, _ := html.Parse(r.Body)
	_ = r.Body.Close()

	type fields struct {
		doc *html.Node
	}
	type args struct {
		fullXPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "correct",
			fields: fields{doc: correctDoc},
			args:   args{fullXPath: "/html/body/div[1]/div[1]/div[1]/div[2]/div[3]/div/div[2]/div/div[1]/div[1]/div/span[1]/span/span/span/span/span/span/span/span/span/span/span/span"},
			want:   "12 981 - 14 444",
		},
		{
			name:    "non text node type",
			fields:  fields{doc: correctDoc},
			args:    args{fullXPath: "/html"},
			wantErr: true,
		},
		{
			name:    "without pathDelimiter",
			fields:  fields{doc: correctDoc},
			args:    args{fullXPath: "html/body/div[1]/div[1]/div[1]/div[2]/div[3]/div/div[2]/div/div[1]/div[1]/div/span[1]/span/span"},
			wantErr: true,
		},
		{
			name:    "incorrect path",
			fields:  fields{doc: correctDoc},
			args:    args{fullXPath: "/html/7956ody/div[1]/div[1]/div[1]/di1v[2]/div[3]/div/diw4v[2]/div/div[1]/div[1]/div/span[1]/span/span"},
			wantErr: true,
		},
		{
			name:    "element not found",
			fields:  fields{doc: correctDoc},
			args:    args{fullXPath: "/html/body/div[1]/div[1]/div[1]/div[2]/div[3]/div/div[2]/div/div[1]/div[1]/div/span[1]/span/span/div"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scraper{
				doc: tt.fields.doc,
			}
			got, err := s.GetValue(tt.args.fullXPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultHTTPClientWithRetry(t *testing.T) {
	tests := []struct {
		name string
		want HTTPClient
	}{
		{
			name: "correct_1",
			want: &httpClientWithRetry{
				client: http.Client{
					Transport: &http.Transport{
						DisableKeepAlives: true,
						MaxIdleConns:      10,
						IdleConnTimeout:   30 * time.Second,
					},
				},
				retries:      3,
				retryTimeout: 30 * time.Second,
			},
		},
		{
			name: "correct global variable",
			want: DefaultHTTPClient,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultHTTPClientWithRetry(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultHTTPClientWithRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func readTestTagPaths() string {
	b, _ := os.ReadFile("./test-data/tagPaths.txt")
	return string(b)
}
