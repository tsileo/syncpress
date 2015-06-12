package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/GeertJohan/go.ask"
	"github.com/extemporalgenome/slug"
	"github.com/spf13/cobra"
	"github.com/tsileo/syncpress"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var emptyPost = `title: {{ .title }}
slug: {{ .slug }}
date: {{ .date }}

# {{ .title }}

<!--more-->
`

var tpl = template.Must(template.New("post").Parse(emptyPost))

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var cmdNew = &cobra.Command{
		Use:   "new [post title]",
		Short: "Create an empty blog post with the given title",
		Run: func(cmd *cobra.Command, args []string) {
			buf := bytes.NewBufferString("")
			title := args[0]
			slu := slug.Slug(title)
			fmt.Printf("Created with %v", slu)
			payload := map[string]string{
				"title": title,
				"slug":  slu,
				"date":  time.Now().Format("2006-01-02 15:04:05"),
			}
			if err := tpl.Execute(buf, payload); err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(slu+".md", buf.Bytes(), 0644); err != nil {
				panic(err)
			}
		},
	}
	var cmdSync = &cobra.Command{
		Use:   "sync [path]",
		Short: "Sync the given path with the database",
		Run: func(cmd *cobra.Command, args []string) {
			//fmt.Printf("%v", ask.MustAskf("ok?"))
			session, err := mgo.Dial(os.Getenv("SYNCPRESS_MONGODB"))
			dbname := os.Getenv("SYNCPRESS_DB")
			defer session.Close()
			check(err)
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			lindex := map[string]*syncpress.Post{}
			lposts, err := syncpress.PostsFromPath(path)
			check(err)
			col := session.DB(dbname).C(syncpress.ColPosts)
			colraw := session.DB(dbname).C(syncpress.ColRaw)
			for _, post := range lposts {
				lindex[post.Slug] = post
				rpost := &syncpress.Post{}
				err := col.Find(bson.M{"slug": post.Slug}).One(rpost)
				if err == mgo.ErrNotFound {
					if ask.MustAskf("Upload new post: \"%v\" ?", post.Title) {
						if err := col.Insert(post); err != nil {
							panic(err)
						}
						if err := colraw.Insert(bson.M{"hash": post.Hash, "raw": post.Raw}); err != nil {
							panic(err)
						}
						fmt.Printf("post uploaded.\n")
					}
				} else {
					check(err)
				}
				if post.Hash != rpost.Hash {
					// TODO handle conflict
				}
			}
			rposts, err := syncpress.PostsFromDB(session, dbname)
			for _, post := range rposts {
				lpost, exists := lindex[post.Slug]
				if exists {
					if post.Hash != lpost.Hash {
						// TODO handle conflict
					}

				} else {
					if ask.MustAskf("Remove post from database: \"%v\" ?", post.Title) {
						if err := col.Remove(bson.M{"hash": post.Hash}); err != nil {
							panic(err)
						}
						if err := colraw.Remove(bson.M{"hash": post.Hash}); err != nil {
							panic(err)
						}
					} else {
						if ask.MustAskf("Redownload post from database: \"%v\" ?", post.Title) {
							raw := map[string]interface{}{}
							if err := colraw.Find(bson.M{"hash": post.Hash}).One(&raw); err != nil {
								panic(err)
							}
							if err := ioutil.WriteFile(filepath.Join(path, post.Slug+".md"), raw["raw"].([]byte), 0644); err != nil {
								panic(err)
							}
						}
					}
				}
			}
			fmt.Printf("Sync done.")
		},
	}
	var rootCmd = &cobra.Command{Use: "syncpress"}
	rootCmd.AddCommand(cmdNew, cmdSync)
	rootCmd.Execute()
}
