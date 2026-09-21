package staffseed

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func cryptoRand() io.Reader { return crand.Reader }

func equalJSON(t *testing.T, a, b any) bool {
	t.Helper()
	ja, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	jb, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return string(ja) == string(jb)
}

// fmtAll memformat nilai dengan semua verb umum untuk memastikan tidak ada yang membocorkan rahasia.
func fmtAll(v any) string {
	return fmt.Sprintf("%v | %+v | %#v | %s", v, v, v, v)
}
