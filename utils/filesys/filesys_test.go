package filesys

import (
	"os"
	"testing"
)

func TestAtomicCopyFileContent(t *testing.T) {
	srcFile, err := os.CreateTemp("", "src")
	if err != nil {
		t.Fatalf("Failed to create src file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	destFile, err := os.CreateTemp("", "dest")
	if err != nil {
		t.Fatalf("Failed to create dest file: %v", err)
	}
	defer os.Remove(destFile.Name())

	srcContent := []byte("This is a test.")
	if _, err := srcFile.Write(srcContent); err != nil {
		t.Fatalf("Failed to write to src file: %v", err)
	}
	srcFile.Close()

	if err := AtomicCopyFileContent(srcFile.Name(), destFile.Name()); err != nil {
		t.Errorf("Failed to copy file content atomically: %v", err)
	}

	destContent, err := os.ReadFile(destFile.Name())
	if err != nil {
		t.Errorf("Failed to read dest file: %v", err)
	}
	if string(destContent) != string(srcContent) {
		t.Errorf("Content mismatch: got %v, want %v", string(destContent), string(srcContent))
	}
}
