package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestStoresListJSONUsesResultEnvelope(t *testing.T) {
	umbrella := t.TempDir()
	storePath := filepath.Join(umbrella, "work")
	if err := os.MkdirAll(storePath, 0755); err != nil {
		t.Fatal(err)
	}
	writeStores(t, umbrella, "schema: 2\ndefault: work\nstores:\n  work: { path: "+storePath+", kind: folio }\n")
	t.Setenv("FOLIO_HOME", umbrella)
	t.Chdir(umbrella)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := runStoresList([]string{"--json"})
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout
	defer r.Close()

	var output bytes.Buffer
	if _, err := output.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if code != dendrik.ExitOK {
		t.Fatalf("code = %d, want OK", code)
	}

	var envelope struct {
		Data []storeJSON `json:"data"`
	}
	if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
		t.Fatalf("JSON output is not a result envelope: %v\n%s", err, output.String())
	}
	if len(envelope.Data) != 1 || envelope.Data[0].Name != "work" || envelope.Data[0].Path != storePath || envelope.Data[0].Kind != "folio" {
		t.Fatalf("data = %#v, want the registered work store", envelope.Data)
	}
}
