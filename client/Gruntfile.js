/*jshint node:true*/
module.exports = function(grunt) {
    grunt.initConfig({

        clean: {
            main: ['js']
        },

        typescript: {
            base: {
                src: ['ts/**/*.ts'],
                dest: 'js',
                options: {
                    comments: true,
                    basePath: 'ts',
                    module: 'amd',
                    target: 'es5',
                    sourceMap: true
                },
            }
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

    grunt.registerTask('default', [ 'clean', 'typescript' ]);
    grunt.registerTask('three', [ 'uglify:js' ]);

};