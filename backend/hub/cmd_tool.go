package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"tool-hub/backend/hub/cmd"
	"tool-hub/backend/hub/fifo"

	"gorm.io/gorm"
)

// CmdToolBody represents the request body for command line tool execution.
type CmdToolBody struct {
	Args       map[string]string `json:"args"`
	Stdin      string            `json:"stdin"`
	Env        map[string]string `json:"env"`
	WD         string            `json:"working_dir"`
	Timeout    string            `json:"timeout"`
	CallerName string            `json:"caller"`
}

func cmdToolHandler(cmdTool CommandLineTool, r *http.Request, w http.ResponseWriter) {
	var (
		body CmdToolBody
		bs   []byte
		err  error
	)
	callingLog := CallingLog{
		CalleeID:   cmdTool.ID,
		CalleeType: "command_line",
	}
	bs, err = io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("failed to read request body: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	callingLog.Input = string(bs)
	if err = json.Unmarshal(bs, &body); err != nil {
		err = fmt.Errorf("invalid request body: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var (
		res  cmd.Result
		code int
	)
	if cmdTool.ConcurrencyGroup != nil {
		group := cmdTool.ConcurrencyGroup
		err = fifo.DefaultGroupLimiter.Acquire(r.Context(), group.ID, group.MaxConcurrent)
		if err != nil {
			err = fmt.Errorf("failed to acquire concurrency group semaphore: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer fifo.DefaultGroupLimiter.Release(group.ID)
	}
	res, code, err = runCmdTool(r.Context(), cmdTool, body, callingLog)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	w.Write(res.Stdout)
	w.WriteHeader(http.StatusOK)
}

// RunCmdToolForTesting is used for testing purposes to allow injection of the runCmdTool function.
var RunCmdToolForTesting = runCmdTool

func runCmdTool(ctx context.Context, cmdTool CommandLineTool, body CmdToolBody, callingLog CallingLog) (res cmd.Result, code int, err error) {
	defer func() {
		if err != nil && db != nil {
			callingLog.Error = err.Error()
			dbRes := db.Create(&callingLog)
			if dbRes.Error != nil {
				log.Printf("error logging calling log: %v", dbRes.Error)
			}
		}
	}()

	// Handle caller tool lookup
	if body.CallerName != "" {
		if db == nil {
			err, code = fmt.Errorf("database not initialized"), http.StatusInternalServerError
			return
		}
		var tool Tool
		tool, err = gorm.G[Tool](db).Where("name = ?", body.CallerName).Take(ctx)
		if err != nil {
			err, code = fmt.Errorf("invalid request Caller tool: %w", err), http.StatusBadRequest
			return
		}
		callingLog.CallerID = &tool.ID
		callingLog.CallerType = &tool.Type
	}

	// Validate command
	if len(cmdTool.Cmd) == 0 {
		err, code = fmt.Errorf("command is empty"), http.StatusInternalServerError
		return
	}

	// Prepare environment variables
	env := make(map[string]string)
	maps.Copy(env, cmdTool.Env)
	if body.Env != nil {
		maps.Copy(env, body.Env)
	}

	// Prepare commands with argument substitution
	commands := slices.Clone(cmdTool.Cmd)
	if len(commands) == 3 && strings.HasSuffix(commands[0], "sh") && commands[1] == "-c" {
		// Handle bash -c pipeline commands with word boundary matching
		for k, v := range body.Args {
			// Use word boundary \b to match whole variable names
			commands[2] = regexp.MustCompile(`\$`+k+`\b`).ReplaceAllString(commands[2], v)
		}
	} else {
		// Handle regular commands with positional arguments
		for i, arg := range commands {
			for k, v := range body.Args {
				if arg == "$"+k {
					commands[i] = v
				}
			}
		}
	}

	// Parse timeout duration
	var timeout time.Duration
	if body.Timeout != "" {
		timeout, err = time.ParseDuration(body.Timeout)
		if err != nil {
			err, code = fmt.Errorf("invalid timeout: %w", err), http.StatusBadRequest
			return
		}
	} else if cmdTool.Timeout != "" {
		timeout, err = time.ParseDuration(cmdTool.Timeout)
		if err != nil {
			err, code = fmt.Errorf("invalid timeout config: %w", err), http.StatusInternalServerError
			return
		}
	}

	// Execute command
	options := cmd.Options{
		Cwd:     cmdTool.WD,
		Env:     env,
		Stdin:   bytes.NewReader([]byte(body.Stdin)),
		Timeout: timeout,
	}
	res, err = cmd.Run(ctx, options, commands...)
	if err != nil {
		err, code = fmt.Errorf("command execution failed: %w", err), http.StatusInternalServerError
		return
	}
	return
}
