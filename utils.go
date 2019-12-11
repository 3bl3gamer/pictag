package main

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/ansel1/merry"
)

type Env struct {
	Val string
}

func (e *Env) Set(name string) error {
	if name != "dev" && name != "prod" {
		return merry.New("wrong env: " + name)
	}
	e.Val = name
	return nil
}

func (e Env) String() string {
	return e.Val
}

func (e Env) Type() string {
	return "string"
}

func (e Env) IsDev() bool {
	return e.Val == "dev"
}

func (e Env) IsProd() bool {
	return e.Val == "prod"
}

func MakeCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", merry.Wrap(err)
	}
	dir = dir + "/pictag"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", merry.Wrap(err)
	}
	return dir, nil
}

func MakeConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", merry.Wrap(err)
	}
	dir = dir + "/pictag"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", merry.Wrap(err)
	}
	return dir, nil
}

type StringsFlag []string

func (f *StringsFlag) String() string {
	return strings.Join(*f, " ")
}

func (f *StringsFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func loadImagesTags(dir string) (map[string]string, error) {
	tags := make(map[string]string)
	f, err := os.Open(dir + "/tags.json")
	if os.IsNotExist(err) {
		return tags, nil
	}
	if err != nil {
		return nil, merry.Wrap(err)
	}
	if err := json.NewDecoder(f).Decode(&tags); err != nil {
		return nil, merry.Wrap(err)
	}
	return tags, nil
}
