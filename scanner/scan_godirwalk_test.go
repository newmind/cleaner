package scanner

import (
	"testing"
)

func TestGoFileWalk(t *testing.T) {
	root := "../test"
	files, err := GoFileWalk(root)
	if err != nil {
		t.Error(err)
	}

	for _, f := range files {
		t.Logf("%#v", *f)
		//fmt.Printf("%#v", *f)
	}
}
