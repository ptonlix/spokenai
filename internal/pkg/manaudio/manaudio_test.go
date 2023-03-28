package manaudio

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestBackupRecordAudio(t *testing.T) {
	log := &zap.Logger{}
	audioManager, _ := NewManager(log,
		WithRoleId("1"),
		WithUserId("test"),
		WithDataRecordDir("./"),
	)
	audioManager.count += 1
	if f, err := os.Create(audioManager.GetRecordAudio()); err != nil {
		t.Log(err)
	} else {
		f.Close()
	}
	t.Log(audioManager.GetRecordAudio())
	t.Log(audioManager.BackupRecordAudio())

}
