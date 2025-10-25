package hub

// BaseModel provides common fields for all database models.
type BaseModel struct {
	ID        uint  `json:"id" gorm:"primarykey"`
	CreatedAt int64 `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

// BaseTool represents tool fields exposed to LLM.
type BaseTool struct {
	Name        string `json:"name" gorm:"uniqueIndex"`
	Description string `json:"description"`
	Parameters  string `json:"parameters"`  // JSON schema for parameters
	Type        string `json:"type"`        // "command_line", "http", "service"
	LogLifeSpan string `json:"logLifeSpan"` // log life span e.g. "24h", "7d"
}

// Tool represents a tool exposed to Large Language Models (LLMs).
type Tool struct {
	BaseModel
	BaseTool
}

// CmdToolTestcase represents a test case for a command line tool.
// test with func `runCmdTool()`
type CmdToolTestcase struct {
	Input     CmdToolBody `json:"input"`
	Expect    string      `json:"expect"`
	MatchType string      `json:"matchType"` // "contains", "equals", "regex",  "prefix", "suffix"
}

// CommandLineTool represents a command line tool that can be invoked via HTTP POST request
// CommandLineTool.ID is the same as Tool.ID
type CommandLineTool struct {
	BaseModel
	BaseTool
	WD  string   `json:"wd"`
	Cmd []string `json:"cmd" gorm:"serializer:json"` // shell commands e.g. ["node", "dist/index.js" "--model" "$model"]
	// or ["/bin/bash", "-c", "sqlite3 ./hub.db \"SELECT sql FROM sqlite_master WHERE type='table' AND name='dependencies'\"| pg_format > hub.sql"]
	Env                map[string]string `json:"env" gorm:"serializer:json"` // environment variables in JSON format
	Timeout            string            `json:"timeout"`
	IsStream           bool              `json:"isStream"`
	Dependencies       []Dependency      `json:"dependencies" gorm:"many2many:command_line_tool_dependencies;"`
	Error              string            `json:"error"`  // last error message
	Status             string            `json:"status"` // "active", "error", "ready"
	Testcases          []CmdToolTestcase `json:"testcases" gorm:"serializer:json"`
	ConcurrencyGroupID *uint             `json:"concurrencyGroupID"` // nullable, foreign key to ConcurrencyGroup.ID
	ConcurrencyGroup   *ConcurrencyGroup `json:"concurrencyGroup" gorm:"foreignKey:ConcurrencyGroupID"`
}

// HTTPToolTestcase represents a test case for an HTTP tool.
// test with curl cmd tool
type HTTPToolTestcase struct {
	Curl      string `json:"curl"`
	Expect    string `json:"expect"`
	MatchType string `json:"matchType"` // "contains", "equals", "regex",  "prefix", "suffix"
}

// HTTPTool represents a tool that makes HTTP calls to external endpoints.
// HTTPTool.ID is the same as Tool.ID
// test with curl cmd tool
type HTTPTool struct {
	BaseModel
	BaseTool
	Endpoint           string             `json:"endpoint"` // HTTP endpoint URL
	Method             string             `json:"method"`   // HTTP method e.g. "POST", "GET"
	Query              string             `json:"query"`    // URL query parameters in JSON format
	Headers            string             `json:"headers"`  // HTTP headers in JSON format
	Body               string             `json:"body"`     // request body template
	Timeout            string             `json:"timeout"`  // e.g. "30s"
	Error              string             `json:"error"`    // last error message
	Status             string             `json:"status"`   // "active", "error", "ready"
	Testcases          []HTTPToolTestcase `json:"testcases" gorm:"serializer:json"`
	ConcurrencyGroupID *uint              `json:"concurrencyGroupID"` // nullable, foreign key to ConcurrencyGroup.ID
	ConcurrencyGroup   *ConcurrencyGroup  `json:"concurrencyGroup" gorm:"foreignKey:ConcurrencyGroupID"`
}

// ServiceTool represents a service-based tool with its status and error information.
// ServiceTool.ID is the same as Tool.ID
type ServiceTool struct {
	BaseModel
	BaseTool
	StartCmd           string            `json:"startCmd"`           // command to start the service
	Error              string            `json:"error"`              // last error message
	Status             string            `json:"status"`             // "active", "error", "ready"
	ConcurrencyGroupID *uint             `json:"concurrencyGroupID"` // nullable, foreign key to ConcurrencyGroup.ID
	ConcurrencyGroup   *ConcurrencyGroup `json:"concurrencyGroup" gorm:"foreignKey:ConcurrencyGroupID"`
}

// CallingLog represents a record of a tool invocation
// caller calls callee with input and output
type CallingLog struct {
	BaseModel
	CallerID   *uint   `json:"callerID"`   // nullable, if null means external caller, caller is not in our ecosystem
	CallerType *string `json:"callerType"` // "http", "command", "service"
	CalleeID   uint    `json:"calleeID"`
	CalleeType string  `json:"calleeType"`
	Input      string  `json:"input"`
	Output     string  `json:"output"`
	Error      string  `json:"error"`
	Duration   string  `json:"duration"`
}

// TestcaseForDependency represents a test case for a dependency.
type TestcaseForDependency struct {
	Cmd       []string `json:"cmd"`
	Expect    string   `json:"expect"`
	MatchType string   `json:"matchType"` // "contains", "equals", "regex",  "prefix", "suffix"
}

// Dependency represents an external dependency required by a CommandLineTool
type Dependency struct {
	BaseModel
	Name        string                  `json:"name" gorm:"uniqueIndex"`
	Description string                  `json:"description"`
	Doc         string                  `json:"doc"`
	URL         string                  `json:"url"`
	InstallCmd  string                  `json:"installCmd"`                       // command to install e.g. "brew install python"
	Testcases   []TestcaseForDependency `json:"testcases" gorm:"serializer:json"` // json format e.g. [{ "cmd": ["python", "--version"], "expect": "Python 3.", "matchType": "contains"}]
}

// ConcurrencyGroup represents a group of tools with concurrency control.
type ConcurrencyGroup struct {
	BaseModel
	Name          string `json:"name" gorm:"uniqueIndex"`
	MaxConcurrent uint   `json:"maxConcurrent"` // max concurrent tasks in the group
}
