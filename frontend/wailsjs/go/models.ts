export namespace ffmpeg {
	
	export class Format {
	    filename: string;
	    format_name: string;
	    duration: string;
	    size: string;
	    bit_rate: string;
	
	    static createFrom(source: any = {}) {
	        return new Format(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.format_name = source["format_name"];
	        this.duration = source["duration"];
	        this.size = source["size"];
	        this.bit_rate = source["bit_rate"];
	    }
	}
	export class Stream {
	    index: number;
	    codec_type: string;
	    codec_name: string;
	    width?: number;
	    height?: number;
	    r_frame_rate?: string;
	
	    static createFrom(source: any = {}) {
	        return new Stream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.codec_type = source["codec_type"];
	        this.codec_name = source["codec_name"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.r_frame_rate = source["r_frame_rate"];
	    }
	}
	export class MediaInfo {
	    format: Format;
	    streams: Stream[];
	
	    static createFrom(source: any = {}) {
	        return new MediaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = this.convertValues(source["format"], Format);
	        this.streams = this.convertValues(source["streams"], Stream);
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

export namespace jobs {
	
	export class Job {
	    id: string;
	    status: string;
	    progress: number;
	    type: string;
	    inputFile: string;
	    outputFile?: string;
	    error?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.type = source["type"];
	        this.inputFile = source["inputFile"];
	        this.outputFile = source["outputFile"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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

export namespace ops {
	
	export class ConvertParams {
	    inputPath: string;
	    outputPath: string;
	    preset?: string;
	    crf?: number;
	
	    static createFrom(source: any = {}) {
	        return new ConvertParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputPath = source["inputPath"];
	        this.outputPath = source["outputPath"];
	        this.preset = source["preset"];
	        this.crf = source["crf"];
	    }
	}

}

