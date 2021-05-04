package main

import (
	"io"
	"log"
	"net/http"
)

func main() {

	//buf := bytes.NewBuffer(make([]byte, resp.ContentLength))
	//n, err := buf.ReadFrom(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("written %d bytes\n", n)
}

func DoRequestDiscardCopy() int64 {
	resp, err := http.Get("https://hotline.ua/computer-videokarty/gigabyte-geforce-gtx-1060-g1-gaming-6g-gv-n1060g1-gaming-6gd")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("resp body close error: %s", closeErr.Error())
		}
	}()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("status code is not 200: %d", resp.StatusCode)
	}
	b, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func DoRequestReadAll() int64 {
	resp, err := http.Get("https://hotline.ua/computer-videokarty/gigabyte-geforce-gtx-1060-g1-gaming-6g-gv-n1060g1-gaming-6gd")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("resp body close error: %s", closeErr.Error())
		}
	}()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("status code is not 200: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return int64(len(b))
}
