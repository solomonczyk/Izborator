package categorytree

import (
	"encoding/json"
	"fmt"
	"sync"

	_ "embed"
)

//go:embed canonical_tree_v1.json
var rawTree []byte

type Node struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Level    string `json:"level"`
	Flags    []string `json:"flags"`
	Children []Node  `json:"children"`
}

type Tree struct {
	Version    string `json:"version"`
	Levels     []string `json:"levels"`
	Categories []Node `json:"categories"`
}

var (
	loadOnce sync.Once
	loadErr  error
	loaded   Tree
)

func Load() (Tree, error) {
	loadOnce.Do(load)
	if loadErr != nil {
		return Tree{}, loadErr
	}
	return loaded, nil
}

func load() {
	if err := json.Unmarshal(rawTree, &loaded); err != nil {
		loadErr = fmt.Errorf("failed to parse canonical category tree: %w", err)
	}
}
