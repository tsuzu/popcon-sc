package sctypes

import (
	"net/http"
)

var InternalHTTPToken = "Popcon-Authentication"

// BR400 is "400 Bad Request"
const BR400 = `
<!DOCTYPE html>
<html>
<head>
<title>
400 Bad Request
</title>
</head>
<body>
<h1>400 Bad Request</h1>
Received an illegal request.
</body>
</html>
`

// FBD403 is "403 Forbidden"
const FBD403 = `
<!DOCTYPE html>
<html>
<head>
<title>
403 Forbidden
</title>
</head>
<body>
<h1>403 Forbidden</h1>
You can't see this page.
</body>
</html>
`

// NF404 is "404 Not Found"
const NF404 = `
<!DOCTYPE html>
<html>
<head>
<title>
404 Not Found
</title>
</head>
<body>
<h1>404 Not Found</h1>
The page is not found in this server.
</body>
</html>
`

// RETL413 is "413 Request Entity Too Large"
const RETL413 = `
<!DOCTYPE html>
<html>
<head>
<title>
413 Request Entity Too Large
</title>
</head>
<body>
<h1>Request Entity Too Large</h1>
Your file is too large.(>10MB)
</body>
</html>
`

// ISE500 is "500 Internal Server Error"
const ISE500 = `
<!DOCTYPE html>
<html>
<head>
<title>
500 Internal Server Error
</title>
</head>
<body>
<h1>500 Internal Server Error</h1>
Some errors occured in this server.
</body>
</html>
`

// NI501 is "501 Not Implemented"
const NI501 = `
<!DOCTYPE html>
<html>
<head>
<title>
501 Not Implemented
</title>
</head>
<body>
<h1>501 Not Implemented</h1>
The service is not implemented.
</body>
</html>
`

var ResponseTemplateMap = map[int]string{
	http.StatusBadRequest:            BR400,
	http.StatusForbidden:             FBD403,
	http.StatusNotFound:              NF404,
	http.StatusInternalServerError:   ISE500,
	http.StatusNotImplemented:        NI501,
	http.StatusRequestEntityTooLarge: RETL413,
}

func ResponseTemplateWrite(status int, rw http.ResponseWriter) {
	if data, ok := ResponseTemplateMap[status]; ok {
		rw.WriteHeader(status)
		rw.Write([]byte(data))
	} else {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(ISE500))
	}
}
