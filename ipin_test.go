package ipingo_test

import (
	"os"
	"testing"

	ipingo "github.com/bslizon/ipin-go"
)

func TestGetNormalizedPNG(t *testing.T) {
	b, err := ipingo.GetNormalizedPNG("/tmp/from.png")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("/tmp/to.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
}
