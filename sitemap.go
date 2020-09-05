package sitemap

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	TypeSitemap = `sitemap`
	TypeURL     = `url`
)

var ErrWrongContentType = errors.New(`wrong content type`)

type Item struct {
	Loc        *url.URL
	ChangeFreq string
	LastMod    time.Time
	Priority   float64
	Type       string
}

func ParseStreamXML(stream io.Reader, callback func(*Item)) error {
	s := bufio.NewScanner(stream)
	s.Split(scanTag)

	var item = &Item{}

	for s.Scan() {
		chunk := strings.TrimSpace(s.Text())

		split := strings.Split(chunk, `>`)
		if len(split) != 2 {
			continue
		}

		tag := strings.ToLower(split[0])
		value := strings.TrimSpace(split[1])

		if tag == `/sitemap` || tag == `/url` {
			if item.Loc == nil {
				continue
			}

			callback(item)
		} else if tag == `sitemap` || tag == `url` {
			item = &Item{Type: tag}
		} else if tag[0] != '/' {
			fillItem(item, tag, value)
		}
	}

	return s.Err()
}

func ParseStreamPlain(stream io.Reader, callback func(*Item)) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		u, err := url.Parse(scanner.Text())
		if err == nil {
			callback(&Item{Loc: u, Type: TypeURL})
		}
	}
}

func ParseResponse(resp *http.Response, callback func(*Item)) error {
	var reader io.Reader
	var err error

	if strings.Contains(resp.Request.URL.String(), `.gz`) ||
		strings.Contains(resp.Header.Get(`Content-Encoding`), `gzip`) {
		reader, err = gzip.NewReader(resp.Body)
		// todo: processing ErrHeader and return original reader
		if err != nil {
			return err
		}
	} else {
		reader = resp.Body
	}

	mime := resp.Header.Get(`Content-Type`)
	baseName := path.Base(resp.Request.URL.Path)

	wrapCallback := func(i *Item) {
		normalizeURL(i.Loc, resp.Request.URL)
		callback(i)
	}

	if strings.Contains(mime, `xml`) || strings.Contains(baseName, `.xml`) {
		return ParseStreamXML(reader, wrapCallback)
	}

	if strings.Contains(mime, `plain`) || strings.Contains(baseName, `.txt`) {
		ParseStreamPlain(reader, wrapCallback)
	}

	return ErrWrongContentType
}

func CallbackWithClient(client *http.Client, callback func(i *Item)) func(i *Item) {
	return func(i *Item) {
		if i.Type == TypeURL {
			callback(i)
		} else {
			req, _ := http.NewRequest(`GET`, i.Loc.String(), nil)
			req.Header.Set(`User-Agent`, `Mozilla/5.0 (SiteMap-Reader)`)
			resp, _ := client.Do(req)
			if resp == nil {
				return
			}

			_ = ParseResponse(resp, CallbackWithClient(client, callback))
			_ = resp.Body.Close()
		}
	}
}

func fillItem(item *Item, tag, value string) {
	if tag == `loc` {
		item.Loc, _ = url.Parse(value)
	}

	if tag == `changefreq` {
		item.ChangeFreq = value
	}

	if tag == `priority` {
		item.Priority, _ = strconv.ParseFloat(value, 64)
	}

	if tag == `lastmod` {
		var err error
		item.LastMod, err = time.Parse(`2006-01-02T15:04:05Z07:00`, value)
		if err != nil {
			item.LastMod, _ = time.Parse(`2006-01-02`, value)
		}
	}
}

func normalizeURL(loc *url.URL, baseURL *url.URL) {
	// example source link:
	// /page?abc=1
	// http:///page?abc=1
	// localhost/page?abc=1
	if loc.Host == `localhost` || loc.Host == `` {
		loc.Host = baseURL.Host
	}
	// example source link: example.com/page?abc=1
	if loc.Scheme == `` {
		loc.Scheme = baseURL.Scheme
	}
}

func scanTag(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '<'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
