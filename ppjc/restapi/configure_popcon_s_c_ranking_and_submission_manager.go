package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	graceful "github.com/tylerb/graceful"

	"github.com/cs3238-tsuzu/popcon-sc/ranking/restapi/swagger"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target .. --name  --spec ../../../../../../../../../../tmp/swagger.yaml --api-package swagger

func configureFlags(api *swagger.PopconSCRankingAndSubmissionManagerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *swagger.PopconSCRankingAndSubmissionManagerAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "X-Auth-Token" header is set
	api.AuthAuth = func(token string) (interface{}, error) {
		return nil, errors.NotImplemented("api key auth (auth) X-Auth-Token from header param [X-Auth-Token] has not yet been implemented")
	}

	api.GetContestsCidJoinUIDHandler = swagger.GetContestsCidJoinUIDHandlerFunc(func(params swagger.GetContestsCidJoinUIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .GetContestsCidJoinUID has not yet been implemented")
	})
	api.GetContestsCidRankingHandler = swagger.GetContestsCidRankingHandlerFunc(func(params swagger.GetContestsCidRankingParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .GetContestsCidRanking has not yet been implemented")
	})
	api.PostContestsCidJoinUIDHandler = swagger.PostContestsCidJoinUIDHandlerFunc(func(params swagger.PostContestsCidJoinUIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostContestsCidJoinUID has not yet been implemented")
	})
	api.PostContestsCidSubmissionResultHandler = swagger.PostContestsCidSubmissionResultHandlerFunc(func(params swagger.PostContestsCidSubmissionResultParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostContestsCidSubmissionResult has not yet been implemented")
	})
	api.PostContestsCidUpdateHandler = swagger.PostContestsCidUpdateHandlerFunc(func(params swagger.PostContestsCidUpdateParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostContestsCidUpdate has not yet been implemented")
	})
	api.PostContestsCreateCidHandler = swagger.PostContestsCreateCidHandlerFunc(func(params swagger.PostContestsCreateCidParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostContestsCreateCid has not yet been implemented")
	})
	api.PostContestsRemoveCidHandler = swagger.PostContestsRemoveCidHandlerFunc(func(params swagger.PostContestsRemoveCidParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostContestsRemoveCid has not yet been implemented")
	})
	api.PostShutdownHandler = swagger.PostShutdownHandlerFunc(func(params swagger.PostShutdownParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostShutdown has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
