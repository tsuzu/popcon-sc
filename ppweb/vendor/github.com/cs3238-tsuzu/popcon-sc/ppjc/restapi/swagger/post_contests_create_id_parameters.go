package swagger

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPostContestsCreateIDParams creates a new PostContestsCreateIDParams object
// with the default values initialized.
func NewPostContestsCreateIDParams() PostContestsCreateIDParams {
	var ()
	return PostContestsCreateIDParams{}
}

// PostContestsCreateIDParams contains all the bound params for the post contests create ID operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostContestsCreateID
type PostContestsCreateIDParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*Contest Finish Time
	  Required: true
	  In: query
	*/
	FinishTime string
	/*Contest ID
	  Required: true
	  In: path
	*/
	ID int64
	/*Contest Start Time
	  Required: true
	  In: query
	*/
	StartTime string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *PostContestsCreateIDParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qFinishTime, qhkFinishTime, _ := qs.GetOK("finishTime")
	if err := o.bindFinishTime(qFinishTime, qhkFinishTime, route.Formats); err != nil {
		res = append(res, err)
	}

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	qStartTime, qhkStartTime, _ := qs.GetOK("startTime")
	if err := o.bindStartTime(qStartTime, qhkStartTime, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *PostContestsCreateIDParams) bindFinishTime(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("finishTime", "query")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if err := validate.RequiredString("finishTime", "query", raw); err != nil {
		return err
	}

	o.FinishTime = raw

	return nil
}

func (o *PostContestsCreateIDParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("id", "path", "int64", raw)
	}
	o.ID = value

	return nil
}

func (o *PostContestsCreateIDParams) bindStartTime(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("startTime", "query")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if err := validate.RequiredString("startTime", "query", raw); err != nil {
		return err
	}

	o.StartTime = raw

	return nil
}