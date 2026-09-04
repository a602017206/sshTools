package copilot

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"AHaSSHTools/internal/config"
)

const (
	sshProbeTimeout         = 5 * time.Second
	schemaReaderUnavailable = "schema reader unavailable"
	nativeReaderUnavailable = "native resource reader unavailable"
)

func schemaToolSpecs() []ToolSpec {
	empty := json.RawMessage(`{"type":"object","properties":{}}`)
	tableParams := json.RawMessage(`{"type":"object","properties":{"table":{"type":"string"}}}`)
	return []ToolSpec{
		{Name: "list_databases", Description: "列出当前连接可见的数据库", Parameters: empty},
		{Name: "list_tables", Description: "列出当前打开的库或 schema 下的表", Parameters: empty},
		{Name: "get_table_schema", Description: "获取指定表的列结构；未传 table 时使用当前打开的表", Parameters: tableParams},
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

func nativeToolSpecs() []ToolSpec {
	empty := json.RawMessage(`{"type":"object","properties":{}}`)
	childParams := json.RawMessage(`{"type":"object","properties":{"parent":{"type":"string"}}}`)
	describeParams := json.RawMessage(`{"type":"object","properties":{"parent":{"type":"string"},"name":{"type":"string"}}}`)
	queryParams := json.RawMessage(`{"type":"object","properties":{"parent":{"type":"string"},"name":{"type":"string"},"query":{"type":"string"}},"required":["query"]}`)
	mutationParams := json.RawMessage(`{"type":"object","properties":{"operation":{"type":"string"},"parent":{"type":"string"},"name":{"type":"string"},"payload":{"type":"string"},"summary":{"type":"string"}},"required":["operation"]}`)
	return []ToolSpec{
		{Name: "list_resources", Description: "列出当前原生数据源的一级资源（如 Redis 逻辑库、ES 索引、Kafka Topic）", Parameters: empty},
		{Name: "list_child_resources", Description: "列出指定父资源下的子资源；未传 parent 时使用当前打开的父对象", Parameters: childParams},
		{Name: "describe_resource", Description: "读取当前或指定资源的只读详情；未传 name 时使用当前打开的对象", Parameters: describeParams},
		{Name: "execute_native_query", Description: "执行只读原生查询：Redis CLI（只读命令）或 Elasticsearch DSL/_search/Dev Tools GET；结果有条数与字节上限", Parameters: queryParams},
		{Name: "propose_native_mutation", Description: "提出写入/删除变更意图，不直接执行；前端会展示确认卡片。operation 如 save/create_key/delete/delete_keys/index_document/delete_document/create_index/delete_index", Parameters: mutationParams},
	}
}

func IsNativeDBType(dbType string) bool {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "redis", "mongodb", "elasticsearch", "opensearch", "memcached", "cassandra", "couchbase", "influxdb", "neo4j", "kafka", "rocketmq", "rabbitmq":
		return true
	default:
		return false
	}
}

func toolsForMode(mode string) []ToolSpec {
	return toolsForRequest(ChatRequest{Mode: mode})
}

func toolsForRequest(req ChatRequest) []ToolSpec {
	if strings.EqualFold(strings.TrimSpace(req.Mode), "ssh") {
		return sshToolSpecs()
	}
	if IsNativeDBType(req.DBType) {
		return nativeToolSpecs()
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

func allowedTool(req ChatRequest, name string) bool {
	for _, t := range toolsForRequest(req) {
		if t.Name == name {
			return true
		}
	}
	return false
}

func (s *Service) runTool(req ChatRequest, name, arguments string) (result, note string) {
	if !allowedTool(req, name) {
		return "工具被拒绝", "工具被拒绝"
	}
	switch name {
	case "list_working_directory":
		return s.listWorkingDirectory(req.SessionID, req.WorkingDir)
	case "list_databases":
		return s.listDatabases(req.SessionID)
	case "list_tables":
		return s.listTables(req)
	case "get_table_schema":
		return s.getTableSchema(req, arguments)
	case "list_resources":
		return s.listResources(req.SessionID)
	case "list_child_resources":
		return s.listChildResources(req, arguments)
	case "describe_resource":
		return s.describeResource(req, arguments)
	case "execute_native_query":
		return s.executeNativeQuery(req, arguments)
	case "propose_native_mutation":
		return s.proposeNativeMutation(arguments)
	case "ssh_probe":
		return s.sshProbe(req.SessionID, arguments)
	default:
		return "工具被拒绝", "工具被拒绝"
	}
}

func (s *Service) listWorkingDirectory(sessionID, workingDir string) (string, string) {
	cmd, ok := WorkingDirectoryCommand(workingDir)
	if !ok {
		return "当前工作目录不可用", "工具失败"
	}
	return s.executeToolCommand(sessionID, cmd)
}

func (s *Service) listDatabases(sessionID string) (string, string) {
	if s.schema == nil {
		return schemaReaderUnavailable, "工具失败"
	}
	dbs, err := s.schema.ListDatabases(sessionID)
	if err != nil {
		return err.Error(), "工具失败"
	}
	return marshalToolJSON(dbs), ""
}

func (s *Service) listTables(req ChatRequest) (string, string) {
	if s.schema == nil {
		return schemaReaderUnavailable, "工具失败"
	}
	var (
		tables []string
		err    error
	)
	if scoped, ok := s.schema.(ScopedSchemaReader); ok && (strings.TrimSpace(req.Database) != "" || strings.TrimSpace(req.Schema) != "") {
		tables, err = scoped.ListTablesInScope(req.SessionID, req.Database, req.Schema)
	} else {
		tables, err = s.schema.ListTables(req.SessionID)
	}
	if err != nil {
		return err.Error(), "工具失败"
	}
	return marshalToolJSON(tables), ""
}

func (s *Service) getTableSchema(req ChatRequest, arguments string) (string, string) {
	if s.schema == nil {
		return schemaReaderUnavailable, "工具失败"
	}
	var args struct {
		Table string `json:"table"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	table := strings.TrimSpace(args.Table)
	if table == "" {
		table = strings.TrimSpace(req.ObjectName)
	}
	if table == "" {
		return "未指定表", "工具失败"
	}
	var (
		schema *config.TableSchema
		err    error
	)
	if scoped, ok := s.schema.(ScopedSchemaReader); ok && (strings.TrimSpace(req.Database) != "" || strings.TrimSpace(req.Schema) != "") {
		schema, err = scoped.GetTableSchemaInScope(req.SessionID, req.Database, req.Schema, table)
	} else {
		schema, err = s.schema.GetTableSchema(req.SessionID, table)
	}
	if err != nil {
		return err.Error(), "工具失败"
	}
	if schema == nil {
		schema = &config.TableSchema{}
	}
	return marshalToolJSON(schema), ""
}

func (s *Service) listResources(sessionID string) (string, string) {
	if s.native == nil {
		return nativeReaderUnavailable, "工具失败"
	}
	items, err := s.native.ListResources(sessionID)
	if err != nil {
		return err.Error(), "工具失败"
	}
	return marshalToolJSON(items), ""
}

func (s *Service) listChildResources(req ChatRequest, arguments string) (string, string) {
	if s.native == nil {
		return nativeReaderUnavailable, "工具失败"
	}
	var args struct {
		Parent string `json:"parent"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	parent := strings.TrimSpace(args.Parent)
	if parent == "" {
		parent = strings.TrimSpace(req.ObjectParent)
	}
	items, err := s.native.ListChildResources(req.SessionID, parent)
	if err != nil {
		return err.Error(), "工具失败"
	}
	return marshalToolJSON(items), ""
}

func (s *Service) describeResource(req ChatRequest, arguments string) (string, string) {
	if s.native == nil {
		return nativeReaderUnavailable, "工具失败"
	}
	var args struct {
		Parent string `json:"parent"`
		Name   string `json:"name"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	parent := strings.TrimSpace(args.Parent)
	if parent == "" {
		parent = strings.TrimSpace(req.ObjectParent)
	}
	name := strings.TrimSpace(args.Name)
	if name == "" {
		name = strings.TrimSpace(req.ObjectName)
	}
	if name == "" {
		return "未指定资源", "工具失败"
	}
	view, err := s.native.DescribeResource(req.SessionID, parent, name)
	if err != nil {
		return err.Error(), "工具失败"
	}
	if view == nil {
		view = &NativeResourceView{}
	}
	view.Summary = Redact(view.Summary)
	view.Content = Redact(view.Content)
	return marshalToolJSON(view), ""
}

func (s *Service) executeNativeQuery(req ChatRequest, arguments string) (string, string) {
	querier, ok := s.native.(NativeQuerier)
	if !ok || s.native == nil {
		return nativeReaderUnavailable, "工具失败"
	}
	var args struct {
		Parent string `json:"parent"`
		Name   string `json:"name"`
		Query  string `json:"query"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	parent := strings.TrimSpace(args.Parent)
	if parent == "" {
		parent = strings.TrimSpace(req.ObjectParent)
		if parent == "" {
			parent = strings.TrimSpace(req.Database)
		}
	}
	name := strings.TrimSpace(args.Name)
	if name == "" {
		name = strings.TrimSpace(req.ObjectName)
	}
	query := strings.TrimSpace(args.Query)
	if query == "" {
		query = strings.TrimSpace(req.EditorContent)
	}
	if query == "" {
		return "未指定查询", "工具失败"
	}
	// Force Redis CLI into readonly for AI tool path.
	if IsNativeDBType(req.DBType) && strings.EqualFold(req.DBType, "redis") {
		if strings.HasPrefix(query, "{") {
			var envelope map[string]any
			if err := json.Unmarshal([]byte(query), &envelope); err == nil {
				envelope["readOnly"] = true
				if _, hasMode := envelope["mode"]; !hasMode {
					envelope["mode"] = "cli"
				}
				if raw, err := json.Marshal(envelope); err == nil {
					query = string(raw)
				}
			}
		} else {
			payload, _ := json.Marshal(map[string]any{"mode": "cli", "command": query, "readOnly": true})
			query = string(payload)
		}
	}
	out, err := querier.ExecuteQuery(req.SessionID, parent, name, query)
	if err != nil {
		return err.Error(), "工具失败"
	}
	return truncateToolResult(Redact(out)), ""
}

func (s *Service) proposeNativeMutation(arguments string) (string, string) {
	var args struct {
		Operation string `json:"operation"`
		Parent    string `json:"parent"`
		Name      string `json:"name"`
		Payload   string `json:"payload"`
		Summary   string `json:"summary"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	op := strings.TrimSpace(args.Operation)
	if op == "" {
		return "未指定 operation", "工具失败"
	}
	payload := strings.TrimSpace(args.Payload)
	if payload == "" {
		payload = "{}"
	}
	summary := strings.TrimSpace(args.Summary)
	if summary == "" {
		summary = op + " " + strings.TrimSpace(args.Name)
	}
	art := map[string]any{
		"type":        "native_mutation",
		"summary":     summary,
		"destructive": strings.Contains(strings.ToLower(op), "delete"),
		"content": map[string]string{
			"operation": op,
			"parent":    strings.TrimSpace(args.Parent),
			"name":      strings.TrimSpace(args.Name),
			"payload":   payload,
		},
	}
	// Flatten content as JSON string for artifact parser compatibility when model copies tool output.
	contentObj := map[string]string{
		"operation": op,
		"parent":    strings.TrimSpace(args.Parent),
		"name":      strings.TrimSpace(args.Name),
		"payload":   payload,
	}
	contentRaw, _ := json.Marshal(contentObj)
	art["content"] = string(contentRaw)
	return marshalToolJSON(art), "已生成变更提案，需用户确认后执行"
}

func (s *Service) sshProbe(sessionID, arguments string) (string, string) {
	var args struct {
		Command string `json:"command"`
	}
	_ = json.Unmarshal([]byte(arguments), &args)
	cmd := strings.TrimSpace(args.Command)
	if !AllowSSHProbe(cmd) {
		return "工具被拒绝", "工具被拒绝"
	}
	return s.executeToolCommand(sessionID, cmd)
}

func (s *Service) executeToolCommand(sessionID, command string) (string, string) {
	if s.commands == nil {
		return "command runner unavailable", "工具失败"
	}
	stdout, stderr, err := s.commands.ExecuteCommand(sessionID, command, sshProbeTimeout)
	if err != nil {
		if stderr != "" {
			return stderr + "\n" + err.Error(), "工具失败"
		}
		return err.Error(), "工具失败"
	}
	if stderr != "" {
		return stdout + "\n" + stderr, ""
	}
	return stdout, ""
}

func marshalToolJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(raw)
}
