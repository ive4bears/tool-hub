export namespace hub {
	
	export class CmdToolBody {
	    args: Record<string, string>;
	    stdin: string;
	    env: Record<string, string>;
	    working_dir: string;
	    timeout: string;
	    caller: string;
	
	    static createFrom(source: any = {}) {
	        return new CmdToolBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.args = source["args"];
	        this.stdin = source["stdin"];
	        this.env = source["env"];
	        this.working_dir = source["working_dir"];
	        this.timeout = source["timeout"];
	        this.caller = source["caller"];
	    }
	}
	export class CmdToolTestcase {
	    input: CmdToolBody;
	    expect: string;
	    matchType: string;
	
	    static createFrom(source: any = {}) {
	        return new CmdToolTestcase(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = this.convertValues(source["input"], CmdToolBody);
	        this.expect = source["expect"];
	        this.matchType = source["matchType"];
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
	export class ConcurrencyGroup {
	    id: number;
	    createdAt: number;
	    updatedAt: number;
	    name: string;
	    maxConcurrent: number;
	
	    static createFrom(source: any = {}) {
	        return new ConcurrencyGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.name = source["name"];
	        this.maxConcurrent = source["maxConcurrent"];
	    }
	}
	export class TestcaseForDependency {
	    cmd: string[];
	    expect: string;
	    matchType: string;
	
	    static createFrom(source: any = {}) {
	        return new TestcaseForDependency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cmd = source["cmd"];
	        this.expect = source["expect"];
	        this.matchType = source["matchType"];
	    }
	}
	export class Dependency {
	    id: number;
	    createdAt: number;
	    updatedAt: number;
	    name: string;
	    description: string;
	    doc: string;
	    url: string;
	    installCmd: string;
	    testcases: TestcaseForDependency[];
	
	    static createFrom(source: any = {}) {
	        return new Dependency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.doc = source["doc"];
	        this.url = source["url"];
	        this.installCmd = source["installCmd"];
	        this.testcases = this.convertValues(source["testcases"], TestcaseForDependency);
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
	export class CommandLineTool {
	    id: number;
	    createdAt: number;
	    updatedAt: number;
	    name: string;
	    description: string;
	    parameters: string;
	    type: string;
	    logLifeSpan: string;
	    wd: string;
	    cmd: string[];
	    env: Record<string, string>;
	    timeout: string;
	    isStream: boolean;
	    dependencies: Dependency[];
	    error: string;
	    status: string;
	    testcases: CmdToolTestcase[];
	    concurrencyGroupID?: number;
	    concurrencyGroup?: ConcurrencyGroup;
	
	    static createFrom(source: any = {}) {
	        return new CommandLineTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.parameters = source["parameters"];
	        this.type = source["type"];
	        this.logLifeSpan = source["logLifeSpan"];
	        this.wd = source["wd"];
	        this.cmd = source["cmd"];
	        this.env = source["env"];
	        this.timeout = source["timeout"];
	        this.isStream = source["isStream"];
	        this.dependencies = this.convertValues(source["dependencies"], Dependency);
	        this.error = source["error"];
	        this.status = source["status"];
	        this.testcases = this.convertValues(source["testcases"], CmdToolTestcase);
	        this.concurrencyGroupID = source["concurrencyGroupID"];
	        this.concurrencyGroup = this.convertValues(source["concurrencyGroup"], ConcurrencyGroup);
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
	
	
	export class Dirs {
	    home: string;
	    temp: string;
	
	    static createFrom(source: any = {}) {
	        return new Dirs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.home = source["home"];
	        this.temp = source["temp"];
	    }
	}
	export class ServiceTool {
	    id: number;
	    createdAt: number;
	    updatedAt: number;
	    name: string;
	    description: string;
	    parameters: string;
	    type: string;
	    logLifeSpan: string;
	    startCmd: string;
	    error: string;
	    status: string;
	    concurrencyGroupID?: number;
	    concurrencyGroup?: ConcurrencyGroup;
	
	    static createFrom(source: any = {}) {
	        return new ServiceTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.parameters = source["parameters"];
	        this.type = source["type"];
	        this.logLifeSpan = source["logLifeSpan"];
	        this.startCmd = source["startCmd"];
	        this.error = source["error"];
	        this.status = source["status"];
	        this.concurrencyGroupID = source["concurrencyGroupID"];
	        this.concurrencyGroup = this.convertValues(source["concurrencyGroup"], ConcurrencyGroup);
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
	
	export class Tool {
	    id: number;
	    createdAt: number;
	    updatedAt: number;
	    name: string;
	    description: string;
	    parameters: string;
	    type: string;
	    logLifeSpan: string;
	
	    static createFrom(source: any = {}) {
	        return new Tool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.parameters = source["parameters"];
	        this.type = source["type"];
	        this.logLifeSpan = source["logLifeSpan"];
	    }
	}

}

export namespace main {
	
	export enum StringValues {
	    ColorOfIcebear = "white",
	    ColorOfMeiMeiBear = "pink",
	    MyFavoriteBear = "MeiMeiBear",
	}
	export enum IntValues {
	    NumberOfBears = 4,
	}

}

