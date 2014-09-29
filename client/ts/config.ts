/// <reference path="def/require.d.ts" />
/// <reference path="def/jquery.d.ts" />
/// <reference path="def/three.d.ts" />
/// <reference path="def/three.external.d.ts" />
/// <reference path="def/lodash.d.ts" />
/// <reference path="def/stats.d.ts" />
/// <reference path="def/detector.d.ts" />
/// <reference path="def/term.d.ts" />

require.config({
    urlArgs: "bust=" + (new Date()).getTime(),
    paths: {
        jquery: "//cdnjs.cloudflare.com/ajax/libs/jquery/2.1.1/jquery.min",
        lodash: "//cdnjs.cloudflare.com/ajax/libs/lodash.js/2.4.1/lodash.min",
        three: "../vendor/three.min",
        detector: '../vendor/Detector',
        stats: '../vendor/stats.min',
        term: "../vendor/term"
    },
    shim: {
        'detector': { exports: 'Detector' },
        'stats': { exports: 'Stats' },
        'term': { exports: 'Terminal' }
    },
    waitSeconds: 20
});

require(['main', 'three', 'jquery', 'lodash'], (main, THREE, $, _) => {
    var app = new main();
    app.start();
});