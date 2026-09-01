package safety

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmNonEmptyDefaultsFalse(t *testing.T) {
	ok, err := ConfirmNonEmpty(strings.NewReader("\n"), &bytes.Buffer{}, "benchmark", 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ConfirmNonEmpty default = true, want false")
	}
}

func TestConfirmNonEmptyYes(t *testing.T) {
	ok, err := ConfirmNonEmpty(strings.NewReader("yes\n"), &bytes.Buffer{}, "benchmark", 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ConfirmNonEmpty yes = false, want true")
	}
}

func TestConfirmNonEmptyForce(t *testing.T) {
	ok, err := ConfirmNonEmpty(strings.NewReader(""), &bytes.Buffer{}, "benchmark", 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ConfirmNonEmpty force = false, want true")
	}
}
