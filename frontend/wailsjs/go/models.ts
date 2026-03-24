export namespace settings {
	
	export class Data {
	    token: string;
	    masterPassword: string;
	    syncPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.masterPassword = source["masterPassword"];
	        this.syncPath = source["syncPath"];
	    }
	}

}

