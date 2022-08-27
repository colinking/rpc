package xpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Rel struct {
	root    string
	relpath string
	abspath string
}

func NewRel(root string, path string) (Rel, error) {
	relpath, err := filepath.Rel(root, path)
	if err != nil {
		return Rel{}, err
	}

	return Rel{
		root:    filepath.Clean(root),
		relpath: relpath,
		abspath: filepath.Clean(path),
	}, nil
}

func (p Rel) Abs() string {
	return p.abspath
}

func (p Rel) Rel() string {
	return p.relpath
}

func (p Rel) RelDirs() []string {
	dirpath, _ := filepath.Split(p.relpath)
	if len(dirpath) == 0 {
		return []string{}
	}
	components := strings.Split(dirpath[:len(dirpath)-1], "/")
	return components
}

func (p Rel) AbsDirs() []string {
	dirpath, _ := filepath.Split(p.abspath)
	if len(dirpath) == 0 {
		return []string{}
	}
	components := strings.Split(dirpath[:len(dirpath)-1], "/")
	return components
}

func (p Rel) FileName() string {
	_, f := filepath.Split(p.relpath)
	return f
}

func (p Rel) String() string {
	return fmt.Sprintf("Rel<%s, %s>", p.root, p.relpath)
}
