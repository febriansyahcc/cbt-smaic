package repository

import (
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	cases := []struct {
		name, in string
		hide     string
		keep     []string
	}{
		{"url", "postgres://cbt:S3cretPw@cbt-db:5432/cbt?sslmode=disable", "S3cretPw", []string{"cbt-db:5432", "sslmode=disable"}},
		{"url tanpa sandi", "postgres://cbt@cbt-db:5432/cbt", "", []string{"postgres://cbt@cbt-db:5432/cbt"}},
		{"key value", "host=cbt-db user=cbt password=S3cretPw dbname=cbt port=5432", "S3cretPw", []string{"host=cbt-db", "dbname=cbt", "port=5432"}},
		{"key value berkutip", "host=db password='rahasia panjang' dbname=cbt", "rahasia", []string{"dbname=cbt"}},
		{"huruf besar", "HOST=db PASSWORD=S3cretPw", "S3cretPw", []string{"HOST=db"}},
		{"query", "postgres://cbt@db/cbt?password=S3cretPw&sslmode=disable", "S3cretPw", []string{"postgres://cbt@db/cbt?password=REDACTED"}}, // berlebih menyamarkan sisanya: sengaja, lebih aman
		{"dalam pesan galat", `failed to connect to postgres://cbt:S3cretPw@db:5432/cbt: dial error`, "S3cretPw", []string{"dial error"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactSecrets(tc.in)
			if tc.hide != "" && strings.Contains(got, tc.hide) {
				t.Fatalf("rahasia masih ada: %q", got)
			}
			for _, k := range tc.keep {
				if !strings.Contains(got, k) {
					t.Fatalf("bagian %q hilang: %q", k, got)
				}
			}
		})
	}
}

func TestOpenTanpaDSN(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("DSN kosong seharusnya galat")
	}
}
