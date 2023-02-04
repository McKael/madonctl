/*
Copyright 2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"github.com/sendgrid/rest"
)

// GetSuggestions returns a list of follow suggestions from the server
func (mc *Client) GetSuggestions(lopt *LimitParams) ([]Account, error) {
	endPoint := "suggestions"
	method := rest.Get
	var accountList []Account
	if err := mc.apiCall("v1/"+endPoint, method, nil, lopt, nil, &accountList); err != nil {
		return nil, err
	}
	return accountList, nil
}

// DeleteSuggestion removes the account from the suggestion list
func (mc *Client) DeleteSuggestion(accountID ActivityID) error {
	endPoint := "suggestions/" + accountID
	method := rest.Delete
	return mc.apiCall("v1/"+endPoint, method, nil, nil, nil, nil)
}
