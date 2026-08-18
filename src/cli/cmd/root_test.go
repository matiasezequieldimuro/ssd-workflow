package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestArgumentErrorsUseJSONEnvelope(t *testing.T) {
	root := NewRootCommand(Application{})
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--json", "status"})

	if err := executeRoot(root); err == nil {
		t.Fatal("executeRoot() error = nil")
	}
	var response JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, stdout.String())
	}
	if response.Success || response.Error == nil || response.Error.Code != "invalid_arguments" {
		t.Fatalf("response = %#v", response)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRequiredFlagErrorsUseJSONEnvelope(t *testing.T) {
	root := NewRootCommand(Application{})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--json", "start", "missing-title"})

	if err := executeRoot(root); err == nil {
		t.Fatal("executeRoot() error = nil")
	}
	var response JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v; output = %s", err, stdout.String())
	}
	if response.Error == nil || response.Error.Code != "invalid_arguments" {
		t.Fatalf("response = %#v", response)
	}
}
