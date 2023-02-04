/*
Copyright 2017-2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"fmt"

	"github.com/sendgrid/rest"
)

// GetReports returns the current user's reports
// (I don't know if the limit options are used by the API server.)
func (mc *Client) GetReports(lopt *LimitParams) ([]Report, error) {
	var reports []Report
	if err := mc.apiCall("v1/reports", rest.Get, nil, lopt, nil, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// ReportUser reports the user account
func (mc *Client) ReportUser(accountID ActivityID, statusIDs []ActivityID, comment string) (*Report, error) {
	if accountID == "" || comment == "" || len(statusIDs) < 1 {
		return nil, ErrInvalidParameter
	}

	params := make(apiCallParams)
	params["account_id"] = accountID
	params["comment"] = comment
	for i, id := range statusIDs {
		if id == "" {
			return nil, ErrInvalidID
		}
		qID := fmt.Sprintf("[%d]status_ids", i)
		params[qID] = id
	}

	var report Report
	if err := mc.apiCall("v1/reports", rest.Post, params, nil, nil, &report); err != nil {
		return nil, err
	}
	return &report, nil
}
