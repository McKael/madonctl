/*
Copyright 2017-2018 Mikael Berthe
Copyright 2017 Ollivier Robert

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sendgrid/rest"
)

type apiLinks struct {
	next, prev *LimitParams
}

func parseLink(links []string) (*apiLinks, error) {
	if len(links) == 0 {
		return nil, nil
	}

	al := new(apiLinks)
	linkRegex := regexp.MustCompile(`<([^>]+)>; rel="([^"]+)`)
	for _, l := range links {
		m := linkRegex.FindAllStringSubmatch(l, -1)
		for _, submatch := range m {
			if len(submatch) != 3 {
				continue
			}
			// Parse URL
			u, err := url.Parse(submatch[1])
			if err != nil {
				return al, err
			}
			var lp *LimitParams
			since := u.Query().Get("since_id")
			max := u.Query().Get("max_id")
			lim := u.Query().Get("limit")
			if since == "" && max == "" {
				continue
			}
			lp = new(LimitParams)
			if since != "" {
				lp.SinceID = since
				if err != nil {
					return al, err
				}
			}
			if max != "" {
				lp.MaxID = max
				if err != nil {
					return al, err
				}
			}
			if lim != "" {
				lp.Limit, err = strconv.Atoi(lim)
				if err != nil {
					return al, err
				}
			}
			switch submatch[2] {
			case "prev":
				al.prev = lp
			case "next":
				al.next = lp
			}
		}
	}
	return al, nil
}

// restAPI actually does the HTTP query
// It is a copy of rest.API with better handling of parameters with multiple values
func restAPI(request rest.Request) (*rest.Response, error) {
	// Our encoded parameters
	var urlpstr string

	c := &rest.Client{HTTPClient: http.DefaultClient}

	// Build the HTTP request object.
	if len(request.QueryParams) != 0 {
		urlp := url.Values{}
		arrayRe := regexp.MustCompile(`^\[\d+\](.*)$`)
		for key, value := range request.QueryParams {
			// It seems Mastodon doesn't like parameters with index
			// numbers, but it needs the brackets.
			// Let's check if the key matches '^.+\[.*\]$'
			// Do not proceed if there's another bracket pair.
			klen := len(key)
			if klen == 0 {
				continue
			}
			if m := arrayRe.FindStringSubmatch(key); len(m) > 0 {
				// This is an array, let's remove the index number
				key = m[1] + "[]"
			}
			urlp.Add(key, value)
		}
		urlpstr = urlp.Encode()
	}

	switch request.Method {
	case "GET":
		// Add parameters to the URL if we have any.
		if len(urlpstr) > 0 {
			request.BaseURL += "?" + urlpstr
		}
	default:
		// Pleroma at least needs the API parameters in the body rather than
		// the URL for `POST` requests.  Which is fair according to
		// https://html.spec.whatwg.org/multipage/form-control-infrastructure.html#form-submission-2
		// which suggests that `GET` requests should have URL parameters
		// and `POST` requests should have the encoded parameters in the body.
		//
		// HOWEVER for file uploads, we've already got a properly encoded body
		// which means we ignore this step.
		if len(request.Body) == 0 {
			request.Body = []byte(urlpstr)
		}
	}

	req, err := http.NewRequest(string(request.Method), request.BaseURL, bytes.NewBuffer(request.Body))
	if err != nil {
		return nil, err
	}

	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	_, exists := req.Header["Content-Type"]
	if len(request.Body) > 0 && !exists {
		// Make sure we have the correct content type for form submission.
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Build the HTTP client and make the request.
	res, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var errorText string
		// Try to unmarshal the returned error object for a description
		mastodonError := Error{}
		decodeErr := json.NewDecoder(res.Body).Decode(&mastodonError)
		if decodeErr != nil {
			// Decode unsuccessful, fallback to generic error based on response code
			errorText = http.StatusText(res.StatusCode)
		} else {
			errorText = mastodonError.Text
		}

		// Please note that the error string code is used by Search()
		// to check the error cause.
		const errFormatString = "bad server status code (%d)"
		return nil, errors.Errorf(errFormatString+": %s",
			res.StatusCode, errorText)
	}

	// Build Response object.
	response, err := rest.BuildResponse(res)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// prepareRequest inserts all pre-defined stuff
func (mc *Client) prepareRequest(target string, method rest.Method, params apiCallParams) (rest.Request, error) {
	var req rest.Request

	if mc == nil {
		return req, ErrUninitializedClient
	}

	endPoint := mc.APIBase + "/" + target

	// Request headers
	hdrs := make(map[string]string)
	hdrs["User-Agent"] = fmt.Sprintf("madon/%s", MadonVersion)
	if mc.UserToken != nil {
		hdrs["Authorization"] = fmt.Sprintf("Bearer %s", mc.UserToken.AccessToken)
	}

	req = rest.Request{
		BaseURL:     endPoint,
		Headers:     hdrs,
		Method:      method,
		QueryParams: params,
	}
	return req, nil
}

// apiCall makes a call to the Mastodon API server
// If links is not nil, the prev/next links from the API response headers
// will be set (if they exist) in the structure.
func (mc *Client) apiCall(endPoint string, method rest.Method, params apiCallParams, limitOptions *LimitParams, links *apiLinks, data interface{}) error {
	if mc == nil {
		return errors.New("use of uninitialized madon client")
	}

	if limitOptions != nil {
		if params == nil {
			params = make(apiCallParams)
		}
		if limitOptions.Limit > 0 {
			params["limit"] = strconv.Itoa(limitOptions.Limit)
		}
		if limitOptions.SinceID != "" {
			params["since_id"] = limitOptions.SinceID
		}
		if limitOptions.MaxID != "" {
			params["max_id"] = limitOptions.MaxID
		}
	}

	// Prepare query
	req, err := mc.prepareRequest(endPoint, method, params)
	if err != nil {
		return err
	}

	// Make API call
	r, err := restAPI(req)
	if err != nil {
		return errors.Wrapf(err, "API query (%s) failed", endPoint)
	}

	if links != nil {
		pLinks, err := parseLink(r.Headers["Link"])
		if err != nil {
			return errors.Wrapf(err, "cannot decode header links (%s)", method)
		}
		if pLinks != nil {
			*links = *pLinks
		}
	}

	// Check for error reply
	var errorResult Error
	if err := json.Unmarshal([]byte(r.Body), &errorResult); err == nil {
		// The empty object is not an error
		if errorResult.Text != "" {
			return errors.New(errorResult.Text)
		}
	}

	// Not an error reply; let's unmarshal the data
	err = json.Unmarshal([]byte(r.Body), &data)
	if err != nil {
		return errors.Wrapf(err, "cannot decode API response (%s)", method)
	}
	return nil
}

/* Mastodon timestamp handling */

// MastodonDate is a custom type for the timestamps returned by some API calls
// It is used, for example, by 'v1/instance/activity' and 'v2/search'.
// The date returned by those Mastodon API calls is a string containing a
// timestamp in seconds...

// UnmarshalJSON handles deserialization for custom MastodonDate type
func (act *MastodonDate) UnmarshalJSON(b []byte) error {
	s, err := strconv.ParseInt(strings.Trim(string(b), "\""), 10, 64)
	if err != nil {
		return err
	}
	if s == 0 {
		act.Time = time.Time{}
		return nil
	}
	act.Time = time.Unix(s, 0)
	return nil
}

// MarshalJSON handles serialization for custom MastodonDate type
func (act *MastodonDate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d\"", act.Unix())), nil
}
