package example

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func twolines(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		log.Print("got error", err)
		return err
	}
	_ = f
	return nil
}

func WriteFile(fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte{1})
	if err != nil {
		return err
	}
	_, err = f.Write([]byte{8, 17, 33})
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	fmt.Println("wrote", fname)
	return nil
}

func Display(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(f)
	if !(err == nil) {
		return fmt.Errorf("reading %q: %v", fname, err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

var useCache bool

func slurpURL(urlStr string) []byte {
	if useCache {
		log.Fatalf("Invalid use of slurpURL in cached mode for URL %s", urlStr)
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error fetching URL %s: %v", urlStr, err)
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading body of URL %s: %v", urlStr, err)
	}
	return bs
}
