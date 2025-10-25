package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newCallingLog(body CmdToolBody, t *testing.T) CallingLog {
	inputBytes, err := json.Marshal(body)
	assert.NoError(t, err, "should marshal CmdToolBody without error")
	return CallingLog{
		CalleeID:   0,
		CalleeType: "test",
		Input:      string(inputBytes),
	}
}

func TestRunCmdToolForTesting_SimpleCommand(t *testing.T) {
	// Use nil db since we're not testing database logging
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 1},
		BaseTool: BaseTool{
			Name: "echo_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"echo", "$message"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{
			"message": "Hello World",
		},
	}

	callingLog := newCallingLog(body, t)
	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, callingLog)

	assert.NoError(t, err, "should execute successfully")
	assert.Equal(t, 0, code, "should return status code 0")
	assert.Contains(t, string(res.Stdout), "Hello World", "should contain the message")
}

func TestRunCmdToolForTesting_BashPipelineCommand(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	tests := []struct {
		name           string
		cmdTool        CommandLineTool
		body           CmdToolBody
		expectContains string
		description    string
	}{
		{
			name: "simple pipe with grep",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 2},
				BaseTool: BaseTool{
					Name: "grep_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/bash", "-c", "echo -e 'apple\\nbanana\\ncherry' | grep $pattern"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"pattern": "banana",
				},
			},
			expectContains: "banana",
			description:    "should filter with grep in pipeline",
		},
		{
			name: "pipe with sort",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 3},
				BaseTool: BaseTool{
					Name: "sort_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/bash", "-c", "echo -e '$input' | tr ' ' '\\n' | sort"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"input": "zebra apple monkey",
				},
			},
			expectContains: "apple",
			description:    "should sort output in pipeline",
		},
		{
			name: "multiple pipes",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 4},
				BaseTool: BaseTool{
					Name: "multi_pipe_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/sh", "-c", "echo $text | tr '[:lower:]' '[:upper:]' | sed 's/WORLD/UNIVERSE/g'"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"text": "hello world",
				},
			},
			expectContains: "UNIVERSE",
			description:    "should handle multiple pipe operations",
		},
		{
			name: "pipe with wc",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 5},
				BaseTool: BaseTool{
					Name: "wc_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/bash", "-c", "echo $data | wc -w"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"data": "one two three four five",
				},
			},
			expectContains: "5",
			description:    "should count words in pipeline",
		},
		{
			name: "pipe with awk",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 6},
				BaseTool: BaseTool{
					Name: "awk_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/bash", "-c", "echo '$csv' | awk -F',' '{print $2}'"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"csv": "name,age,city\npanda,30,shanghai",
				},
			},
			expectContains: "age\n30",
			description:    "should extract column with awk in pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, code, err := RunCmdToolForTesting(context.Background(), tt.cmdTool, tt.body, newCallingLog(tt.body, t))

			assert.NoError(t, err, tt.description)
			assert.Equal(t, 0, code, "should return status code 0")
			output := strings.TrimSpace(string(res.Stdout))
			assert.Contains(t, output, tt.expectContains, tt.description)
		})
	}
}

func TestRunCmdToolForTesting_CommandsWithoutBashC(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	tests := []struct {
		name           string
		cmdTool        CommandLineTool
		body           CmdToolBody
		expectContains string
		description    string
	}{
		{
			name: "uname command",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 7},
				BaseTool: BaseTool{
					Name: "uname_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"uname", "$flag"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"flag": "-s",
				},
			},
			expectContains: "Darwin",
			description:    "should show system name",
		},
		{
			name: "cat command",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 8},
				BaseTool: BaseTool{
					Name: "cat_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"cat", "$file"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"file": "/etc/hosts",
				},
			},
			expectContains: "localhost",
			description:    "should read file contents",
		},
		{
			name: "pwd command",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 9},
				BaseTool: BaseTool{
					Name: "pwd_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"pwd"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{},
			},
			expectContains: "/tmp",
			description:    "should print working directory",
		},
		{
			name: "date command",
			cmdTool: CommandLineTool{
				BaseModel: BaseModel{ID: 10},
				BaseTool: BaseTool{
					Name: "date_test",
					Type: "command_line",
				},
				WD:      "/tmp",
				Cmd:     []string{"date", "$format"},
				Timeout: "5s",
			},
			body: CmdToolBody{
				Args: map[string]string{
					"format": "+%Y",
				},
			},
			expectContains: "202",
			description:    "should display formatted date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, code, err := RunCmdToolForTesting(context.Background(), tt.cmdTool, tt.body, newCallingLog(tt.body, t))

			assert.NoError(t, err, tt.description)
			assert.Equal(t, 0, code, "should return status code 0")
			output := strings.TrimSpace(string(res.Stdout))
			assert.Contains(t, output, tt.expectContains, tt.description)
		})
	}
}

func TestRunCmdToolForTesting_WithStdin(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 11},
		BaseTool: BaseTool{
			Name: "stdin_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"/bin/bash", "-c", "cat | grep $pattern"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{
			"pattern": "test",
		},
		Stdin: "this is a test line\nanother line\ntest again",
	}

	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.NoError(t, err, "should execute successfully")
	assert.Equal(t, 0, code, "should return status code 0")
	output := strings.TrimSpace(string(res.Stdout))
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		assert.Contains(t, line, "test", "should filter stdin content")
	}
	assert.NotContains(t, output, "another line", "should exclude non-matching lines")
}

func TestRunCmdToolForTesting_WithEnvironmentVariables(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 12},
		BaseTool: BaseTool{
			Name: "env_test",
			Type: "command_line",
		},
		WD:  "/tmp",
		Cmd: []string{"/bin/bash", "-c", "echo $MY_VAR"},
		Env: map[string]string{
			"MY_VAR": "default_value",
		},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{},
		Env: map[string]string{
			"MY_VAR": "custom_value",
		},
	}

	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.NoError(t, err, "should execute successfully")
	assert.Equal(t, 0, code, "should return status code 0")
	assert.Contains(t, string(res.Stdout), "custom_value", "should use environment variable")
}

func TestRunCmdToolForTesting_WithTimeout(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 13},
		BaseTool: BaseTool{
			Name: "timeout_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"sleep", "$duration"},
		Timeout: "100ms",
	}

	body := CmdToolBody{
		Args: map[string]string{
			"duration": "5",
		},
	}

	start := time.Now()
	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))
	duration := time.Since(start)

	assert.Error(t, err, "should timeout")
	assert.Equal(t, http.StatusInternalServerError, code, "should return 500")
	assert.Greater(t, res.Duration, 100*time.Millisecond, "should timeout quickly")
	assert.Less(t, res.Duration, 200*time.Millisecond, "should timeout quickly")
	assert.Less(t, duration, 2*time.Second, "should timeout quickly")
}

func TestRunCmdToolForTesting_InvalidCommand(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 14},
		BaseTool: BaseTool{
			Name: "invalid_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"nonexistent_command_12345"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{},
	}

	_, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.Error(t, err, "should fail for nonexistent command")
	assert.Equal(t, http.StatusInternalServerError, code, "should return 500")
}

func TestRunCmdToolForTesting_EmptyCommand(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 15},
		BaseTool: BaseTool{
			Name: "empty_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{},
	}

	_, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.Error(t, err, "should fail for empty command")
	assert.Equal(t, http.StatusInternalServerError, code, "should return 400")
}

func TestRunCmdToolForTesting_InvalidTimeout(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 16},
		BaseTool: BaseTool{
			Name: "timeout_parse_test",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"echo", "test"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args:    map[string]string{},
		Timeout: "invalid_duration",
	}

	_, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.Error(t, err, "should fail for invalid timeout")
	assert.Equal(t, http.StatusBadRequest, code, "should return 400")
	assert.Contains(t, err.Error(), "invalid timeout", "error should mention timeout")
}

func TestRunCmdToolForTesting_WorkingDirectory(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test_wd_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks for macOS /var -> /private/var
	tmpDir, err = os.Readlink(tmpDir)
	if err != nil {
		// If not a symlink, use original path
		tmpDir, _ = os.MkdirTemp("", "test_wd_*")
		defer os.RemoveAll(tmpDir)
	}

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 17},
		BaseTool: BaseTool{
			Name: "wd_test",
			Type: "command_line",
		},
		WD:      tmpDir,
		Cmd:     []string{"pwd"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{},
	}

	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.NoError(t, err, "should execute successfully")
	assert.Equal(t, 0, code, "should return status code 0")
	output := strings.TrimSpace(string(res.Stdout))
	// On macOS, pwd might return the canonical path
	assert.True(t, strings.HasSuffix(output, tmpDir) || output == tmpDir, "should execute in correct working directory")
}

func TestRunCmdToolForTesting_ComplexBashPipeline(t *testing.T) {
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	cmdTool := CommandLineTool{
		BaseModel: BaseModel{ID: 18},
		BaseTool: BaseTool{
			Name: "complex_pipeline",
			Type: "command_line",
		},
		WD:      "/tmp",
		Cmd:     []string{"/bin/bash", "-c", "echo $input | tr '[:lower:]' '[:upper:]' | rev | cut -c1-$length"},
		Timeout: "5s",
	}

	body := CmdToolBody{
		Args: map[string]string{
			"input":  "hello",
			"length": "3",
		},
	}

	res, code, err := RunCmdToolForTesting(context.Background(), cmdTool, body, newCallingLog(body, t))

	assert.NoError(t, err, "should execute successfully")
	assert.Equal(t, 0, code, "should return status code 0")
	output := strings.TrimSpace(string(res.Stdout))
	// "hello" -> "HELLO" -> "OLLEH" -> "OLL"
	assert.Equal(t, "OLL", output, "should handle complex pipeline correctly")
}
