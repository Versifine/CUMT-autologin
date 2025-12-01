export namespace config {
	
	export class AccountConfig {
	    StudentID: string;
	    Carrier: string;
	    Password: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StudentID = source["StudentID"];
	        this.Carrier = source["Carrier"];
	        this.Password = source["Password"];
	    }
	}
	export class UIConfig {
	    Width: number;
	    Height: number;
	
	    static createFrom(source: any = {}) {
	        return new UIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	    }
	}
	export class PortalConfig {
	    LoginURL: string;
	    Method: string;
	    Form: Record<string, string>;
	    LogoutForm: Record<string, string>;
	    Headers: Record<string, string>;
	    SuccessKeywords: string[];
	
	    static createFrom(source: any = {}) {
	        return new PortalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LoginURL = source["LoginURL"];
	        this.Method = source["Method"];
	        this.Form = source["Form"];
	        this.LogoutForm = source["LogoutForm"];
	        this.Headers = source["Headers"];
	        this.SuccessKeywords = source["SuccessKeywords"];
	    }
	}
	export class Config {
	    WifiSSID: string;
	    CheckURL: string;
	    Account: AccountConfig;
	    Portal: PortalConfig;
	    UI: UIConfig;
	    auto_login_interval: number;
	    login_mode: string;
	    auto_start: boolean;
	    open_settings_on_run: boolean;
	    WindowX: number;
	    WindowY: number;
	    WindowW: number;
	    WindowH: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WifiSSID = source["WifiSSID"];
	        this.CheckURL = source["CheckURL"];
	        this.Account = this.convertValues(source["Account"], AccountConfig);
	        this.Portal = this.convertValues(source["Portal"], PortalConfig);
	        this.UI = this.convertValues(source["UI"], UIConfig);
	        this.auto_login_interval = source["auto_login_interval"];
	        this.login_mode = source["login_mode"];
	        this.auto_start = source["auto_start"];
	        this.open_settings_on_run = source["open_settings_on_run"];
	        this.WindowX = source["WindowX"];
	        this.WindowY = source["WindowY"];
	        this.WindowW = source["WindowW"];
	        this.WindowH = source["WindowH"];
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

export namespace main {
	
	export class Status {
	    online: boolean;
	    message: string;
	    // Go type: time
	    last_check: any;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.online = source["online"];
	        this.message = source["message"];
	        this.last_check = this.convertValues(source["last_check"], null);
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

