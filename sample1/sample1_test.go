package sample1

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Sample1 struct {
	Hoge int `json:"hoge"`
	Fuga string `json:"fuga"`
}

// パッケージスコープでのみ宣言可能
//go:embed file/sample1.json
var sample1Bytes []byte

func TestSample1_File(t *testing.T)  {
	sample1 := Sample1{}
	err := json.Unmarshal(sample1Bytes, &sample1)
	assert.NoError(t, err)
	assert.Equal(t, Sample1{
		Hoge: 1,
		Fuga: "2",
	}, sample1)
}

//go:embed file/*
var sample1File embed.FS

//go:embed file/*.json
var sample1FileJson embed.FS

func TestSample1_Dir(t *testing.T)  {
	t.Run("ファイルを参照する", func(t *testing.T) {
		t.Run("ディレクトリから指定する必要がある", func(t *testing.T) {
			sb, err := sample1File.ReadFile("file/sample1.json")
			assert.NoError(t, err)

			sample1 := Sample1{}
			err = json.Unmarshal(sb, &sample1)
			assert.NoError(t, err)
			assert.Equal(t, Sample1{
				Hoge: 1,
				Fuga: "2",
			}, sample1)
		})

		t.Run("ディレクトリを省略する", func(t *testing.T) {
			fileDir, err := fs.Sub(sample1File, "file")
			assert.NoError(t, err)

			sb, err := fileDir.(fs.ReadFileFS).ReadFile("sample1.json")
			assert.NoError(t, err)

			sample1 := Sample1{}
			err = json.Unmarshal(sb, &sample1)
			assert.NoError(t, err)
			assert.Equal(t, Sample1{
				Hoge: 1,
				Fuga: "2",
			}, sample1)
		})
	})

	t.Run("ディレクトリを参照する", func(t *testing.T) {
		t.Run("全ファイル", func(t *testing.T) {
			entries, err := sample1File.ReadDir("file")
			assert.NoError(t, err)

			fileNames := []string{}
			for i := range entries {
				fileNames = append(fileNames, entries[i].Name())
			}
			assert.ElementsMatch(t, fileNames, []string{"sample1.json", "sample1.txt"})
		})

		t.Run("jsonのみ", func(t *testing.T) {
			entries, err := sample1FileJson.ReadDir("file")
			assert.NoError(t, err)

			fileNames := []string{}
			for i := range entries {
				fileNames = append(fileNames, entries[i].Name())
			}
			assert.ElementsMatch(t, fileNames, []string{"sample1.json"})
		})
	})
}

func TestSample1_FileServer(t *testing.T)  {
	fileDir, err := fs.Sub(sample1File, "file")
	assert.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/",http.FileServer(http.FS(fileDir)))
	srv := http.Server{
		Addr:              ":8088",
		Handler:           mux,
	}
	t.Cleanup(func() {
		srv.Close()
	})
	go func() {
		srv.ListenAndServe()
	}()

	time.Sleep(time.Second)
	res, err := http.Get("http://localhost:8088/sample1.json")
	assert.NoError(t, err)
	defer res.Body.Close()

	sample1 := Sample1{}
	err = json.NewDecoder(res.Body).Decode(&sample1)
	assert.NoError(t, err)
	assert.Equal(t, Sample1{
		Hoge: 1,
		Fuga: "2",
	}, sample1)
}
