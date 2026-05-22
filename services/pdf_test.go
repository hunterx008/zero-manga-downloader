package services

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePDFFileName(t *testing.T) {
	cases := []struct {
		template string
		chapter  string
		want     string
	}{
		{"", "12", "12.pdf"},
		{"", " 第1话 ", "第1话.pdf"},
		{"MyComic_{chapter}", "3", "MyComic_3.pdf"},
		{"Vol-{chapter}.pdf", "10", "Vol-10.pdf"},
		{"nested/{chapter}.pdf", "1", "1.pdf"},
	}
	for _, tc := range cases {
		got := ResolvePDFFileName(tc.template, tc.chapter)
		if got != tc.want {
			t.Errorf("ResolvePDFFileName(%q, %q) = %q, want %q", tc.template, tc.chapter, got, tc.want)
		}
	}
}

func TestBuildPdfFromImagePaths_smoke(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "a.png")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out.pdf")
	if err := BuildPdfFromImagePaths([]string{imgPath}, out); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() < 200 {
		t.Fatalf("pdf unexpectedly small: %d bytes", st.Size())
	}
}
