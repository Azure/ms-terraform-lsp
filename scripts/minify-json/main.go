package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: minify-json <dir|file> [...]\n")
		os.Exit(1)
	}

	for _, arg := range os.Args[1:] {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", arg, err)
			os.Exit(1)
		}
		if info.IsDir() {
			if err := minifyDir(arg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := minifyFile(arg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func minifyDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			if err := minifyFile(path); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
		return nil
	})
}

func minifyFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var v json.RawMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	compact, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if len(compact) == len(data) {
		return nil
	}

	return os.WriteFile(path, compact, 0644)
}
