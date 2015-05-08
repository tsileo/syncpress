package syncpress

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/extemporalgenome/slug"
	"github.com/jinzhu/now"
	"github.com/russross/blackfriday"
	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v1"
)

var (
	DBName   = "syncpress"
	ColPosts = "posts"
	ColRaw   = "raw"
)
var PostTpl = template.Must(template.New("post").Parse(`title: {{ .Title }}
slug: {{ .Slug }}
date: {{ .Date }}

{{ .Body }}`))

type Post struct {
	Raw     []byte    `bson:"-"`
	Hash    string    `bson:"hash"`
	Title   string    `bson:"title"`
	Slug    string    `bson:"slug"`
	Date    time.Time `bson:"date"`
	Updated time.Time `bson:"updated,omitempty"`
	Excerpt []byte    `bson:"excerpt"`
	Body    []byte    `bson:"body"`
}

func PostsFromPath(path string) ([]*Post, error) {
	res := []*Post{}
	posts, err := filepath.Glob(filepath.Join(path, "./*.md"))
	if err != nil {
		return nil, err
	}
	for _, f := range posts {
		p, err := openPost(f)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func PostsFromDB(session *mgo.Session) ([]*Post, error) {
	res := []*Post{}
	//session, err := mgo.Dial(host)
	//defer session.Close()
	//if err != nil {
	//	return nil, err
	//}
	iter := session.DB(DBName).C(ColPosts).Find(nil).Iter()
	if err := iter.All(&res); err != nil {
		return nil, err
	}
	return res, nil
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
		Title:   meta["title"],
		Slug:    slug.Slug(meta["title"]),
		Body:    body,
		Excerpt: excerpt,
		Hash:    fmt.Sprintf("%x", sha1.Sum(res)),
		Date:    createdAt,
		Updated: updatedAt,
		Raw:     res,
	}
	return post, nil
}
