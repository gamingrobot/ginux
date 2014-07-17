declare module "term" {
    class Terminal {
        x: any;
        y: any;
        constructor(options);
        open(parent):void;
        write(data):void;
        on(type, handler: (data: any) => any):void;
    }
    export = Terminal;
}