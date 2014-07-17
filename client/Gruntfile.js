/*jshint node:true*/
module.exports = function(grunt) {
    grunt.initConfig({

        clean: {
            main: ['js']
        },

        typescript: {
            options: {
                basePath: 'ts',
                comments: true,
                module: 'amd',
                target: 'es5',
                sourceMap: true
            },
            amd: {
                dest: 'js',
                options: { module: 'amd' },
                src: [ 'ts/**/*.ts' ]
            },
        },
        uglify: {
            js: {
                files: { 'vendor/three.min.js': ['vendor/three/three.js', 'vendor/three/*.js'] },
                options: {
                    preserveComments: false
                }
            }
        }
    });

    require('load-grunt-tasks')(grunt);

    grunt.registerTask('default', [ 'clean', 'typescript:amd' ]);
    grunt.registerTask('three', [ 'uglify:js' ]);

};