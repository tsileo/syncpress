package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"
	"time"

	"github.com/extemporalgenome/slug"
	"github.com/spf13/cobra"
)

var emptyPost = `title: {{ .title }}
slug: {{ .slug }}
date: {{ .date }}

# {{ .title }}

<!--more-->
`

var tpl = template.Must(template.New("post").Parse(emptyPost))

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

	var rootCmd = &cobra.Command{Use: "syncpress"}
	rootCmd.AddCommand(cmdNew)
	rootCmd.Execute()
}
