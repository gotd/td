package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := filepath.Walk("_testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, readErr := ioutil.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		s := bufio.NewScanner(bytes.NewReader(data))
		for s.Scan() {
			text := s.Text()
			// name, O_RDWR|O_CREATE|O_TRUNC, 0666
			if !strings.HasSuffix(text, ";") {
				continue
			}
			targetName := fmt.Sprintf("testdata:%x", md5.Sum(s.Bytes()))
			targetPath := filepath.Join("_fuzz", "corpus", targetName)
			if err := ioutil.WriteFile(targetPath, []byte(strings.TrimSuffix(text, ";")), 0666); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
