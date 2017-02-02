package swagger

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/cs3238-tsuzu/popcon-sc/ranking/models"
)

// NewPostContestsCidSubmissionResultParams creates a new PostContestsCidSubmissionResultParams object
// with the default values initialized.
func NewPostContestsCidSubmissionResultParams() PostContestsCidSubmissionResultParams {
	var ()
	return PostContestsCidSubmissionResultParams{}
}

// PostContestsCidSubmissionResultParams contains all the bound params for the post contests cid submission result operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostContestsCidSubmissionResult
type PostContestsCidSubmissionResultParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*Contest ID
	  Required: true
	  In: path
	*/
	Cid int64
	/*
	  Required: true
	  In: body
	*/
	SubmissionResult *models.SubmissionResult
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *PostContestsCidSubmissionResultParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	rCid, rhkCid, _ := route.Params.GetOK("cid")
	if err := o.bindCid(rCid, rhkCid, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.SubmissionResult
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("submissionResult", "body"))
			} else {
				res = append(res, errors.NewParseError("submissionResult", "body", "", err))
			}

		} else {
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.SubmissionResult = &body
			}
		}

	} else {
		res = append(res, errors.Required("submissionResult", "body"))
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *PostContestsCidSubmissionResultParams) bindCid(rawData []string, hasKey bool, formats strfmt.Registry) error {
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