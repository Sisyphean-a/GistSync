export namespace settings {
	
	export class ProfileItem {
	    id: string;
	    sourcePathTemplate: string;
	    relativePath: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourcePathTemplate = source["sourcePathTemplate"];
	        this.relativePath = source["relativePath"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Profile {
	    id: string;
	    name: string;
	    restoreMode: string;
	    restoreRoot: string;
	    enabled: boolean;
	    items: ProfileItem[];
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.restoreMode = source["restoreMode"];
	        this.restoreRoot = source["restoreRoot"];
	        this.enabled = source["enabled"];
	        this.items = this.convertValues(source["items"], ProfileItem);
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
	export class Data {
	    token: string;
	    masterPassword: string;
	    activeProfileId: string;
	    profiles: Profile[];
	    cloudBootstrapDone?: boolean;
	    syncPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.masterPassword = source["masterPassword"];
	        this.activeProfileId = source["activeProfileId"];
	        this.profiles = this.convertValues(source["profiles"], Profile);
	        this.cloudBootstrapDone = source["cloudBootstrapDone"];
	        this.syncPath = source["syncPath"];
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

export namespace syncflow {
	
	export class ApplyConflict {
	    itemId: string;
	    targetPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyConflict(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.targetPath = source["targetPath"];
	    }
	}
	export class ApplyItemResult {
	    itemId: string;
	    targetPath: string;
	    status: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyItemResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.targetPath = source["targetPath"];
	        this.status = source["status"];
	        this.reason = source["reason"];
	    }
	}
	export class ApplySnapshotRequest {
	    profileId: string;
	    snapshotId: string;
	    masterPassword: string;
	    restoreMode: string;
	    restoreRoot: string;
	    overwriteItemIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new ApplySnapshotRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.snapshotId = source["snapshotId"];
	        this.masterPassword = source["masterPassword"];
	        this.restoreMode = source["restoreMode"];
	        this.restoreRoot = source["restoreRoot"];
	        this.overwriteItemIds = source["overwriteItemIds"];
	    }
	}
	export class ApplySnapshotResult {
	    applied: number;
	    skipped: number;
	    items: ApplyItemResult[];
	
	    static createFrom(source: any = {}) {
	        return new ApplySnapshotResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applied = source["applied"];
	        this.skipped = source["skipped"];
	        this.items = this.convertValues(source["items"], ApplyItemResult);
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
	export class SnapshotMeta {
	    id: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class UploadProfileResult {
	    snapshotId: string;
	    uploaded: number;
	
	    static createFrom(source: any = {}) {
	        return new UploadProfileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.snapshotId = source["snapshotId"];
	        this.uploaded = source["uploaded"];
	    }
	}

}

