package util

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// ファイルが存在するか確認
func ExistsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 文字列をスペースとタブでtrimする
func Trim(str string) string {
	return strings.Trim(str, " 　\t")
}

// sliceにsearchが含まれているか
func ContainsStr(search string, slice []string) bool {
	for _, val := range slice {
		if val == search {
			return true
		}
	}
	return false
}

// 指定したディレクトリにあるpatternにマッチするファイルリストを取得する
func FileList(dir string, pattern string) ([]string, error) {
	files := []string{}

	if stat, err := os.Stat(dir); stat == nil || !stat.IsDir() {
		log.Println(dir + " is not directory")
		return files, err
	}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return files, nil
	}

	for _, info := range infos {
		name := info.Name()
		matched, err := regexp.Match(pattern, []byte(name))
		if err != nil {
			return files, err
		}
		if matched {
			files = append(files, name)
		}
	}

	return files, nil
}
