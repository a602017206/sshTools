package copilot

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"AHaSSHTools/internal/config"
)

const sshProbeTimeout = 5 * time.Second

func schemaToolSpecs() []ToolSpec {
	empty := json.RawMessage(`{"type":"object","properties":{}}`)
	tableParams := json.RawMessage(`{"type":"object","properties":{"table":{"type":"string"}},"required":["table"]}`)
	return []ToolSpec{
		{Name: "list_databases", Description: "列出当前连接可见的数据库", Parameters: empty},
		{Name: "list_tables", Description: "列出当前库的表", Parameters: empty},
		{Name: "get_table_schema", Description: "获取指定表的列结构", Parameters: tableParams},
	}
}

func sshToolSpecs() []ToolSpec {
	params := json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`)
	empty := json.RawMessage(`{"type":"object","properties":{}}`)
	return []ToolSpec{
		{Name: "ssh_probe", Description: "在独立会话执行只读探测命令", Parameters: params},
		{Name: "list_working_directory", Description: "列出当前工作目录的文件与权限；仅用于识别启动、停止或重启脚本", Parameters: empty},
	}
}

func toolsForMode(mode string) []ToolSpec {
	if mode == "ssh" {
		return sshToolSpecs()
	}
	return schemaToolSpecs()
}

func truncateToolResult(s string) string {
	return truncateToolResultTo(s, MaxToolResultChars)
}
func truncateToolResultTo(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}

func allowedTool(mode, name string) bool {
	for _, t := range toolsForMode(mode) {
		if t.Name == name {
			return true
		}
	}
	return false
}

func (s *Service) runTool(mode, sessionID, workingDir, name, arguments string) (result, note string) {
	if !allowedTool(mode, name) {
		return "工具被拒绝", "工具被拒绝"
	}
	switch name {
	case "list_working_directory":
		if s.commands == nil {
			return "command runner unavailable", "工具失败"
		}
		cmd, ok := WorkingDirectoryCommand(workingDir)
		if !ok {
			return "当前工作目录不可用", "工具失败"
		}
		stdout, stderr, err := s.commands.ExecuteCommand(sessionID, cmd, sshProbeTimeout)
		if err != nil {
			msg := err.Error()
			if stderr != "" {
				msg = stderr + "\n" + msg
			}
			return msg, "工具失败"
		}
		if stderr != "" {
			return stdout + "\n" + stderr, ""
		}
		return stdout, ""
	case "list_databases":
		if s.schema == nil {
			return "schema reader unavailable", "工具失败"
		}
		dbs, err := s.schema.ListDatabases(sessionID)
		if err != nil {
			return err.Error(), "工具失败"
		}
		return marshalToolJSON(dbs), ""
	case "list_tables":
		if s.schema == nil {
			return "schema reader unavailable", "工具失败"
		}
		tables, err := s.schema.ListTables(sessionID)
		if err != nil {
			return err.Error(), "工具失败"
		}
		return marshalToolJSON(tables), ""
	case "get_table_schema":
		if s.schema == nil {
			return "schema reader unavailable", "工具失败"
		}
		var args struct {
			Table string `json:"table"`
		}
		_ = json.Unmarshal([]byte(arguments), &args)
		schema, err := s.schema.GetTableSchema(sessionID, args.Table)
		if err != nil {
			return err.Error(), "工具失败"
		}
		if schema == nil {
			schema = &config.TableSchema{}
		}
		return marshalToolJSON(schema), ""
	case "ssh_probe":
		var args struct {
			Command string `json:"command"`
		}
		_ = json.Unmarshal([]byte(arguments), &args)
		cmd := strings.TrimSpace(args.Command)
		if !AllowSSHProbe(cmd) {
			return "工具被拒绝", "工具被拒绝"
		}
		if s.commands == nil {
			return "command runner unavailable", "工具失败"
		}
		stdout, stderr, err := s.commands.ExecuteCommand(sessionID, cmd, sshProbeTimeout)
		if err != nil {
			msg := err.Error()
			if stderr != "" {
				msg = stderr + "\n" + msg
			}
			return msg, "工具失败"
		}
		out := stdout
		if stderr != "" {
			out = stdout + "\n" + stderr
		}
		return out, ""
	default:
		return "工具被拒绝", "工具被拒绝"
	}
}

func marshalToolJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(raw)
}
