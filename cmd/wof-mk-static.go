package main

// this is a badly named file - we'll figure it out... (20170203/thisisaaronland)

import (
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Parse(path string) (string, error) {

	fh, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	bytes, err := ioutil.ReadAll(fh)

	if err != nil {
		return "", err
	}

	result := gjson.GetBytes(bytes, "wof:id")
	wofid := int(result.Int())

	rel_path, err := uri.Id2Path(wofid)

	if err != nil {
		return "", err
	}

	fname := fmt.Sprintf("%d.json", wofid)
	new_path := filepath.Join(rel_path, fname)

	return new_path, nil
}

func Copy(src string, dest string) error {

	root := filepath.Dir(dest)

	_, err := os.Stat(root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(root, 0755)

		if err != nil {
			return err
		}

	}

	bytes, err := ioutil.ReadFile(src)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(dest, bytes, 0644)
}

func main() {

	var static = flag.String("static", "", "...")

	flag.Parse()

	info, err := os.Stat(*static)

	if os.IsNotExist(err) {
		log.Fatal(err)
	} else {

		if !info.IsDir() {
			log.Fatal("Not a directory")
		}
	}

	possible := flag.Args()

	for _, src_path := range possible {

		rel_path, err := Parse(src_path)

		if err != nil {
			log.Fatal(err)
		}

		dest_path := filepath.Join(*static, rel_path)

		log.Printf("copy %s to %s\n", src_path, dest_path)

		err = Copy(src_path, dest_path)

		if err != nil {
			log.Fatal(err)
		}

	}
}
