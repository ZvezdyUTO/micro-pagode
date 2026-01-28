package flusher

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func FlushJSONAtomic(dir string, payload any) error {
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(payload, "", "  ") // 将之前存储的亏惨转换为 json 格式的字节流
	if err != nil {
		return err
	}

	dst := filepath.Join(dir, "monitor.json")
	tmp := filepath.Join(dir, ".monitor.json.tmp")

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
