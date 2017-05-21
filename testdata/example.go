package example

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func Display(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err == io.EOF {
		fmt.Println("EOF")
	}
	if !(err == nil) {
		return fmt.Errorf("reading %q: %v", fname, err)
	}

	a, b := pair(fname)
	fmt.Println(a, b)
	fmt.Println(string(bytes))
	return nil
}

func pair(s string) (int, int) {
	return 3, 5
}
