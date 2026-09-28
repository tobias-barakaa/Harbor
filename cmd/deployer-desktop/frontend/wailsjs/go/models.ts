export namespace main {
	
	export class AddServerInput {
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    authMethod: string;
	    privateKeyPath: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new AddServerInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	        this.privateKeyPath = source["privateKeyPath"];
	        this.password = source["password"];
	    }
	}
	export class ApplicationEntry {
	    id: string;
	    path: string;
	    name: string;
	    runtime: string;
	    framework: string;
	    strategy: string;
	    port: number;
	    addedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.name = source["name"];
	        this.runtime = source["runtime"];
	        this.framework = source["framework"];
	        this.strategy = source["strategy"];
	        this.port = source["port"];
	        this.addedAt = source["addedAt"];
	    }
	}
	export class DeployInput {
	    applicationId: string;
	    serverName: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applicationId = source["applicationId"];
	        this.serverName = source["serverName"];
	    }
	}
	export class PortMappingInfo {
	    host: number;
	    container: number;
	
	    static createFrom(source: any = {}) {
	        return new PortMappingInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.container = source["container"];
	    }
	}
	export class DeploymentInfo {
	    name: string;
	    serverName: string;
	    containerName: string;
	    image: string;
	    ports: PortMappingInfo[];
	    deployedAt: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.serverName = source["serverName"];
	        this.containerName = source["containerName"];
	        this.image = source["image"];
	        this.ports = this.convertValues(source["ports"], PortMappingInfo);
	        this.deployedAt = source["deployedAt"];
	        this.status = source["status"];
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
	
	export class ServerInfo {
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    authMethod: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	    }
	}

}

