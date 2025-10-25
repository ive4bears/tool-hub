package hub

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Model is the main struct for tool hub operations.
type Model struct {
	ctx     context.Context
	HomeDir string `json:"homeDir"`
}

var model Model = func() Model {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return Model{
		HomeDir: homeDir,
	}
}()

// ExposeModel returns the singleton instance of Model.
func ExposeModel() *Model {
	return &model
}

// Dirs represents a map of directory paths.
type Dirs struct {
	Home string `json:"home"`
	Temp string `json:"temp"`
}

// GetDirs returns the Local directories path.
func (m *Model) GetDirs() Dirs {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	temp := os.TempDir()
	return Dirs{
		Home: home,
		Temp: temp,
	}
}

// CreateCommandLineTool creates a new CommandLineTool in the database.
func (m *Model) CreateCommandLineTool(cmdTool CommandLineTool) (id uint, er string) {
	tool := Tool{
		BaseTool: BaseTool{
			Name:        cmdTool.Name,
			Description: cmdTool.Description,
			Parameters:  cmdTool.Parameters,
			Type:        cmdTool.Type,
			LogLifeSpan: cmdTool.LogLifeSpan,
		},
	}
	tx := db.Create(&tool)
	if tx.Error != nil {
		err := fmt.Errorf("failed to create tool: %w", tx.Error)
		if m.ctx != nil {
			runtime.LogError(m.ctx, err.Error())
		}
		return 0, err.Error()
	}
	cmdTool.ID = tool.ID
	tx = db.Create(&cmdTool)
	if tx.Error != nil {
		err := fmt.Errorf("failed to create command line tool: %w", tx.Error)
		if m.ctx != nil {
			runtime.LogError(m.ctx, err.Error())
		}
		return 0, err.Error()
	}
	return tool.ID, ""
}

// UpdateCommandLineTool updates an existing CommandLineTool in the database.
func (m *Model) UpdateCommandLineTool(cmdTool CommandLineTool) bool {
	tx := db.Save(&cmdTool)
	if tx.Error != nil {
		if m.ctx != nil {
			runtime.LogErrorf(m.ctx, "failed to save command line tool: %v", tx.Error)
		}
		return false
	}
	return true
}

// UpdateTool updates an existing Tool in the database.
func (m *Model) UpdateTool(tool Tool) bool {
	tx := db.Save(&tool)
	if tx.Error != nil {
		if m.ctx != nil {
			runtime.LogErrorf(m.ctx, "failed to save tool: %v", tx.Error)
		}
		return false
	}
	return true
}

// CreateServiceTool creates a new ServiceTool in the database.
func (m *Model) CreateServiceTool(serviceTool ServiceTool) uint {
	tool := Tool{
		BaseTool: BaseTool{
			Name:        serviceTool.Name,
			Description: serviceTool.Description,
			Parameters:  serviceTool.Parameters,
			Type:        serviceTool.Type,
			LogLifeSpan: serviceTool.LogLifeSpan,
		},
	}
	tx := db.Create(&tool)
	if tx.Error != nil {
		if m.ctx != nil {
			runtime.LogErrorf(m.ctx, "failed to create tool: %v", tx.Error)
		}
		return 0
	}
	serviceTool.ID = tool.ID
	tx = db.Create(&serviceTool)
	if tx.Error != nil {
		if m.ctx != nil {
			runtime.LogErrorf(m.ctx, "failed to create service tool: %v", tx.Error)
		}
		return 0
	}
	return tool.ID
}

// UpdateServiceTool updates an existing ServiceTool in the database.
func (m *Model) UpdateServiceTool(serviceTool ServiceTool) bool {
	tx := db.Save(&serviceTool)
	if tx.Error != nil {
		if m.ctx != nil {
			runtime.LogErrorf(m.ctx, "failed to save service tool: %v", tx.Error)
		}
		return false
	}
	return true
}
