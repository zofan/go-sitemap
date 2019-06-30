[![Go Report Card](https://goreportcard.com/badge/github.com/zofan/go-sitemap)](https://goreportcard.com/report/github.com/zofan/go-sitemap)
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/zofan/go-sitemap)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/zofan/go-sitemap/master/LICENSE)
[![Sourcegraph](https://sourcegraph.com/github.com/zofan/go-sitemap/-/badge.svg)](https://sourcegraph.com/github.com/zofan/go-sitemap?badge)
[![Code Climate](https://codeclimate.com/github/zofan/go-sitemap/badges/gpa.svg)](https://codeclimate.com/github/zofan/go-sitemap)
[![Test Coverage](https://codeclimate.com/github/zofan/go-sitemap/badges/coverage.svg)](https://codeclimate.com/github/zofan/go-sitemap)
[![HitCount](http://hits.dwyl.io/zofan/go-sitemap.svg)](http://hits.dwyl.io/zofan/go-sitemap)

#### Features
- Support XML and plain formats
- Support invalid XML files
- Support gzip
- Normalizer of invalid urls
- Not used xml parser or another external dependency
- Callback for deep parsing large sitemaps

#### Install

> go get -u github.com/zofan/go-sitemap

#### Usage example

```$go
package main

import (
	"github.com/zofan/go-sitemap"
	"fmt"
	"net/http"
)

func main() {
	client := &http.Client{}
	req, _ := http.NewRequest(`GET`, `https://www.bbc.com/sitemaps/https-index-com-news.xml`, nil)

	resp, _ := client.Do(req)
	if resp == nil {
		return
	}

	sitemap.ParseResponse(resp, sitemap.CallbackWithClient(client, func (item *sitemap.Item) {
		fmt.Println(item)
	}))
	resp.Body.Close()
}
```