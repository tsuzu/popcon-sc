
var express = require('express')
,   http = require('http')
;

var app = express();

app.configure(function() {
    app.set('port', process.env.PORT || 3333)
    app.use(express.bodyParser());
    app.use(express.methodOverride());

    app.use(app.router);
    app.use(express.static(__dirname + "/dist"));
});

http.createServer(app).listen(app.get('port'), function() {
    console.log("Express server listening on port " + app.get('port'));
});