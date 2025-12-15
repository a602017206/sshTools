export namespace config {
	
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

