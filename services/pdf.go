package services

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ResolvePDFFileName returns the PDF filename (not a full path) for a chapter.
// If template is empty, the stem is the chapter name. Placeholder {chapter} is
// replaced with the chapter label (same numbering/name as the downloaded folder).
func ResolvePDFFileName(template, chapter string) string {
	chapter = strings.TrimSpace(chapter)
	base := chapter
	if template != "" {
		base = strings.ReplaceAll(template, "{chapter}", chapter)
	}
	base = filepath.Base(base)
	if !strings.HasSuffix(strings.ToLower(base), ".pdf") {
		base += ".pdf"
	}
	return base
}

// BuildPdfFromImagePaths writes one PDF with one page per image, in the given order.
func BuildPdfFromImagePaths(imagePaths []string, pdfPath string) error {
	var valid []string
	for _, p := range imagePaths {
		if _, err := os.Stat(p); err == nil {
			valid = append(valid, p)
		}
	}
	if len(valid) == 0 {
		return fmt.Errorf("no image files on disk for pdf: %s", pdfPath)
	}

	pdf := gofpdf.New("P", "pt", "", "")
	for i, imgPath := range valid {
		if err := addImageAsPDFPage(pdf, imgPath, i); err != nil {
			return err
		}
		if err := pdf.Error(); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(pdfPath), 0o744); err != nil {
		return err
	}
	return pdf.OutputFileAndClose(pdfPath)
}

func addImageAsPDFPage(pdf *gofpdf.Fpdf, imgPath string, seq int) error {
	ext := strings.ToLower(filepath.Ext(imgPath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		info := pdf.RegisterImageOptions(imgPath, gofpdf.ImageOptions{ReadDpi: false})
		if err := pdf.Error(); err != nil {
			return fmt.Errorf("register image %q: %w", imgPath, err)
		}
		if info == nil {
			return fmt.Errorf("register image %q: nil info", imgPath)
		}
		w, h := info.Width(), info.Height()
		pdf.AddPageFormat("P", gofpdf.SizeType{Wd: w, Ht: h})
		pdf.ImageOptions(imgPath, 0, 0, w, h, false, gofpdf.ImageOptions{ReadDpi: false}, 0, "")
		return pdf.Error()
	default:
		return addImagePageFromDecodedPNG(pdf, imgPath, seq)
	}
}

func addImagePageFromDecodedPNG(pdf *gofpdf.Fpdf, imgPath string, seq int) error {
	data, err := os.ReadFile(imgPath)
	if err != nil {
		return err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode image %q: %w", imgPath, err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("encode png for %q: %w", imgPath, err)
	}
	name := fmt.Sprintf("__pdfimg_%d", seq)
	info := pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{ImageType: "png", ReadDpi: false}, bytes.NewReader(buf.Bytes()))
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("register decoded image %q: %w", imgPath, err)
	}
	if info == nil {
		return fmt.Errorf("register decoded image %q: nil info", imgPath)
	}
	w, h := info.Width(), info.Height()
	pdf.AddPageFormat("P", gofpdf.SizeType{Wd: w, Ht: h})
	pdf.ImageOptions(name, 0, 0, w, h, false, gofpdf.ImageOptions{ImageType: "png", ReadDpi: false}, 0, "")
	return pdf.Error()
}
