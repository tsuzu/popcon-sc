package swagger

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"
	"strings"

	errors "github.com/go-openapi/errors"
	loads "github.com/go-openapi/loads"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	security "github.com/go-openapi/runtime/security"
	spec "github.com/go-openapi/spec"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewPopconSCRankingAndSubmissionManagerAPI creates a new PopconSCRankingAndSubmissionManager instance
func NewPopconSCRankingAndSubmissionManagerAPI(spec *loads.Document) *PopconSCRankingAndSubmissionManagerAPI {
	return &PopconSCRankingAndSubmissionManagerAPI{
		handlers:        make(map[string]map[string]http.Handler),
		formats:         strfmt.Default,
		defaultConsumes: "application/json",
		defaultProduces: "application/json",
		ServerShutdown:  func() {},
		spec:            spec,
		ServeError:      errors.ServeError,
		JSONConsumer:    runtime.JSONConsumer(),
		JSONProducer:    runtime.JSONProducer(),
		GetContestsCidJoinUIDHandler: GetContestsCidJoinUIDHandlerFunc(func(params GetContestsCidJoinUIDParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation GetContestsCidJoinUID has not yet been implemented")
		}),
		GetContestsCidRankingHandler: GetContestsCidRankingHandlerFunc(func(params GetContestsCidRankingParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation GetContestsCidRanking has not yet been implemented")
		}),
		PostContestsCidJoinUIDHandler: PostContestsCidJoinUIDHandlerFunc(func(params PostContestsCidJoinUIDParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostContestsCidJoinUID has not yet been implemented")
		}),
		PostContestsCidSubmissionResultHandler: PostContestsCidSubmissionResultHandlerFunc(func(params PostContestsCidSubmissionResultParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostContestsCidSubmissionResult has not yet been implemented")
		}),
		PostContestsCidUpdateHandler: PostContestsCidUpdateHandlerFunc(func(params PostContestsCidUpdateParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostContestsCidUpdate has not yet been implemented")
		}),
		PostContestsCreateCidHandler: PostContestsCreateCidHandlerFunc(func(params PostContestsCreateCidParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostContestsCreateCid has not yet been implemented")
		}),
		PostContestsRemoveCidHandler: PostContestsRemoveCidHandlerFunc(func(params PostContestsRemoveCidParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostContestsRemoveCid has not yet been implemented")
		}),
		PostShutdownHandler: PostShutdownHandlerFunc(func(params PostShutdownParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation PostShutdown has not yet been implemented")
		}),

		// Applies when the "X-Auth-Token" header is set
		AuthAuth: func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (auth) X-Auth-Token from header param [X-Auth-Token] has not yet been implemented")
		},
	}
}

/*PopconSCRankingAndSubmissionManagerAPI the popcon s c ranking and submission manager API */
type PopconSCRankingAndSubmissionManagerAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler
	// JSONConsumer registers a consumer for a "application/json" mime type
	JSONConsumer runtime.Consumer

	// JSONProducer registers a producer for a "application/json" mime type
	JSONProducer runtime.Producer

	// AuthAuth registers a function that takes a token and returns a principal
	// it performs authentication based on an api key X-Auth-Token provided in the header
	AuthAuth func(string) (interface{}, error)

	// GetContestsCidJoinUIDHandler sets the operation handler for the get contests cid join UID operation
	GetContestsCidJoinUIDHandler GetContestsCidJoinUIDHandler
	// GetContestsCidRankingHandler sets the operation handler for the get contests cid ranking operation
	GetContestsCidRankingHandler GetContestsCidRankingHandler
	// PostContestsCidJoinUIDHandler sets the operation handler for the post contests cid join UID operation
	PostContestsCidJoinUIDHandler PostContestsCidJoinUIDHandler
	// PostContestsCidSubmissionResultHandler sets the operation handler for the post contests cid submission result operation
	PostContestsCidSubmissionResultHandler PostContestsCidSubmissionResultHandler
	// PostContestsCidUpdateHandler sets the operation handler for the post contests cid update operation
	PostContestsCidUpdateHandler PostContestsCidUpdateHandler
	// PostContestsCreateCidHandler sets the operation handler for the post contests create cid operation
	PostContestsCreateCidHandler PostContestsCreateCidHandler
	// PostContestsRemoveCidHandler sets the operation handler for the post contests remove cid operation
	PostContestsRemoveCidHandler PostContestsRemoveCidHandler
	// PostShutdownHandler sets the operation handler for the post shutdown operation
	PostShutdownHandler PostShutdownHandler

	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// SetDefaultProduces sets the default produces media type
func (o *PopconSCRankingAndSubmissionManagerAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *PopconSCRankingAndSubmissionManagerAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *PopconSCRankingAndSubmissionManagerAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *PopconSCRankingAndSubmissionManagerAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *PopconSCRankingAndSubmissionManagerAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *PopconSCRankingAndSubmissionManagerAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *PopconSCRankingAndSubmissionManagerAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the PopconSCRankingAndSubmissionManagerAPI
func (o *PopconSCRankingAndSubmissionManagerAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}

	if o.AuthAuth == nil {
		unregistered = append(unregistered, "XAuthTokenAuth")
	}

	if o.GetContestsCidJoinUIDHandler == nil {
		unregistered = append(unregistered, "GetContestsCidJoinUIDHandler")
	}

	if o.GetContestsCidRankingHandler == nil {
		unregistered = append(unregistered, "GetContestsCidRankingHandler")
	}

	if o.PostContestsCidJoinUIDHandler == nil {
		unregistered = append(unregistered, "PostContestsCidJoinUIDHandler")
	}

	if o.PostContestsCidSubmissionResultHandler == nil {
		unregistered = append(unregistered, "PostContestsCidSubmissionResultHandler")
	}

	if o.PostContestsCidUpdateHandler == nil {
		unregistered = append(unregistered, "PostContestsCidUpdateHandler")
	}

	if o.PostContestsCreateCidHandler == nil {
		unregistered = append(unregistered, "PostContestsCreateCidHandler")
	}

	if o.PostContestsRemoveCidHandler == nil {
		unregistered = append(unregistered, "PostContestsRemoveCidHandler")
	}

	if o.PostShutdownHandler == nil {
		unregistered = append(unregistered, "PostShutdownHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *PopconSCRankingAndSubmissionManagerAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *PopconSCRankingAndSubmissionManagerAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {

	result := make(map[string]runtime.Authenticator)
	for name, scheme := range schemes {
		switch name {

		case "auth":

			result[name] = security.APIKeyAuth(scheme.Name, scheme.In, o.AuthAuth)

		}
	}
	return result

}

// ConsumersFor gets the consumers for the specified media types
func (o *PopconSCRankingAndSubmissionManagerAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {

	result := make(map[string]runtime.Consumer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONConsumer

		}
	}
	return result

}

// ProducersFor gets the producers for the specified media types
func (o *PopconSCRankingAndSubmissionManagerAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {

	result := make(map[string]runtime.Producer)
	for _, mt := range mediaTypes {
		switch mt {

		case "application/json":
			result["application/json"] = o.JSONProducer

		}
	}
	return result

}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *PopconSCRankingAndSubmissionManagerAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the popcon s c ranking and submission manager API
func (o *PopconSCRankingAndSubmissionManagerAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *PopconSCRankingAndSubmissionManagerAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened

	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["GET"] == nil {
		o.handlers[strings.ToUpper("GET")] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/contests/{cid}/join/{uid}"] = NewGetContestsCidJoinUID(o.context, o.GetContestsCidJoinUIDHandler)

	if o.handlers["GET"] == nil {
		o.handlers[strings.ToUpper("GET")] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/contests/{cid}/ranking"] = NewGetContestsCidRanking(o.context, o.GetContestsCidRankingHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/contests/{cid}/join/{uid}"] = NewPostContestsCidJoinUID(o.context, o.PostContestsCidJoinUIDHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/contests/{cid}/submission_result"] = NewPostContestsCidSubmissionResult(o.context, o.PostContestsCidSubmissionResultHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/contests/{cid}/update"] = NewPostContestsCidUpdate(o.context, o.PostContestsCidUpdateHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/contests/create/{cid}"] = NewPostContestsCreateCid(o.context, o.PostContestsCreateCidHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/contests/remove/{cid}"] = NewPostContestsRemoveCid(o.context, o.PostContestsRemoveCidHandler)

	if o.handlers["POST"] == nil {
		o.handlers[strings.ToUpper("POST")] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/shutdown"] = NewPostShutdown(o.context, o.PostShutdownHandler)

}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *PopconSCRankingAndSubmissionManagerAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middelware as you see fit
func (o *PopconSCRankingAndSubmissionManagerAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}