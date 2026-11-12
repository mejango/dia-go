package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOnlyAndExplicitPreparation(t *testing.T) {
	config := filepath.Join("..", "..", "deploy", "examples", "gcp-postgres.json")
	var out, stderr bytes.Buffer
	if err := run([]string{"--config", config}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if out.String() != "configuration valid\n" {
		t.Fatal(out.String())
	}
	directory := filepath.Join(t.TempDir(), "prepared")
	if err := run([]string{"--config", config, "--out", directory}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "cloud-run.json")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{}, {"--execute"}, {"--config", config, "extra"}} {
		if err := run(args, &out, &stderr); err == nil {
			t.Fatal("accepted invalid flags", args)
		}
	}
}

func TestInvalidApplicationNeverWritesOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	raw, err := os.ReadFile(filepath.Join("..", "..", "deploy", "examples", "gcp-postgres.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err = json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	delete(config["secrets"].(map[string]any), "DIA_HMAC_KEY")
	raw, err = json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	outdir := filepath.Join(dir, "must-not-exist")
	var out, stderr bytes.Buffer
	if err := run([]string{"--config", path, "--out", outdir}, &out, &stderr); err == nil {
		t.Fatal("accepted invalid app")
	}
	if _, err := os.Stat(outdir); !os.IsNotExist(err) {
		t.Fatal("invalid configuration created artifacts", err)
	}
}
