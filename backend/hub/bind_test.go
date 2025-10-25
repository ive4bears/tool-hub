package hub

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) *gorm.DB {
	// Create a temporary database file
	dbFile := "test_hub.db"

	// Remove existing test database if it exists
	os.Remove(dbFile)

	testDB, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	assert.NoError(t, err, "failed to open test database")

	// Auto migrate all models
	err = testDB.AutoMigrate(models...)
	assert.NoError(t, err, "failed to migrate test database")

	return testDB
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, testDB *gorm.DB) {
	sqlDB, err := testDB.DB()
	if err == nil {
		sqlDB.Close()
	}
	os.Remove("test_hub.db")
}

func TestCreateCommandLineTool(t *testing.T) {
	testDB := setupTestDB(t)
	defer teardownTestDB(t, testDB)

	// Set the global db to test db
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()

	// Initialize model context
	model.ctx = context.Background()

	tests := []struct {
		name        string
		cmdTool     CommandLineTool
		expectID    bool
		description string
	}{
		{
			name: "simple echo command",
			cmdTool: CommandLineTool{
				BaseTool: BaseTool{
					Name:        "echo_tool",
					Description: "Simple echo command",
					Parameters:  `{"type":"object","properties":{"message":{"type":"string"}}}`,
					Type:        "command_line",
					LogLifeSpan: "24h",
				},
				WD:      "/tmp",
				Cmd:     []string{"echo", "$message"},
				Timeout: "5s",
				Status:  "ready",
			},
			expectID:    true,
			description: "Should create a simple command line tool with positional arguments",
		},
		{
			name: "bash -c pipeline command - grep and sort",
			cmdTool: CommandLineTool{
				BaseTool: BaseTool{
					Name:        "grep_sort_tool",
					Description: "Grep and sort pipeline command",
					Parameters:  `{"type":"object","properties":{"pattern":{"type":"string"},"file":{"type":"string"}}}`,
					Type:        "command_line",
					LogLifeSpan: "24h",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/bash", "-c", "cat $file | grep $pattern | sort"},
				Timeout: "10s",
				Status:  "ready",
			},
			expectID:    true,
			description: "Should create a bash pipeline command with variable substitution",
		},
		{
			name: "bash -c pipeline command - complex sqlite query",
			cmdTool: CommandLineTool{
				BaseTool: BaseTool{
					Name:        "sqlite_format_tool",
					Description: "SQLite query with formatting pipeline",
					Parameters:  `{"type":"object","properties":{"table":{"type":"string"}}}`,
					Type:        "command_line",
					LogLifeSpan: "48h",
				},
				WD:      ".",
				Cmd:     []string{"/bin/bash", "-c", "sqlite3 ./hub.db \"SELECT sql FROM sqlite_master WHERE type='table' AND name='$table'\" | pg_format"},
				Timeout: "30s",
				Status:  "ready",
			},
			expectID:    true,
			description: "Should create a complex bash pipeline with database query",
		},
		{
			name: "command with environment variables",
			cmdTool: CommandLineTool{
				BaseTool: BaseTool{
					Name:        "env_tool",
					Description: "Command with environment variables",
					Parameters:  `{"type":"object","properties":{"var":{"type":"string"}}}`,
					Type:        "command_line",
					LogLifeSpan: "12h",
				},
				WD:  "/tmp",
				Cmd: []string{"printenv", "$var"},
				Env: map[string]string{
					"TEST_VAR": "test_value",
				},
				Timeout: "5s",
				Status:  "ready",
			},
			expectID:    true,
			description: "Should create a tool with environment variables",
		},
		{
			name: "bash -c with multiple pipes",
			cmdTool: CommandLineTool{
				BaseTool: BaseTool{
					Name:        "multi_pipe_tool",
					Description: "Multiple pipe operations",
					Parameters:  `{"type":"object","properties":{"input":{"type":"string"},"count":{"type":"string"}}}`,
					Type:        "command_line",
					LogLifeSpan: "24h",
				},
				WD:      "/tmp",
				Cmd:     []string{"/bin/sh", "-c", "echo $input | tr '[:lower:]' '[:upper:]' | head -n $count"},
				Timeout: "5s",
				Status:  "ready",
			},
			expectID:    true,
			description: "Should create a bash command with multiple pipe operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the command line tool
			id, errStr := model.CreateCommandLineTool(tt.cmdTool)
			if tt.expectID {
				assert.Empty(t, errStr, "should not return an error for successful creation")
				assert.NotZero(t, id, tt.description)

				// Verify the tool was created in the database
				tool, err := gorm.G[Tool](testDB).Where("id = ?", id).First(context.Background())
				assert.NoError(t, err, "should find the created tool")
				assert.Equal(t, tt.cmdTool.Name, tool.Name)
				assert.Equal(t, tt.cmdTool.Description, tool.Description)
				assert.Equal(t, tt.cmdTool.Type, tool.Type)

				// Verify the command line tool was created
				cmdTool, err := gorm.G[CommandLineTool](testDB).Where("id = ?", id).First(context.Background())
				assert.NoError(t, err, "should find the created command line tool")
				assert.Equal(t, tt.cmdTool.WD, cmdTool.WD)
				assert.Equal(t, tt.cmdTool.Cmd, cmdTool.Cmd)
				assert.Equal(t, tt.cmdTool.Timeout, cmdTool.Timeout)

				// Verify environment variables if present
				if tt.cmdTool.Env != nil {
					assert.Equal(t, tt.cmdTool.Env, cmdTool.Env)
				}
			} else {
				assert.Zero(t, id, "should return 0 for failed creation")
			}
		})
	}
}

func TestCreateCommandLineTool_DuplicateName(t *testing.T) {
	testDB := setupTestDB(t)
	defer teardownTestDB(t, testDB)

	originalDB := db
	db = testDB
	defer func() { db = originalDB }()

	// Save original context and clear it to avoid runtime.LogErrorf call
	originalCtx := model.ctx
	model.ctx = nil
	defer func() { model.ctx = originalCtx }()

	cmdTool := CommandLineTool{
		BaseTool: BaseTool{
			Name:        "duplicate_tool",
			Description: "Test duplicate",
			Parameters:  `{}`,
			Type:        "command_line",
			LogLifeSpan: "24h",
		},
		WD:      "/tmp",
		Cmd:     []string{"echo", "test"},
		Timeout: "5s",
		Status:  "ready",
	}

	// First creation should succeed
	id1, errStr1 := model.CreateCommandLineTool(cmdTool)
	assert.NotZero(t, id1, "first creation should succeed")
	assert.Empty(t, errStr1, "should not return an error for successful creation")

	// Second creation with same name should fail due to unique constraint
	id2, errStr2 := model.CreateCommandLineTool(cmdTool)
	fmt.Println(errStr2)
	assert.Zero(t, id2, "second creation with duplicate name should fail")
	assert.Equal(t, "failed to create tool: UNIQUE constraint failed: tools.name", errStr2)
}
