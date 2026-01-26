export namespace config {
	
	export class AppSettings {
	    theme: string;
	    font_family: string;
	    font_size: number;
	    terminal_theme: string;
	    sidebar_width: number;
	    monitor_collapsed: boolean;
	    monitor_width: number;
	    monitor_refresh_interval: number;
	    file_manager_collapsed: boolean;
	    file_manager_width: number;
	    file_manager_show_hidden: boolean;
	    file_manager_sort_by: string;
	    file_manager_sort_order: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.font_family = source["font_family"];
	        this.font_size = source["font_size"];
	        this.terminal_theme = source["terminal_theme"];
	        this.sidebar_width = source["sidebar_width"];
	        this.monitor_collapsed = source["monitor_collapsed"];
	        this.monitor_width = source["monitor_width"];
	        this.monitor_refresh_interval = source["monitor_refresh_interval"];
	        this.file_manager_collapsed = source["file_manager_collapsed"];
	        this.file_manager_width = source["file_manager_width"];
	        this.file_manager_show_hidden = source["file_manager_show_hidden"];
	        this.file_manager_sort_by = source["file_manager_sort_by"];
	        this.file_manager_sort_order = source["file_manager_sort_order"];
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
	    }
	}

}

export namespace service {
	
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

