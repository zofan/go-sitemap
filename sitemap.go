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
	_ = iota
	// TypeSitemap is element type of sitemap index
	TypeSitemap
	// TypeURL is element type of sitemap
	TypeURL
)

// ErrWrongContentType is the error of wrong content type
var ErrWrongContentType = errors.New(`wrong content type`)

// Item is structure of the element sitemap
type Item struct {
	Loc        *url.URL
	ChangeFreq string
	LastMod    time.Time
	Priority   float64
	Type       int
}

func ParseStreamXML(stream io.Reader, callback func(*Item)) error {
	scanner := bufio.NewScanner(stream)
	scanner.Split(scanTag)

	var isOpen bool
	var lastTag string
	var valueBuffer string
	var kv = map[string]string{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		split := strings.Split(line, `>`)
		split[0] = strings.ToLower(split[0])

		if isOpen {
			if len(split) != 2 {
				valueBuffer += line
				continue
			}

			if split[0] == `/sitemap` || split[0] == `/url` {
				isOpen = false
				if item, err := parseXMLItem(kv); err == nil {
					callback(item)
				}
			} else if split[0][0:1] != `/` {
				lastTag = split[0]
				valueBuffer += split[1]
			} else if lastTag != `` {
				kv[lastTag] = valueBuffer
				lastTag = ``
				valueBuffer = ``
			}
		} else if split[0] == `sitemap` || split[0] == `url` {
			isOpen = true
			kv = map[string]string{`type`: split[0]}
		}
	}

	return nil
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
	} else if strings.Contains(mime, `plain`) || strings.Contains(baseName, `.txt`) {
		ParseStreamPlain(reader, wrapCallback)
	} else {
		return ErrWrongContentType
	}

	return nil
}

// CallbackWithClient can be use for deep parsing nested locations to another sitemaps
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

func parseXMLItem(kv map[string]string) (*Item, error) {
	var err error
	item := &Item{}

	if loc, ok := kv[`loc`]; ok {
		item.Loc, err = url.Parse(loc)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New(`"loc" tag undefined`)
	}

	if changeFreq, ok := kv[`changefreq`]; ok {
		item.ChangeFreq = changeFreq
	}

	if priority, ok := kv[`priority`]; ok {
		item.Priority, _ = strconv.ParseFloat(priority, 64)
	}

	if lastMod, ok := kv[`lastmod`]; ok {
		item.LastMod, err = time.Parse(`2006-01-02T15:04:05Z07:00`, lastMod)
		if err != nil {
			item.LastMod, _ = time.Parse(`2006-01-02`, lastMod)
		}
	}

	switch kv[`type`] {
	case `sitemap`:
		item.Type = TypeSitemap
	case `url`:
		item.Type = TypeURL
	}

	return item, nil
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
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
