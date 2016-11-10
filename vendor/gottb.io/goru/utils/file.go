package utils

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gottb.io/goru/errors"
)

func PkgPath(filename string) (string, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return "", errors.Wrap(err)
	}
	gopathAbsPath, err := filepath.Abs(os.Getenv("GOPATH"))
	if err != nil {
		return "", errors.Wrap(err)
	}
	path := filename
	if !stat.IsDir() {
		path = filepath.Dir(filename)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return filepath.ToSlash(strings.TrimPrefix(absPath, gopathAbsPath+"/src/")), nil
}

type IgnoreRules []string

func (r IgnoreRules) IsIgnored(name string, ignoreFilename string) bool {
	if name == ignoreFilename {
		return true
	}
	for _, rule := range r {
		if matched, _ := filepath.Match(rule, name); matched {
			return true
		}
	}
	return false
}

func LoadIgnoreRule(folder, ignoreFilename string) IgnoreRules {
	f, err := os.Open(filepath.Join(folder, ignoreFilename))
	rules := IgnoreRules{}
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			rules = append(rules, scanner.Text())
		}
		if err = scanner.Err(); err != nil {
			return rules
		}
	}
	return rules
}

type WalkFunc func(path string, info os.FileInfo) error

func Walk(root string, walkFn WalkFunc, ignoreFilename string) error {
	rules := LoadIgnoreRule(filepath.Dir(root), ignoreFilename)
	if rules.IsIgnored(filepath.Base(root), ignoreFilename) {
		return nil
	}
	return walk(root, walkFn, ignoreFilename)
}

func walk(root string, walkFn WalkFunc, ignoreFilename string) error {
	info, err := os.Stat(root)
	if err != nil {
		info = nil
	}
	if info != nil && info.IsDir() {
		rules := LoadIgnoreRule(root, ignoreFilename)
		childrenInfos, err := ioutil.ReadDir(root)
		if err != nil {
			return errors.Wrap(err)
		}
		for _, childInfo := range childrenInfos {
			if rules.IsIgnored(childInfo.Name(), ignoreFilename) {
				continue
			}
			err = walk(filepath.Join(root, childInfo.Name()), walkFn, ignoreFilename)
			if err != nil {
				return err
			}
		}
	}
	return walkFn(root, info)
}
