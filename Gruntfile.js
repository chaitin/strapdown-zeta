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
        // use prettify js or highlight.js by uncomment the corresponding line
        src: ['vendor/marked.min.js', 'vendor/highlight.pack.js', 'vendor/jquery.min.js', 'vendor/bootstrap.min.js', 'vendor/cookie.js', 'build/strapdown-src.min.js'],
        // src: ['vendor/marked.min.js', 'vendor/prettify.min.js', 'build/strapdown-src.min.js'],
        dest: 'build/strapdown.min.js'
      }
    },
    cssmin: {
      minify: {
        src: ['src/strapdown.css'],
        dest: 'build/strapdown.min.css',
        ext: '.min.css'
      }
    }
  });

  // Load the plugin that provides the "uglify" task.
  grunt.loadNpmTasks('grunt-contrib-uglify');
  grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-cssmin');

  // Default task(s).
  grunt.registerTask('default', ['build']);
  grunt.registerTask('clean', function () {
    grunt.file['delete']('build');
  });
  grunt.registerTask('build', function () {
    // grunt.task.run('jshint');    // too many errors, sadly
    grunt.task.run('uglify');
    grunt.task.run('concat');
    grunt.task.run('cssmin');
  })
};

// vim: ai:ts=2:sts=2:sw=2:
