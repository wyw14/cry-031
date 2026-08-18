package service

import (
	"context"
	"strings"
	"testing"
)

func TestLocalFileStoreRejectsTraversalAndOversize(t *testing.T) {
	store := LocalFileStore{Root: t.TempDir(), MaxBytes: 4}
	if _, err := store.Save(context.Background(), "user-1", "../secret.txt", strings.NewReader("test"), 4); err == nil {
		t.Fatal("expected unsafe filename error")
	}
	if _, err := store.Save(context.Background(), "user-1", "photo.png", strings.NewReader("12345"), 5); err == nil {
		t.Fatal("expected size error")
	}
}
