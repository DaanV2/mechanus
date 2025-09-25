package uvtt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xos"
)

func Open(filename string) (*MapData, error) {
	filename = filepath.Clean(filename)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %s -> %w", filename, err)
	}
	defer xos.CloseOrReport(f, nil)

	var r MapData
	err = json.NewDecoder(f).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
