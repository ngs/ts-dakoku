package app

import (
	"os"
	"sort"
	"testing"
)

func TestBindataRead(t *testing.T) {
	b, err := bindataRead([]byte(""), "foo")
	for _, test := range []Test{
		{0, len(b)},
		{`Read "foo": EOF`, err.Error()},
	} {
		test.Compare(t)
	}
}

func TestAssetDir(t *testing.T) {
	files, err := AssetDir("assets")
	sort.Strings(files)
	Test{[]string{
		"favicon.ico",
		"index.html",
		"success.html",
	}, files}.DeepEqual(t)
	Test{true, err == nil}.Compare(t)
}

func TestAssetNames(t *testing.T) {
	names := AssetNames()
	sort.Strings(names)
	Test{[]string{
		"assets/favicon.ico",
		"assets/index.html",
		"assets/success.html",
	}, names}.DeepEqual(t)
}

func TestAssetInfo(t *testing.T) {
	info, err := AssetInfo("assets/index.html")
	info2, err2 := AssetInfo("assets/index2.html")
	for _, test := range []Test{
		{true, err == nil},
		{"assets/index.html", info.Name()},
		{int64(719), info.Size()},
		{"-rw-r--r--", info.Mode().String()},
		{false, info.ModTime().IsZero()},
		{false, info.IsDir()},
		{true, info.Sys() == nil},
		{"AssetInfo assets/index2.html not found", err2.Error()},
		{true, info2 == nil},
	} {
		test.Compare(t)
	}
}

func TestRestoreAsset(t *testing.T) {
	defer os.RemoveAll(".restored-assets")
	os.RemoveAll(".restored-assets")
	RestoreAssets(".restored-assets", "assets")
	stat, _ := os.Stat(".restored-assets/assets/index.html")
	for _, test := range []Test{
		{"index.html", stat.Name()},
		{int64(719), stat.Size()},
	} {
		test.Compare(t)
	}
}
