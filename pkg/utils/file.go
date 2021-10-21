package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Zip(src, dst string) error {
	baseDir := CreateDirIfNotExists(src)
	CreateDirIfNotExists(dst)
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	zw := zip.NewWriter(fw)
	defer func() {
		if err := zw.Close(); err != nil {
			fmt.Printf("[Zip]close file err: %v", err)
		}
	}()

	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) error {
		if errBack != nil {
			return errBack
		}

		// create zip header
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		// remove baseDir
		fh.Name = strings.TrimPrefix(path, baseDir)

		if fi.IsDir() {
			fh.Name += "/"
		}

		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}

		// check mode 
		if !fh.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fr.Close()

		n, err := io.Copy(w, fr)
		if err != nil {
			return err
		}
		fmt.Printf("[Zip]compress success: %s, %d characters of data were written\n", path, n)

		return nil
	})
}

func UnZip(src, dst string) ([]string, error) {
	files := make([]string, 0)
	zr, err := zip.OpenReader(src)
	if err != nil {
		return files, err
	}
	defer zr.Close()

	if dst != "" {
		if err := os.MkdirAll(dst, os.ModePerm); err != nil {
			return files, nil
		}
	}

	for _, file := range zr.File {
		var decodeName string
		if file.Flags == 0 {
			// old file is gbk encode
			i := bytes.NewReader([]byte(file.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := ioutil.ReadAll(decoder)
			decodeName = string(content)
		} else {
			// old file is utf-8 encode
			decodeName = file.Name
		}
		path := filepath.Join(dst, decodeName)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return files, nil
			}
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return files, nil
		}

		fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			fr.Close()
			return files, nil
		}

		n, err := io.Copy(fw, fr)
		if err != nil {
			fw.Close()
			fr.Close()
			return files, nil
		}

		fmt.Printf("[UnZip]decompress success: %s, %d characters of data were written\n", path, n)
		files = append(files, path)
		fw.Close()
		fr.Close()
	}
	return files, nil
}

func CreateDirIfNotExists(name string) string {
	info, err := os.Stat(name)
	if err == nil {
		// is dir
		if info.IsDir() {
			return name
		}
		// is file
		dir, _ := filepath.Split(name)
		return dir
	} else {
		dir, filename := filepath.Split(name)
		// unix hidden path is start with .
		if strings.HasPrefix(filename, ".") {
			if !strings.Contains(strings.TrimPrefix(filename, "."), ".") {
				dir = name
			}
		} else {
			if !strings.Contains(filename, ".") {
				dir = name
			}
		}
		// make dir
		os.MkdirAll(dir, os.ModePerm)
		return dir
	}
}

func GetWorkDir() string {
	pwd, _ := os.Getwd()
	return pwd
}
