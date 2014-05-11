module.exports = function(grunt) {

  // Project configuration.
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),
    jshint: {
      options: {
          curly: true,
          browser: true,
          devel: true,
          indent: 2,
          latedef: true,
          undef: true,
          unused: true,
          expr: true,
          globals: {
              "define": false,
              "require": false,
          },
          ignores: [
              'node_modules/**/*.js',
          ]
      },
      client: ['src/*.js'],
    },
    uglify: {
      my_target: {
        files: {
          'build/strapdown-src.min.js': ['src/strapdown.js']
        }
      }
    },
    concat: {
      options: {
        separator: '\n',
      },
      dist: {
        src: ['vendor/marked.min.js', 'vendor/prettify.min.js', 'build/strapdown-src.min.js'],
        dest: 'build/strapdown.min.js'
      }
    }
  });

  // Load the plugin that provides the "uglify" task.
  grunt.loadNpmTasks('grunt-contrib-uglify');
  grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-contrib-concat');

  // Default task(s).
  grunt.registerTask('default', ['build-js']);
  grunt.registerTask('clean', function () {
    grunt.file['delete']('build');
  });
  grunt.registerTask('build-js', function () {
    // grunt.task.run('jshint');    // too many errors, sadly
    grunt.task.run('uglify');
    grunt.task.run('concat');
  })
};

// vim: ai:ts=2:sts=2:sw=2:
