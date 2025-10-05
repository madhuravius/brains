package config

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLoggerDisabled(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	err := b.initLogger(false)
	assert.NoError(t, err)

	_, err = os.Stat(".brains.log")
	assert.True(t, os.IsNotExist(err), "log file must not be created when disabled")
}

func TestInitLoggerCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	err := b.initLogger(true)
	assert.NoError(t, err)

	_, err = os.Stat(".brains.log")
	assert.NoError(t, err, "log file should exist after initLogger")
}

func TestLogMessageAndGetLogContext(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	_ = b.initLogger(true)

	b.LogMessage("first line")
	b.LogMessage("second line")

	ctx := b.GetLogContext()
	assert.Contains(t, ctx, "first line")
	assert.Contains(t, ctx, "second line")
}

func TestGetLogContextFallbackToMem(t *testing.T) {
	b := &BrainsConfig{}
	b.logger.mem = []string{"first line", "second line"}

	_, err := os.Stat(".brains.log")
	assert.True(t, os.IsNotExist(err))

	ctx := b.GetLogContext()
	assert.Equal(t, "first line\nsecond line", ctx)
}

func TestPrintLogsTagConversion(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	_ = b.initLogger(true)
	b.LogMessage("[REQUEST] ask")
	b.LogMessage("[RESPONSE] answer")
	b.LogMessage("[RESPONSE FOR CODE] diff")

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	b.PrintLogs()

	_ = w.Close()
	os.Stdout = oldStdout

	var out bytes.Buffer
	_, _ = io.Copy(&out, r)

	rendered := out.String()
	assert.Contains(t, rendered, "[❓ REQUEST]")
	assert.Contains(t, rendered, "[⚡ RESPONSE]")
	assert.Contains(t, rendered, "[🧠 RESPONSE FOR CODE]")
}

func TestResetClearsLog(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	_ = b.initLogger(true)

	b.LogMessage("tmp")
	assert.NotEmpty(t, b.GetLogContext())

	assert.NoError(t, b.Reset())
	assert.Empty(t, b.GetLogContext())

	info, err := os.Stat(".brains.log")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}

func TestDisabledLoggerDoesNothing(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	b := &BrainsConfig{}
	_ = b.initLogger(false)

	b.LogMessage("nope")
	assert.Empty(t, b.GetLogContext())

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	b.PrintLogs()
	_ = w.Close()
	os.Stdout = oldStdout

	var out strings.Builder
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		out.WriteString(scanner.Text())
	}
	assert.Empty(t, strings.TrimSpace(out.String()))
}
