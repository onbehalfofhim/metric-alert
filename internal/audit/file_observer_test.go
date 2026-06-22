package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func TestNewFileObserver(t *testing.T) {
	log := logger.NewLogger()
	o := NewFileObserver("/tmp/audit.log", log)
	assert.Equal(t, "/tmp/audit.log", o.path)
	assert.Equal(t, log, o.logger)
}

func TestFileObserver_GetID(t *testing.T) {
	o := &FileObserver{
		path: "/tmp/audit.log",
	}

	assert.Equal(t, "file-observer-/tmp/audit.log", o.GetID())
}

func TestFileObserver_Notify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	o := &FileObserver{
		path:   path,
		logger: logger.NewLogger(),
	}

	msg := models.AuditMessage{
		IPAddr: "127.0.0.0",
	}

	o.Notify(msg)
	data, err := os.ReadFile(path)

	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestWriteToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	o := &FileObserver{
		path: path,
	}

	msg := models.AuditMessage{
		IPAddr: "127.0.0.0",
	}

	err := o.writeToFile(msg)

	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got models.AuditMessage
	err = json.Unmarshal(bytes.TrimSpace(data), &got)
	require.NoError(t, err)

	assert.Equal(t, msg, got)
}

func TestWriteToFile_Append(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	o := &FileObserver{
		path: path,
	}

	msg1 := models.AuditMessage{IPAddr: "127.0.0.0"}
	msg2 := models.AuditMessage{IPAddr: "127.0.0.1"}

	require.NoError(t, o.writeToFile(msg1))
	require.NoError(t, o.writeToFile(msg2))

	file, err := os.Open(path)
	require.NoError(t, err)
	defer func() {
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	var messages []models.AuditMessage

	for decoder.More() {
		var m models.AuditMessage
		require.NoError(t, decoder.Decode(&m))
		messages = append(messages, m)
	}

	assert.Len(t, messages, 2)
	assert.Equal(t, msg1, messages[0])
	assert.Equal(t, msg2, messages[1])
}
