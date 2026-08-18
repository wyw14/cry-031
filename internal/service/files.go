package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type LocalFileStore struct {
	Root     string
	MaxBytes int64
}

func (s LocalFileStore) Save(ctx context.Context, ownerID, filename string, reader io.Reader, size int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if size <= 0 || size > s.MaxBytes {
		return "", errors.New("file size is outside the allowed range")
	}
	base := filepath.Base(filename)
	if base != filename || strings.ContainsAny(base, `/\`) {
		return "", errors.New("unsafe filename")
	}
	ext := strings.ToLower(filepath.Ext(base))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	if !allowed[ext] || mime.TypeByExtension(ext) == "" {
		return "", errors.New("unsupported attachment type")
	}
	dir := filepath.Join(s.Root, ownerID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	target := filepath.Join(dir, base)
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return "", fmt.Errorf("create attachment: %w", err)
	}
	defer file.Close()
	written, err := io.CopyN(file, reader, size)
	if err != nil && !errors.Is(err, io.EOF) {
		_ = os.Remove(target)
		return "", err
	}
	if written != size {
		_ = os.Remove(target)
		return "", errors.New("attachment length mismatch")
	}
	return filepath.ToSlash(filepath.Join(ownerID, base)), nil
}
