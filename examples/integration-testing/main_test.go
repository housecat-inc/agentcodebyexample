package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type req struct {
	route  string
	body   string
	status int
}

type step struct {
	navigate string
	input    string
	text     string
	click    string
	element  string
}

type want struct {
	selector string
	text     string
	absent   bool
}

type op struct {
	create string
	toggle int64
	delete int64
}

func TestDB(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []op
		out  []string
	}{
		{
			name: "create lists both todos",
			in: []op{
				{create: "buy milk"},
				{create: "walk dog"},
			},
			out: []string{
				`{"_table":"todos","done":0,"id":1,"title":"buy milk"}`,
				`{"_table":"todos","done":0,"id":2,"title":"walk dog"}`,
			},
		},
		{
			name: "toggle flips done",
			in: []op{
				{create: "buy milk"},
				{toggle: 1},
			},
			out: []string{
				`{"_table":"todos","done":1,"id":1,"title":"buy milk"}`,
			},
		},
		{
			name: "delete removes the row",
			in: []op{
				{create: "buy milk"},
				{delete: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			store, err := NewDB()
			if err != nil {
				t.Fatal(err)
			}

			for _, o := range tt.in {
				switch {
				case o.create != "":
					if err := store.Create(ctx, o.create); err != nil {
						t.Fatalf("create %q: %v", o.create, err)
					}
				case o.toggle != 0:
					if err := store.Toggle(ctx, o.toggle); err != nil {
						t.Fatalf("toggle %d: %v", o.toggle, err)
					}
				case o.delete != 0:
					if err := store.Delete(ctx, o.delete); err != nil {
						t.Fatalf("delete %d: %v", o.delete, err)
					}
				}
			}

			assertRows(t, store, tt.out)
		})
	}
}

func TestHTML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []req
		out  []want
	}{
		{
			name: "create lists both todos",
			in: []req{
				{route: "POST /todos", body: "title=buy+milk"},
				{route: "POST /todos", body: "title=walk+dog"},
			},
			out: []want{
				{selector: `#list li:nth-child(1)`, text: "buy milk"},
				{selector: `#list li:nth-child(2)`, text: "walk dog"},
				{selector: `#list li:nth-child(3)`, absent: true},
			},
		},
		{
			name: "empty title rejected",
			in: []req{
				{route: "POST /todos", body: "title=", status: 400},
			},
			out: []want{
				{selector: `#list li`, absent: true},
			},
		},
		{
			name: "toggle strikes through",
			in: []req{
				{route: "POST /todos", body: "title=buy+milk"},
				{route: "POST /todos/1/toggle"},
			},
			out: []want{
				{selector: `#list li[data-id="1"] span[style*="line-through"]`, text: "buy milk"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv, _ := newTestServer(t)

			for _, r := range tt.in {
				method, path, _ := strings.Cut(r.route, " ")
				hr, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(r.body))
				hr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				res, err := http.DefaultClient.Do(hr)
				if err != nil {
					t.Fatalf("%s: %v", r.route, err)
				}
				res.Body.Close()
				want := r.status
				if want == 0 && res.StatusCode >= 400 {
					t.Fatalf("%s: status %d", r.route, res.StatusCode)
				}
				if want != 0 && res.StatusCode != want {
					t.Fatalf("%s: status %d, want %d", r.route, res.StatusCode, want)
				}
			}

			res, err := http.Get(srv.URL + "/")
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			assertHTML(t, doc, tt.out)
		})
	}
}

func TestBrowser(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping browser test in -short mode")
	}

	tests := []struct {
		name string
		in   []step
		out  []want
	}{
		{
			name: "create lists both todos",
			in: []step{
				{navigate: "/"},
				{input: `input[name="title"]`, text: "buy milk"},
				{click: `form[action="/todos"] button`},
				{input: `input[name="title"]`, text: "walk dog"},
				{click: `form[action="/todos"] button`},
				{element: `#list li[data-id="2"]`},
			},
			out: []want{
				{selector: `#list li:nth-child(1)`, text: "buy milk"},
				{selector: `#list li:nth-child(2)`, text: "walk dog"},
				{selector: `#list li:nth-child(3)`, absent: true},
			},
		},
		{
			name: "empty title rejected",
			in: []step{
				{navigate: "/"},
				{click: `form[action="/todos"] button`},
				{element: `input[name="title"]:invalid`},
			},
			out: []want{
				{selector: `#list li`, absent: true},
			},
		},
		{
			name: "toggle strikes through",
			in: []step{
				{navigate: "/"},
				{input: `input[name="title"]`, text: "buy milk"},
				{click: `form[action="/todos"] button`},
				{element: `#list li[data-id="1"]`},
				{click: `form[action="/todos/1/toggle"] button`},
				{element: `#list li[data-id="1"] span[style]`},
			},
			out: []want{
				{selector: `#list li[data-id="1"] span[style*="line-through"]`, text: "buy milk"},
			},
		},
	}

	ctrlURL, err := launcher.New().Headless(true).Launch()
	if err != nil {
		t.Fatal(err)
	}
	browser := rod.New().ControlURL(ctrlURL)
	if err := browser.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = browser.Close() })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv, _ := newTestServer(t)
			page := browser.MustIncognito().MustPage("")
			page.Timeout(10 * time.Second)

			for _, s := range tt.in {
				switch {
				case s.navigate != "":
					page.MustNavigate(srv.URL + s.navigate).MustWaitLoad()
				case s.input != "":
					page.MustElement(s.input).MustInput(s.text)
				case s.click != "":
					page.MustElement(s.click).MustClick()
					page.MustWaitLoad()
				case s.element != "":
					page.MustElement(s.element)
				}
			}

			assertPage(t, page, tt.out)

			_ = os.MkdirAll("testdata", 0o755)
			name := strings.ReplaceAll(tt.name, " ", "-") + ".png"
			page.MustScreenshotFullPage(filepath.Join("testdata", name))
		})
	}
}

func newTestServer(t *testing.T) (*httptest.Server, *DB) {
	t.Helper()
	store, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(Handler(store, log))
	t.Cleanup(srv.Close)
	return srv, store
}

func assertHTML(t *testing.T, doc *goquery.Document, wants []want) {
	t.Helper()
	for _, w := range wants {
		sel := doc.Find(w.selector)
		if w.absent {
			if sel.Length() > 0 {
				t.Errorf("%q: expected absent, got %d match(es)", w.selector, sel.Length())
			}
			continue
		}
		if sel.Length() == 0 {
			t.Errorf("%q: no match", w.selector)
			continue
		}
		if w.text != "" {
			got := strings.TrimSpace(sel.First().Text())
			if !strings.Contains(got, w.text) {
				t.Errorf("%q: text %q not in %q", w.selector, w.text, got)
			}
		}
	}
}

func assertPage(t *testing.T, page *rod.Page, wants []want) {
	t.Helper()
	for _, w := range wants {
		has, el, err := page.Has(w.selector)
		if err != nil {
			t.Errorf("%q: %v", w.selector, err)
			continue
		}
		if w.absent {
			if has {
				t.Errorf("%q: expected absent", w.selector)
			}
			continue
		}
		if !has {
			t.Errorf("%q: not found", w.selector)
			continue
		}
		if w.text != "" {
			got, err := el.Text()
			if err != nil {
				t.Errorf("%q: %v", w.selector, err)
				continue
			}
			got = strings.TrimSpace(got)
			if !strings.Contains(got, w.text) {
				t.Errorf("%q: text %q not in %q", w.selector, w.text, got)
			}
		}
	}
}

func assertRows(t *testing.T, store *DB, want []string) {
	t.Helper()
	got, err := store.DumpRows(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("rows: got %d, want %d\n got: %v\nwant: %v", len(got), len(want), got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("row %d:\n got: %s\nwant: %s", i, got[i], w)
		}
	}
}
