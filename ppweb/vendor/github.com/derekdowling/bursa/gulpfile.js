var gulp = require('gulp');
var shell = require('gulp-shell');
var notify = require('gulp-notify');
var notifier = require('node-notifier');
var path = require('path');

// Slightly Stripped Rubix CSS

var path = require('path');
var gulp = require('gulp');

var flip = require('css-flip');
var map = require('map-stream');
var through = require('through');
var plumber = require('gulp-plumber');
var transform = require('vinyl-transform');
var buffer = require('vinyl-buffer');
var child_process = require('child_process');
var browserSync = require('browser-sync');

var argv = require('yargs').argv;

var cjsx = require('gulp-cjsx');
var filter = require('gulp-filter');
var sass = require('gulp-sass');
var sourcemaps = require('gulp-sourcemaps');
var gutil = require('gulp-util');
var bless = require('gulp-bless');
var insert = require('gulp-insert');
var concat = require('gulp-concat');
var rename = require('gulp-rename');
var replace = require('gulp-replace');
var uglify = require('gulp-uglifyjs');
var webpack = require('gulp-webpack');
var ttf2woff = require('gulp-ttf2woff');
var cssfont64 = require('gulp-cssfont64');
var minifycss = require('gulp-minify-css');
var postcss = require('gulp-postcss');
var autoprefixer = require('gulp-autoprefixer');
var clean = require('gulp-clean');

var watchify = require('watchify');
var browserify = require('browserify');
var coffeeify = require('coffeeify');
var coffee_reactify = require('coffee-reactify');
var source = require('vinyl-source-stream');

var jest = require('jest-cli');

var runSequence = require('run-sequence');

var package = require('./package.json');

var production = argv.production ? true : false;

/* file patterns to watch */
var paths = {
  // l20n: ['src/global/vendor/l20n/*.jsx'],
  scss: ['assets/css/app/**/*.scss'],

  // Ignore bourbon and bootstrap when watching because they're huge.
  watch_scss: ['assets/css/app/**/*.scss', '!assets/css/app/bootstrap-sass/**', '!assets/css/app/bourbon/**'],

  js: ['assets/js/**/*.js', 'assets/js/**/*.coffee', 'assets/js/**/*.cjsx'],

  // woff support base64 encoding by default. Otherwise, there's ttf2woff
  // converters.
  fonts: [
    'assets/fonts/data/*',
    'node_modules/bootstrap/fonts/*',
    'node_modules/font-awesome/fonts/*'
  ],

  // Supporting css classes for icons fonts. Glyphicons are part of our bootstrap
  // build.
  font_styles: [ 'assets/fonts/styles/*', 'node_modules/font-awesome/scss/font-awesome.scss'],

  // This sets browserify off on it's dependency resolution.
  entrypoint: './assets/js/app/app.coffee'
};

// UTILITY
function banner()
{
  return '/*! '+package.name+' - v'+package.version+' - '+gutil.date(new Date(), "yyyy-mm-dd")+
          ' [copyright: '+package.copyright+']'+' */';
}

function logData(data) {
  gutil.log(
    gutil.colors.blue(
      gutil.colors.bold(data)
    )
  );
}

function ready() {
  gutil.log(
    gutil.colors.bgMagenta(
      gutil.colors.red(
        gutil.colors.bold('[          STATUS: READY          ]')
      )
    )
  );
}

logData('Environment : '+ (production ? 'Production':'Development'));

// Javascript Related
// -----------

function onError(source, err) {
  notify.onError({
    title: source,
    subtitle: "Failure!",
    message:  "Error: <%= err.message %>"
  })(err);

  this.emit('end');
}

function bundlefile(file, as, watch) {
  var props = watchify.args;
  props.entries = [file];
  props.debug = true;
  props.extensions = ['.coffee', '.cjsx'];
  var bundler = watch ? watchify(browserify(props)) : browserify(props);
  bundler
    .transform(coffee_reactify);

  function rebundle() {
    var stream = bundler.bundle();
    return stream
      .pipe(plumber({errorHandler: onError.bind('JS Error')}))
      .on('error', gutil.log.bind(gutil, 'error'))
      .pipe(source(as))
      .pipe(gulp.dest('static/js'))
      .pipe(browserSync.reload({stream:true, once: true}));
  }

  bundler.on('update', function() {
    rebundle();
    gutil.log('Rebundle...');
  })
  .on('exit', function() {
    gutil.log('Done...');
  });

  return rebundle();
}

gulp.task('js:app', function() {
  return bundlefile(paths.entrypoint, 'app.js', false);
});

gulp.task('js:app:watch', function() {
  return bundlefile(paths.entrypoint, 'app.js', true);
});

gulp.task('uglify', function() {
  return gulp.src('static/js/app.js')
  .pipe(uglify('build.min.js', {
    preserveComments: false,
    compress: {
      warnings: false
    }
  }))
  .pipe(gulp.dest('static/js'));
});

// CSS Related
// -----------

function sassify(src, alias, dest) {
  return function() {
    return gulp.src(src)
      .pipe(sourcemaps.init())
      .pipe(
        sass({
          errLogToConsole: true,
          sourceComments: 'map',
          onError: notify.onError({
            title: "Compile Error",
            message: "<%= error.message %>"
          })
      }))
      // .pipe(autoprefixer('last 2 versions', '> 1%', 'ie 9'))
      .pipe(sourcemaps.write())
      .pipe(rename(alias))
      .pipe(gulp.dest(dest))
      .pipe(browserSync.reload({stream:true}));
    };
}

gulp.task('app:sass', sassify('./assets/css/app/main.scss', 'app.css', 'static/css'));
gulp.task('styleguide:sass', sassify('./assets/css/styleguide/styleguide.scss', 'styleguide.css', 'static/css'));

gulp.task('minifycss', function() {
  gutil.log("running minfiy?");
  return gulp.src(['static/css/*.css'])
    .pipe(minifycss())
    .pipe(gulp.dest('static/css'));
});

gulp.task('bless', function() {
  return gulp.src('static/css/*.css')
    .pipe(bless())
    .pipe(insert.prepend('@charset "UTF-8";\n'))
    .pipe(gulp.dest('static/css/app/blessed'));
});

// Compile sass in parallel.
gulp.task('sass', ['app:sass', 'styleguide:sass']);

// gulp.task('bless', ['app:bless']);

gulp.task('build:css', ['fonts', 'sass']);

gulp.task('build:js', ['js:app']);
gulp.task('build', ['build:css', 'build:js']);

// Fonts
// -----

gulp.task('fonts:clean', function() {
  return gulp.src('.build/css/fonts/*', {read: false}).pipe(clean());
});

// We base64 encode our fonts.
gulp.task('fonts:copy', ['fonts:clean'], function() {
  return gulp.src(paths.fonts)
  .pipe(gulp.dest('static/fonts'));
});

// Corresponding font defintions e.g. .fa-circle-plus
gulp.task('fonts:sass', ['fonts:clean'], sassify(paths.font_styles, 'fonts.css', '.build/css/fonts'));
gulp.task('fonts:css', ['fonts:clean'], function() {
  return gulp
    .src(paths.font_styles)
    .pipe(filter('*.css'))
    .pipe(gulp.dest('.build/css/fonts'));
});

gulp.task('fonts:concat', ['fonts:sass', 'fonts:css'], function(){
  return gulp
    .src('.build/css/fonts/*')
    .pipe(concat('fonts.css'))
    .pipe(gulp.dest('static/css'));
});

// Combine font definitions and font data into a single file.
gulp.task('fonts', ['fonts:copy', 'fonts:concat']);

gulp.task('dist', function(cb) {
  runSequence('build', 'minifycss', 'uglify', function() {
    cb();
    gutil.log(gutil.colors.green("Build complete"));
  });
});

// Testing
//--------

gulp.task('test', function(done) {
  jest.runCLI({ config: package.jest }, ".", function(result) {
    if (result === false) {
      notifier.notify({
        title: "Fail!",
        message: "Javascript tests failed.",
      });
    }
    done();
  });
});

// Watch Tasks
//------------

gulp.task('build:css:watch', ['sass'], ready);
gulp.task('build:js:watch', ['js:app:watch'], ready);
gulp.task('react-bootstrap:watch', ['react-bootstrap'], ready);

gulp.task('watch:css', function() {
  gulp.watch(paths.watch_scss, ['build:css:watch']);
});

gulp.task('watch:test', function() {
  gulp.watch(paths.js, ['test']);
});

gulp.task('watch', ['watch:css', 'build:js:watch', 'browser-sync']);

gulp.task('browser-sync', function() {
  browserSync({
    proxy: "app.dev.bursa.io, dev.bursa.io",
    debug: true
  });
});
