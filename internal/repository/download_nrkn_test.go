package repository

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNrknDownloaderConfiguration(t *testing.T) {
	t.Setenv("BUCKET_NAME", "dummy-mf-bucket")
	t.Setenv("BUCKET_DIR", "dummy-mf-prefix")
	t.Setenv("NRKN_BUCKET_NAME", "")
	t.Setenv("NRKN_BUCKET_DIR", "")
	t.Setenv("AWS_REGION", "test-region-1")
	d := NewNrknDownloader(t.TempDir())
	if d.BucketName != "" || d.BucketDir != "" {
		t.Fatal("NRKN used another source's bucket settings")
	}
	if err := d.StartMonth(t.Context(), "200001"); err == nil {
		t.Fatal("missing config accepted")
	}
	for _, month := range []string{"", "200013", "2000-01", "../001"} {
		if _, err := nrknMonthPrefix("dummy", month); err == nil {
			t.Fatal("invalid month accepted")
		}
	}
	prefix, err := nrknMonthPrefix("dummy/", "200001")
	if err != nil || prefix != "dummy/2000/01/" {
		t.Fatal("wrong month prefix")
	}
}

func TestNrknDownloaderPaginationAndPrivateFiles(t *testing.T) {
	var lists, gets int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list-type") == "2" {
			lists++
			if r.URL.Query().Get("prefix") != "dummy/2000/01/" {
				t.Error("wrong prefix")
			}
			w.Header().Set("Content-Type", "application/xml")
			if r.URL.Query().Get("continuation-token") == "" {
				fmt.Fprint(w, `<ListBucketResult><IsTruncated>true</IsTruncated><NextContinuationToken>next</NextContinuationToken><Contents><Key>dummy/2000/01/a.json</Key></Contents><Contents><Key>dummy/2000/01/ignored.txt</Key></Contents></ListBucketResult>`)
			} else {
				fmt.Fprint(w, `<ListBucketResult><IsTruncated>false</IsTruncated><Contents><Key>dummy/2000/01/nested/b.JSON</Key></Contents></ListBucketResult>`)
			}
			return
		}
		gets++
		fmt.Fprint(w, `{}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	d := &nrknDownloader{AccessKeyID: "dummy", SecretAccessKey: "dummy", Region: "us-east-1", BucketName: "dummy", BucketDir: "dummy", Endpoint: server.URL, SaveDir: dir}
	if err := d.StartMonth(t.Context(), "200001"); err != nil {
		t.Fatal(err)
	}
	if lists != 2 || gets != 2 {
		t.Fatalf("lists=%d gets=%d", lists, gets)
	}
	for _, name := range []string{"a.json", "nested/b.JSON"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("missing file or incorrect permissions")
		}
	}
}

func TestNrknDownloaderRejectsTraversalAndS3Failure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if fail {
					http.Error(w, "synthetic failure", http.StatusForbidden)
					return
				}
				if r.URL.Query().Get("list-type") == "2" {
					fmt.Fprint(w, `<ListBucketResult><IsTruncated>false</IsTruncated><Contents><Key>dummy/2000/01/../escape.json</Key></Contents></ListBucketResult>`)
					return
				}
				t.Error("unsafe object fetched")
			}))
			defer server.Close()
			d := &nrknDownloader{AccessKeyID: "dummy", SecretAccessKey: "dummy", Region: "us-east-1", BucketName: "dummy", BucketDir: "dummy", Endpoint: server.URL, SaveDir: t.TempDir()}
			err := d.StartMonth(t.Context(), "200001")
			if err == nil {
				t.Fatal("expected failure")
			}
			if !fail && !strings.Contains(err.Error(), "escapes") {
				t.Fatal("expected traversal rejection")
			}
		})
	}
}
