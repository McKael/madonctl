/*
Copyright 2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sendgrid/rest"
)

// GetList returns a List entity
func (mc *Client) GetList(listID ActivityID) (*List, error) {
	if listID == "" {
		return nil, errors.New("invalid list ID")
	}
	endPoint := "lists/" + listID
	var list List
	if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, nil, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetLists returns a list of List entities
// If accountID is > 0, this will return the lists containing this account.
// If lopt.All is true, several requests will be made until the API server
// has nothing to return.
func (mc *Client) GetLists(accountID ActivityID, lopt *LimitParams) ([]List, error) {
	endPoint := "lists"

	if accountID != "" {
		endPoint = "accounts/" + accountID + "/lists"
	}

	var lists []List
	var links apiLinks
	if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, lopt, &links, &lists); err != nil {
		return nil, err
	}
	if lopt != nil { // Fetch more pages to reach our limit
		for (lopt.All || lopt.Limit > len(lists)) && links.next != nil {
			listSlice := []List{}
			newlopt := links.next
			links = apiLinks{}
			if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, newlopt, &links, &listSlice); err != nil {
				return nil, err
			}
			lists = append(lists, listSlice...)
		}
	}
	return lists, nil
}

// CreateList creates a List
func (mc *Client) CreateList(title string) (*List, error) {
	params := apiCallParams{"title": title}
	method := rest.Post
	return mc.setSingleList(method, "", params)
}

// UpdateList updates an existing List
func (mc *Client) UpdateList(listID ActivityID, title string) (*List, error) {
	if listID == "" {
		return nil, errors.New("invalid list ID")
	}
	params := apiCallParams{"title": title}
	method := rest.Put
	return mc.setSingleList(method, listID, params)
}

// DeleteList deletes a list
func (mc *Client) DeleteList(listID ActivityID) error {
	if listID == "" {
		return errors.New("invalid list ID")
	}
	method := rest.Delete
	_, err := mc.setSingleList(method, listID, nil)
	return err
}

// GetListAccounts returns the accounts belonging to a given list
func (mc *Client) GetListAccounts(listID ActivityID, lopt *LimitParams) ([]Account, error) {
	endPoint := "lists/" + listID + "/accounts"
	return mc.getMultipleAccounts(endPoint, nil, lopt)
}

// AddListAccounts adds the accounts to a given list
func (mc *Client) AddListAccounts(listID ActivityID, accountIDs []ActivityID) error {
	endPoint := "lists/" + listID + "/accounts"
	method := rest.Post
	params := make(apiCallParams)
	for i, id := range accountIDs {
		if id == "" {
			return ErrInvalidID
		}
		qID := fmt.Sprintf("[%d]account_ids", i)
		params[qID] = id
	}
	return mc.apiCall("v1/"+endPoint, method, params, nil, nil, nil)
}

// RemoveListAccounts removes the accounts from the given list
func (mc *Client) RemoveListAccounts(listID ActivityID, accountIDs []ActivityID) error {
	endPoint := "lists/" + listID + "/accounts"
	method := rest.Delete
	params := make(apiCallParams)
	for i, id := range accountIDs {
		if id == "" {
			return ErrInvalidID
		}
		qID := fmt.Sprintf("[%d]account_ids", i)
		params[qID] = id
	}
	return mc.apiCall("v1/"+endPoint, method, params, nil, nil, nil)
}

func (mc *Client) setSingleList(method rest.Method, listID ActivityID, params apiCallParams) (*List, error) {
	var endPoint string
	if listID != "" {
		endPoint = "lists/" + listID
	} else {
		endPoint = "lists"
	}
	var list List
	if err := mc.apiCall("v1/"+endPoint, method, params, nil, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}
