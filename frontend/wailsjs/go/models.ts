export namespace config {
	
	export class FileManagerSettings {
	    directory_tracking: boolean;
	    history_enabled: boolean;
	    history_limit: number;
	    history: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileManagerSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.directory_tracking = source["directory_tracking"];
	        this.history_enabled = source["history_enabled"];
	        this.history_limit = source["history_limit"];
	        this.history = source["history"];
	    }
	}
	export class AppSettings {
	    theme: string;
	    theme_mode: string;
	    use_system_theme: boolean;
	    accent_color: string;
	    font_family: string;
	    font_size: number;
	    terminal_theme: string;
	    terminal_font_family: string;
	    terminal_font_size: number;
	    compact_mode: boolean;
	    reduced_motion: boolean;
	    sidebar_width: number;
	    background_image_enabled: boolean;
	    background_image_path: string;
	    background_image_fit: string;
	    background_image_opacity: number;
	    jdbc_runtime_mode: string;
	    jdbc_system_java_path: string;
	    copilot_provider: string;
	    copilot_base_url: string;
	    copilot_model: string;
	    monitor_collapsed: boolean;
	    monitor_width: number;
	    monitor_refresh_interval: number;
	    file_manager_collapsed: boolean;
	    file_manager_width: number;
	    file_manager_show_hidden: boolean;
	    file_manager_sort_by: string;
	    file_manager_sort_order: string;
	    file_manager_per_connection?: Record<string, FileManagerSettings>;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.theme_mode = source["theme_mode"];
	        this.use_system_theme = source["use_system_theme"];
	        this.accent_color = source["accent_color"];
	        this.font_family = source["font_family"];
	        this.font_size = source["font_size"];
	        this.terminal_theme = source["terminal_theme"];
	        this.terminal_font_family = source["terminal_font_family"];
	        this.terminal_font_size = source["terminal_font_size"];
	        this.compact_mode = source["compact_mode"];
	        this.reduced_motion = source["reduced_motion"];
	        this.sidebar_width = source["sidebar_width"];
	        this.background_image_enabled = source["background_image_enabled"];
	        this.background_image_path = source["background_image_path"];
	        this.background_image_fit = source["background_image_fit"];
	        this.background_image_opacity = source["background_image_opacity"];
	        this.jdbc_runtime_mode = source["jdbc_runtime_mode"];
	        this.jdbc_system_java_path = source["jdbc_system_java_path"];
	        this.copilot_provider = source["copilot_provider"];
	        this.copilot_base_url = source["copilot_base_url"];
	        this.copilot_model = source["copilot_model"];
	        this.monitor_collapsed = source["monitor_collapsed"];
	        this.monitor_width = source["monitor_width"];
	        this.monitor_refresh_interval = source["monitor_refresh_interval"];
	        this.file_manager_collapsed = source["file_manager_collapsed"];
	        this.file_manager_width = source["file_manager_width"];
	        this.file_manager_show_hidden = source["file_manager_show_hidden"];
	        this.file_manager_sort_by = source["file_manager_sort_by"];
	        this.file_manager_sort_order = source["file_manager_sort_order"];
	        this.file_manager_per_connection = this.convertValues(source["file_manager_per_connection"], FileManagerSettings, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnSchema {
	    name: string;
	    type: string;
	    nullable: boolean;
	    is_primary_key: boolean;
	    column_size: number;
	    decimal_digits: number;
	    default_value: string;
	    has_default: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.is_primary_key = source["is_primary_key"];
	        this.column_size = source["column_size"];
	        this.decimal_digits = source["decimal_digits"];
	        this.default_value = source["default_value"];
	        this.has_default = source["has_default"];
	        this.description = source["description"];
	    }
	}
	export class ConnectionConfig {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    user: string;
	    auth_type: string;
	    key_path?: string;
	    tags?: string[];
	    metadata?: Record<string, string>;
	    type?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.auth_type = source["auth_type"];
	        this.key_path = source["key_path"];
	        this.tags = source["tags"];
	        this.metadata = source["metadata"];
	        this.type = source["type"];
	    }
	}
	
	export class JDBCProp {
	    name: string;
	    defaultValue?: string;
	    required?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JDBCProp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.defaultValue = source["defaultValue"];
	        this.required = source["required"];
	    }
	}
	export class JDBCJar {
	    name: string;
	    sha256: string;
	    url?: string;
	
	    static createFrom(source: any = {}) {
	        return new JDBCJar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sha256 = source["sha256"];
	        this.url = source["url"];
	    }
	}
	export class JDBCDriverProfile {
	    id: string;
	    version: string;
	    driverClass: string;
	    urlTemplate: string;
	    defaultPort: number;
	    jre: string;
	    jars: JDBCJar[];
	    properties?: JDBCProp[];
	    source?: string;
	    installed?: boolean;
	    installPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new JDBCDriverProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.driverClass = source["driverClass"];
	        this.urlTemplate = source["urlTemplate"];
	        this.defaultPort = source["defaultPort"];
	        this.jre = source["jre"];
	        this.jars = this.convertValues(source["jars"], JDBCJar);
	        this.properties = this.convertValues(source["properties"], JDBCProp);
	        this.source = source["source"];
	        this.installed = source["installed"];
	        this.installPath = source["installPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TableSchema {
	    table_name: string;
	    columns: ColumnSchema[];
	
	    static createFrom(source: any = {}) {
	        return new TableSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table_name = source["table_name"];
	        this.columns = this.convertValues(source["columns"], ColumnSchema);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace copilot {
	
	export class Artifact {
	    type: string;
	    content: string;
	    summary: string;
	    destructive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Artifact(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	        this.summary = source["summary"];
	        this.destructive = source["destructive"];
	    }
	}
	export class ToolCall {
	    id: string;
	    name: string;
	    arguments: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class Message {
	    role: string;
	    content?: string;
	    tool_call_id?: string;
	    tool_calls?: ToolCall[];
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.tool_call_id = source["tool_call_id"];
	        this.tool_calls = this.convertValues(source["tool_calls"], ToolCall);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatRequest {
	    SessionID: string;
	    Mode: string;
	    Message: string;
	    Model: string;
	    History: Message[];
	    EditorContent: string;
	    TerminalTail: string;
	    Host: string;
	    User: string;
	    DBType: string;
	    Database: string;
	    WorkingDir: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SessionID = source["SessionID"];
	        this.Mode = source["Mode"];
	        this.Message = source["Message"];
	        this.Model = source["Model"];
	        this.History = this.convertValues(source["History"], Message);
	        this.EditorContent = source["EditorContent"];
	        this.TerminalTail = source["TerminalTail"];
	        this.Host = source["Host"];
	        this.User = source["User"];
	        this.DBType = source["DBType"];
	        this.Database = source["Database"];
	        this.WorkingDir = source["WorkingDir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatResponse {
	    reply: string;
	    artifact?: Artifact;
	    tool_notes: string[];
	
	    static createFrom(source: any = {}) {
	        return new ChatResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reply = source["reply"];
	        this.artifact = this.convertValues(source["artifact"], Artifact);
	        this.tool_notes = source["tool_notes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Result {
	    Destructive: boolean;
	    Reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Destructive = source["Destructive"];
	        this.Reason = source["Reason"];
	    }
	}

}

export namespace service {
	
	export class BackgroundImageResult {
	    path: string;
	    data_url: string;
	    fit: string;
	    opacity: number;
	
	    static createFrom(source: any = {}) {
	        return new BackgroundImageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.data_url = source["data_url"];
	        this.fit = source["fit"];
	        this.opacity = source["opacity"];
	    }
	}
	export class DriverView {
	    id: string;
	    name: string;
	    recommendedVersion: string;
	    installed: boolean;
	    profiles: config.JDBCDriverProfile[];
	
	    static createFrom(source: any = {}) {
	        return new DriverView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.recommendedVersion = source["recommendedVersion"];
	        this.installed = source["installed"];
	        this.profiles = this.convertValues(source["profiles"], config.JDBCDriverProfile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class JDBCAgentStatus {
	    state: string;
	    runtimeKind: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new JDBCAgentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.runtimeKind = source["runtimeKind"];
	        this.lastError = source["lastError"];
	    }
	}
	export class JDBCLogTail {
	    content: string;
	    truncated: boolean;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new JDBCLogTail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.truncated = source["truncated"];
	        this.size = source["size"];
	    }
	}
	export class RuntimeStatus {
	    kind: string;
	    javaPath: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.javaPath = source["javaPath"];
	        this.version = source["version"];
	    }
	}
	export class JDBCRuntimeActivationResult {
	    runtime: RuntimeStatus;
	    agent: JDBCAgentStatus;
	
	    static createFrom(source: any = {}) {
	        return new JDBCRuntimeActivationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runtime = this.convertValues(source["runtime"], RuntimeStatus);
	        this.agent = this.convertValues(source["agent"], JDBCAgentStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class JSONValidationResult {
	    valid: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new JSONValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.error = source["error"];
	    }
	}
	export class NativeResource {
	    kind: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	    }
	}
	export class NativeResourceDetails {
	    kind: string;
	    name: string;
	    summary: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeResourceDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.summary = source["summary"];
	        this.content = source["content"];
	    }
	}
	export class NativeQueryResult {
	    summary: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeQueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.content = source["content"];
	    }
	}
	export class NativeMutationResult {
	    summary: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new NativeMutationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.content = source["content"];
	    }
	}
	
	export class TableDDL {
	    table_name: string;
	    ddl: string;
	    db_type: string;
	
	    static createFrom(source: any = {}) {
	        return new TableDDL(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table_name = source["table_name"];
	        this.ddl = source["ddl"];
	        this.db_type = source["db_type"];
	    }
	}
	export class URLDecodeResult {
	    decoded: string;
	    params?: Record<string, string>;
	    component?: string;
	
	    static createFrom(source: any = {}) {
	        return new URLDecodeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.decoded = source["decoded"];
	        this.params = source["params"];
	        this.component = source["component"];
	    }
	}
	export class URLEncodeResult {
	    encoded: string;
	    fullUrl?: string;
	    component?: string;
	
	    static createFrom(source: any = {}) {
	        return new URLEncodeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.encoded = source["encoded"];
	        this.fullUrl = source["fullUrl"];
	        this.component = source["component"];
	    }
	}

}

export namespace ssh {
	
	export class CPUMetrics {
	    overall: number;
	    user: number;
	    system: number;
	    iowait: number;
	    idle: number;
	    per_core: number[];
	    load_average: number[];
	
	    static createFrom(source: any = {}) {
	        return new CPUMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.user = source["user"];
	        this.system = source["system"];
	        this.iowait = source["iowait"];
	        this.idle = source["idle"];
	        this.per_core = source["per_core"];
	        this.load_average = source["load_average"];
	    }
	}
	export class PartitionInfo {
	    mount_point: string;
	    total: number;
	    used: number;
	    free: number;
	    used_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new PartitionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mount_point = source["mount_point"];
	        this.total = source["total"];
	        this.used = source["used"];
	        this.free = source["free"];
	        this.used_percent = source["used_percent"];
	    }
	}
	export class DiskMetrics {
	    partitions: PartitionInfo[];
	
	    static createFrom(source: any = {}) {
	        return new DiskMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.partitions = this.convertValues(source["partitions"], PartitionInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    size: number;
	    mode: string;
	    mod_time: string;
	    is_dir: boolean;
	    is_symlink: boolean;
	    link_target?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.mode = source["mode"];
	        this.mod_time = source["mod_time"];
	        this.is_dir = source["is_dir"];
	        this.is_symlink = source["is_symlink"];
	        this.link_target = source["link_target"];
	    }
	}
	export class MemoryMetrics {
	    total: number;
	    used: number;
	    free: number;
	    available: number;
	    used_percent: number;
	    swap_total: number;
	    swap_used: number;
	    swap_free: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.free = source["free"];
	        this.available = source["available"];
	        this.used_percent = source["used_percent"];
	        this.swap_total = source["swap_total"];
	        this.swap_used = source["swap_used"];
	        this.swap_free = source["swap_free"];
	    }
	}
	export class NetworkMetrics {
	    total_rx_bytes: number;
	    total_tx_bytes: number;
	    rx_rate: number;
	    tx_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new NetworkMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_rx_bytes = source["total_rx_bytes"];
	        this.total_tx_bytes = source["total_tx_bytes"];
	        this.rx_rate = source["rx_rate"];
	        this.tx_rate = source["tx_rate"];
	    }
	}
	export class SystemInfo {
	    hostname: string;
	    uptime: string;
	    os: string;
	    kernel: string;
	    username: string;
	    processes: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.uptime = source["uptime"];
	        this.os = source["os"];
	        this.kernel = source["kernel"];
	        this.username = source["username"];
	        this.processes = source["processes"];
	    }
	}
	export class MonitoringData {
	    timestamp: number;
	    system: SystemInfo;
	    cpu: CPUMetrics;
	    memory: MemoryMetrics;
	    network: NetworkMetrics;
	    disk: DiskMetrics;
	
	    static createFrom(source: any = {}) {
	        return new MonitoringData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.system = this.convertValues(source["system"], SystemInfo);
	        this.cpu = this.convertValues(source["cpu"], CPUMetrics);
	        this.memory = this.convertValues(source["memory"], MemoryMetrics);
	        this.network = this.convertValues(source["network"], NetworkMetrics);
	        this.disk = this.convertValues(source["disk"], DiskMetrics);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class SearchResult {
	    path: string;
	    name: string;
	    depth: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.depth = source["depth"];
	    }
	}
	
	export class TransferProgress {
	    transfer_id: string;
	    session_id: string;
	    filename: string;
	    bytes_sent: number;
	    total_bytes: number;
	    percentage: number;
	    speed: number;
	    status: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transfer_id = source["transfer_id"];
	        this.session_id = source["session_id"];
	        this.filename = source["filename"];
	        this.bytes_sent = source["bytes_sent"];
	        this.total_bytes = source["total_bytes"];
	        this.percentage = source["percentage"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.error = source["error"];
	    }
	}

}

