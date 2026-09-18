package upload

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	MaxImageSize = 20 << 20
	uploadDir    = "static/uploads"
)

var (
	ErrImageTooLarge    = errors.New("image must not exceed 20MB")
	ErrInvalidImageType = errors.New("image must be PNG, JPEG, or GIF")
)

func SaveImage(
	file multipart.File,
	header *multipart.FileHeader,
) (string, error) {
	if header.Size > MaxImageSize {
		return "", ErrImageTooLarge
	}

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read uploaded image: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, err := imageExtension(contentType)
	if err != nil {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset uploaded image: %w", err)
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("create upload directory: %w", err)
	}

	filename, err := randomFilename(extension)
	if err != nil {
		return "", err
	}

	diskPath := filepath.Join(
		uploadDir,
		filename,
	)

	destination, err := os.OpenFile(
		diskPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return "", fmt.Errorf("create uploaded image: %w", err)
	}

	defer destination.Close()

	limitedReader := io.LimitReader(
		file,
		MaxImageSize+1,
	)

	written, err := io.Copy(
		destination,
		limitedReader,
	)
	if err != nil {
		os.Remove(diskPath)

		return "", fmt.Errorf("save uploaded image: %w", err)
	}

	if written > MaxImageSize {
		os.Remove(diskPath)

		return "", ErrImageTooLarge
	}

	return "/static/uploads/" + filename, nil
}

func DeleteImage(imagePath string) error {
	if imagePath == "" {
		return nil
	}

	filename := filepath.Base(imagePath)

	if filename == "." ||
		filename == "/" ||
		filename == "" {
		return nil
	}

	diskPath := filepath.Join(
		uploadDir,
		filename,
	)

	err := os.Remove(diskPath)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("delete uploaded image: %w", err)
	}

	return nil
}

func imageExtension(
	contentType string,
) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil

	case "image/png":
		return ".png", nil

	case "image/gif":
		return ".gif", nil

	default:
		return "", ErrInvalidImageType
	}
}

func randomFilename(
	extension string,
) (string, error) {
	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate image filename: %w",
			err,
		)
	}

	return hex.EncodeToString(randomBytes) +
		extension, nil
}
