export namespace main {
	
	export class AppInfo {
	    name: string;
	    runtime: string;
	    framework: string;
	    strategy: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.runtime = source["runtime"];
	        this.framework = source["framework"];
	        this.strategy = source["strategy"];
	        this.port = source["port"];
	    }
	}

}

