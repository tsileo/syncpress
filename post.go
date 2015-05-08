package syncpress

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/extemporalgenome/slug"
	"github.com/jinzhu/now"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v1"
)

type Post struct {
	Hash      string
	Title     string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Excerpt   []byte
	Body      []byte
}

func openPost(path string) (*Post, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	res, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data := bytes.Split(res, []byte("\n\n"))
	meta := map[string]string{}
	if err := yaml.Unmarshal(data[0], &meta); err != nil {
		return nil, err
	}
	body := res[len(data[0])+2:]
	excerpt := []byte{}
	if exc := bytes.Split(body, []byte("<!--more-->")); len(exc) >= 0 {
		excerpt = blackfriday.MarkdownCommon(exc[0])
	}
	body = blackfriday.MarkdownCommon(body)
	createdAt, err := now.Parse(meta["date"])
	if err != nil {
		return nil, err
	}
	var updatedAt time.Time
	if updt, updtOk := meta["updated"]; updtOk {
		updatedAt, err = now.Parse(updt)
		if err != nil {
			return nil, err
		}
	}
	post := &Post{
		Title:     meta["title"],
		Slug:      slug.Slug(meta["title"]),
		Body:      body,
		Excerpt:   excerpt,
		Hash:      fmt.Sprintf("%x", sha1.Sum(res)),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return post, nil
}
