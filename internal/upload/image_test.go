package upload

import (
	"bytes"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImageExtensionJPEG(t *testing.T) {
	extension, err := imageExtension("image/jpeg")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if extension != ".jpg" {
		t.Fatalf(
			"expected .jpg, got %q",
			extension,
		)
	}
}

func TestImageExtensionPNG(t *testing.T) {
	extension, err := imageExtension("image/png")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if extension != ".png" {
		t.Fatalf(
			"expected .png, got %q",
			extension,
		)
	}
}

func TestImageExtensionGIF(t *testing.T) {
	extension, err := imageExtension("image/gif")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if extension != ".gif" {
		t.Fatalf(
			"expected .gif, got %q",
			extension,
		)
	}
}

func TestImageExtensionRejectsInvalidType(t *testing.T) {
	_, err := imageExtension("text/plain")

	if !errors.Is(
		err,
		ErrInvalidImageType,
	) {
		t.Fatalf(
			"expected ErrInvalidImageType, got %v",
			err,
		)
	}
}

func TestRandomFilenameUsesExtension(t *testing.T) {
	filename, err := randomFilename(".png")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasSuffix(
		filename,
		".png",
	) {
		t.Fatalf(
			"expected png filename, got %q",
			filename,
		)
	}
}

func TestRandomFilenameIsUnique(t *testing.T) {
	first, err := randomFilename(".jpg")
	if err != nil {
		t.Fatalf("generate first filename: %v", err)
	}

	second, err := randomFilename(".jpg")
	if err != nil {
		t.Fatalf("generate second filename: %v", err)
	}

	if first == second {
		t.Fatal("expected unique filenames")
	}
}

func TestSaveImageRejectsOversizedHeader(t *testing.T) {
	file := multipart.File(
		&memoryFile{
			Reader: bytes.NewReader(
				[]byte("fake"),
			),
		},
	)

	header := &multipart.FileHeader{
		Filename: "large.png",
		Size:     MaxImageSize + 1,
	}

	_, err := SaveImage(
		file,
		header,
	)

	if !errors.Is(
		err,
		ErrImageTooLarge,
	) {
		t.Fatalf(
			"expected ErrImageTooLarge, got %v",
			err,
		)
	}
}

func TestDeleteImageMissingFile(t *testing.T) {
	err := DeleteImage(
		"/static/uploads/does-not-exist.png",
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestDeleteImage(t *testing.T) {
	if err := os.MkdirAll(
		uploadDir,
		0755,
	); err != nil {
		t.Fatalf(
			"create upload directory: %v",
			err,
		)
	}

	filename := "delete-image-test.png"

	path := filepath.Join(
		uploadDir,
		filename,
	)

	if err := os.WriteFile(
		path,
		[]byte("test"),
		0644,
	); err != nil {
		t.Fatalf(
			"create test image: %v",
			err,
		)
	}

	t.Cleanup(func() {
		os.Remove(path)
	})

	err := DeleteImage(
		"/static/uploads/" + filename,
	)
	if err != nil {
		t.Fatalf(
			"delete image: %v",
			err,
		)
	}

	if _, err := os.Stat(path); !errors.Is(
		err,
		os.ErrNotExist,
	) {
		t.Fatal(
			"expected image to be deleted",
		)
	}
}

type memoryFile struct {
	*bytes.Reader
}

func (f *memoryFile) Close() error {
	return nil
}
