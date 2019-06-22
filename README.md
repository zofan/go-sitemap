[![Build Status](https://travis-ci.org/zofan/go-sitemap.svg?branch=master)](https://travis-ci.org/zofan/go-sitemap)
[![Go Report Card](https://goreportcard.com/badge/github.com/zofan/go-sitemap)](https://goreportcard.com/report/github.com/zofan/go-sitemap)
[![Coverage Status](https://coveralls.io/repos/github/zofan/go-sitemap/badge.svg?branch=master)](https://coveralls.io/github/zofan/go-sitemap?branch=master)
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/zofan/go-sitemap)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/zofan/go-sitemap/master/LICENSE)
[![Sourcegraph](https://sourcegraph.com/github.com/zofan/go-sitemap/-/badge.svg)](https://sourcegraph.com/github.com/zofan/go-sitemap?badge)
[![Release](https://img.shields.io/github/release/zofan/go-sitemap.svg?style=flat-square)](https://github.com/zofan/go-sitemap/releases)

#### Features
- Support XML and plain formats
- Support invalid XML files
- Support gzip (including invalid)
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