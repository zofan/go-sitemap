package sitemap

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

// nolint
var testURL, _ = url.Parse(`http://example.com/sitemap.xml`)

// nolint
var baseHTTPResponse = http.Response{
	StatusCode: 200,
	Header:     map[string][]string{`Content-Type`: {`text/xml`}},
	Body: ioutil.NopCloser(
		bytes.NewBufferString(`<sitemap><url><loc>http://example.com/home</loc></url></sitemap>`)),
	Request: &http.Request{URL: testURL},
}

func TestParseURL(t *testing.T) {
	xmlFile, _ := os.Open(`./sitemap.xml`)
	var items []*Item
	err := ParseStreamXML(xmlFile, func(i *Item) {
		items = append(items, i)
	})

	if err != nil {
		t.Error(err)
	}

	if len(items) != 4 {
		t.Error(`wrong items count, expected 4`)
	}

	cases := []struct {
		Loc        string
		ChangeFreq string
		LastMod    string
		Priority   float64
		Type       string
	}{
		{
			Loc:        `https://example.com/index0.xml.gz?x=☂`,
			Type:       TypeSitemap,
			ChangeFreq: ``,
			Priority:   0,
			LastMod:    `0001-01-01T00:00:00Z`,
		},
		{
			Loc:        `https://example.com/index1.xml.gz`,
			Type:       TypeSitemap,
			ChangeFreq: `yearly`,
			Priority:   0,
			LastMod:    `0001-01-01T00:00:00Z`,
		},
		{
			Loc:        `https://example.com/page/1`,
			Type:       TypeURL,
			ChangeFreq: `yearly`,
			Priority:   0,
			LastMod:    `2019-05-23T00:00:00Z`,
		},
		{
			Loc:        `https://example.com/page/2`,
			Type:       TypeURL,
			ChangeFreq: `weekly`,
			Priority:   1,
			LastMod:    `2019-02-02T14:05:06+06:44`,
		},
	}

	for i, c := range cases {
		if items[i].Loc.String() != c.Loc {
			t.Errorf(`wrong item[%d] Loc`, i)
		}
		if items[i].Type != c.Type {
			t.Errorf(`wrong item[%d] Type`, i)
		}
		if items[i].ChangeFreq != c.ChangeFreq {
			t.Errorf(`wrong item[%d] ChangeFreq`, i)
		}
		if items[i].Priority != c.Priority {
			t.Errorf(`wrong item[%d] Priority`, i)
		}
		lm := items[i].LastMod.Format(`2006-01-02T15:04:05Z07:00`)
		if lm != c.LastMod {
			t.Errorf(`wrong item[%d] LastMod, actual %#v`, i, lm)
		}
	}
}

func TestParseResponseOK(t *testing.T) {
	resp := baseHTTPResponse

	var items []*Item
	err := ParseResponse(&resp, func(i *Item) {
		items = append(items, i)
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(items) == 0 {
		t.Error(`Empty items`)
		return
	}
	if items[0].Loc.Path != `/home` {
		t.Error(`Wrong location path`)
	}
}

func TestParseResponseWrongType(t *testing.T) {
	resp := baseHTTPResponse
	resp.Header = map[string][]string{`Content-Type`: {`text/html`}}
	resp.Request.URL.Path = `/index.html`

	err := ParseResponse(&resp, func(i *Item) {})
	if err != ErrWrongContentType {
		t.Error(err)
		return
	}
}
