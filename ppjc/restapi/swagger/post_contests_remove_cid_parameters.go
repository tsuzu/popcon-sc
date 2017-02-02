package swagger

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPostContestsRemoveCidParams creates a new PostContestsRemoveCidParams object
// with the default values initialized.
func NewPostContestsRemoveCidParams() PostContestsRemoveCidParams {
	var ()
	return PostContestsRemoveCidParams{}
}

// PostContestsRemoveCidParams contains all the bound params for the post contests remove cid operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostContestsRemoveCid
type PostContestsRemoveCidParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*Contest ID
	  Required: true
	  In: path
	*/
	Cid int64
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *PostContestsRemoveCidParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	rCid, rhkCid, _ := route.Params.GetOK("cid")
	if err := o.bindCid(rCid, rhkCid, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *PostContestsRemoveCidParams) bindCid(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("cid", "path", "int64", raw)
	}
	o.Cid = value

	return nil
}