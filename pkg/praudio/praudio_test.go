package praudio

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAudioRecorder(t *testing.T) {
	recorder, err := NewAudioRecorder()
	if err != nil {
		t.Fatalf("Failed to create audio recorder: %v", err)
	}
	defer recorder.Stop()

	// Test reading audio data
	data := recorder.Read()
	if data == nil {
		t.Error("Failed to read audio data from recorder")
	}

	// Test stopping recorder
	recorder.Stop()
	data = recorder.Read()
	if data != nil {
		t.Error("Recorder did not stop recording")
	}
}

func TestRecordAndSaveWav(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "test.wav")
	err = RecordAndSaveWav(filename)
	if err != nil {
		t.Fatalf("Failed to record and save WAV file: %v", err)
	}

	// Test file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("WAV file not created: %v", err)
	}
}

func TestRecordAndSaveWavWithInterrupt(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "test.wav")
	err = RecordAndSaveWithInterrupt(filename)
	if err != nil {
		t.Fatalf("Failed to record and save WAV file with interrupt: %v", err)
	}

	// Test file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("WAV file not created: %v", err)
	}
}

func TestRecordAndSaveWithContext(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "test.wav")
	ctx, cancel := context.WithCancel(context.Background())
	err = RecordAndSaveWithContext(ctx, filename)
	if err != nil {
		t.Fatalf("Failed to record and save WAV file with context: %v", err)
	}

	// Test file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("WAV file not created: %v", err)
	}

	// Test cancel context
	cancel()
	err = RecordAndSaveWithContext(ctx, filename)
	if err == nil {
		t.Error("Recording did not stop after canceling context")
	}
}
