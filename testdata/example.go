package example

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Display(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(f)
	if err == io.EOF {
		fmt.Println("EOF")
	}
	if !(err == nil) {
		return fmt.Errorf("reading %q: %v", fname, err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func slurpURL(urlStr string) []byte {
	if *useCache {
		log.Fatalf("Invalid use of slurpURL in cached mode for URL %s", urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Fatal(err)
	}
	if *publicOnly {
		req.Header.Add("X-User-IP", "0.0.0.0") // hack
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
