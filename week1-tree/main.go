package main

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

const (
	graphicIndent     = "│\t"
	graphicIndentLast = "\t"
	graphicItem       = "├───"
	graphicItemLast   = "└───"
)

type fileList []fs.FileInfo

func (fl fileList) Len() int {
	return len(fl)
}

func (fl fileList) Swap(i int, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl fileList) Less(i int, j int) bool {
	return fl[i].Name() < fl[j].Name()
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return walk(out, path, 0, false, "", true, true, printFiles)
}

func walk(out io.Writer, path string, level int, isLast bool, indentPrefix string, recursive bool, needSort bool, printFiles bool) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	prefix := indentPrefix + getItemPrefix(isLast)

	if fileInfo.IsDir() {
		if level != 0 {
			fmt.Fprintln(out, prefix+fileInfo.Name())
		}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		if level == 0 || recursive {
			isLastPrev := isLast
			fileList := prepareFileList(files, needSort, printFiles)
			for i, file := range fileList {
				subPath := path + string(os.PathSeparator) + file.Name()
				isLast = (i == len(fileList)-1)
				err = walk(out, subPath, level+1, isLast, indentPrefix+getIndentPrefix(isLastPrev, level), recursive, needSort, printFiles)
				if err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Fprintln(out, prefix+fileInfo.Name(), getHumanSize(fileInfo.Size()))
	}
	return nil
}

func prepareFileList(files []fs.FileInfo, needSort bool, printFiles bool) []fs.FileInfo {
	var result []fs.FileInfo
	for _, file := range files {
		if file.IsDir() {
			result = append(result, file)
		}
		if printFiles && !file.IsDir() {
			result = append(result, file)
		}
	}
	if needSort {
		sort.Sort(fileList(result))
	}
	return result
}

func getIndentPrefix(isLast bool, level int) string {
	if level == 0 {
		return ""
	}
	if isLast {
		return graphicIndentLast
	} else {
		return graphicIndent
	}
}

func getItemPrefix(isLast bool) string {
	if isLast {
		return graphicItemLast
	} else {
		return graphicItem
	}
}

func getHumanSize(size int64) string {
	if size == 0 {
		return "(empty)"
	}
	return "(" + strconv.FormatInt(size, 10) + "b)"
}
