/*
Copyright 2017-2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sendgrid/rest"
)

// getAccountsOptions contains option fields for POST and DELETE API calls
type getAccountsOptions struct {
	// The ID is used for most commands
	ID ActivityID

	// Following can be set to true to limit a search to "following" accounts
	Following bool

	// The Q field (query) is used when searching for accounts
	Q string

	Limit *LimitParams
}

// UpdateAccountParams contains option fields for the UpdateAccount command
type UpdateAccountParams struct {
	DisplayName      *string
	Note             *string
	AvatarImagePath  *string
	HeaderImagePath  *string
	Locked           *bool
	Bot              *bool
	FieldsAttributes *[]Field
	Source           *SourceParams
}

// updateRelationship returns a Relationship entity
// The operation 'op' can be "follow", "unfollow", "block", "unblock",
// "mute", "unmute".
// The id is optional and depends on the operation.
func (mc *Client) updateRelationship(op string, id ActivityID, params apiCallParams) (*Relationship, error) {
	var endPoint string
	method := rest.Post

	switch op {
	case "follow", "unfollow", "block", "unblock", "mute", "unmute", "pin", "unpin":
		endPoint = "accounts/" + id + "/" + op
	default:
		return nil, ErrInvalidParameter
	}

	var rel Relationship
	if err := mc.apiCall("v1/"+endPoint, method, params, nil, nil, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// getSingleAccount returns an account entity
// The operation 'op' can be "account", "verify_credentials",
// "follow_requests/authorize" or // "follow_requests/reject".
// The id is optional and depends on the operation.
func (mc *Client) getSingleAccount(op string, id ActivityID) (*Account, error) {
	var endPoint string
	method := rest.Get

	switch op {
	case "account":
		endPoint = "accounts/" + id
	case "verify_credentials":
		endPoint = "accounts/verify_credentials"
	case "follow_requests/authorize", "follow_requests/reject":
		// The documentation is incorrect, the endpoint actually
		// is "follow_requests/:id/{authorize|reject}"
		endPoint = op[:16] + id + "/" + op[16:]
		method = rest.Post
	default:
		return nil, ErrInvalidParameter
	}

	var account Account
	if err := mc.apiCall("v1/"+endPoint, method, nil, nil, nil, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

// getMultipleAccounts returns a list of account entities
// If lopt.All is true, several requests will be made until the API server
// has nothing to return.
func (mc *Client) getMultipleAccounts(endPoint string, params apiCallParams, lopt *LimitParams) ([]Account, error) {
	var accounts []Account
	var links apiLinks
	if err := mc.apiCall("v1/"+endPoint, rest.Get, params, lopt, &links, &accounts); err != nil {
		return nil, err
	}
	if lopt != nil { // Fetch more pages to reach our limit
		for (lopt.All || lopt.Limit > len(accounts)) && links.next != nil {
			accountSlice := []Account{}
			newlopt := links.next
			links = apiLinks{}
			if err := mc.apiCall("v1/"+endPoint, rest.Get, params, newlopt, &links, &accountSlice); err != nil {
				return nil, err
			}
			accounts = append(accounts, accountSlice...)
		}
	}
	return accounts, nil
}

// getMultipleAccountsHelper returns a list of account entities
// The operation 'op' can be "followers", "following", "search", "blocks",
// "mutes", "follow_requests".
// The id is optional and depends on the operation.
// If opts.All is true, several requests will be made until the API server
// has nothing to return.
func (mc *Client) getMultipleAccountsHelper(op string, opts *getAccountsOptions) ([]Account, error) {
	var endPoint string
	var lopt *LimitParams

	if opts != nil {
		lopt = opts.Limit
	}

	switch op {
	case "followers", "following":
		if opts == nil || opts.ID == "" {
			return []Account{}, ErrInvalidID
		}
		endPoint = "accounts/" + opts.ID + "/" + op
	case "follow_requests", "blocks", "mutes":
		endPoint = op
	case "search":
		if opts == nil || opts.Q == "" {
			return []Account{}, ErrInvalidParameter
		}
		endPoint = "accounts/" + op
	case "reblogged_by", "favourited_by":
		if opts == nil || opts.ID == "" {
			return []Account{}, ErrInvalidID
		}
		endPoint = "statuses/" + opts.ID + "/" + op
	default:
		return nil, ErrInvalidParameter
	}

	// Handle target-specific query parameters
	params := make(apiCallParams)
	if op == "search" {
		params["q"] = opts.Q
		if opts.Following {
			params["following"] = "true"
		}
	}

	return mc.getMultipleAccounts(endPoint, params, lopt)
}

// GetAccount returns an account entity
// The returned value can be nil if there is an error or if the
// requested ID does not exist.
func (mc *Client) GetAccount(accountID ActivityID) (*Account, error) {
	account, err := mc.getSingleAccount("account", accountID)
	if err != nil {
		return nil, err
	}
	if account != nil && account.ID == "" {
		return nil, ErrEntityNotFound
	}
	return account, nil
}

// GetCurrentAccount returns the current user account
func (mc *Client) GetCurrentAccount() (*Account, error) {
	account, err := mc.getSingleAccount("verify_credentials", "")
	if err != nil {
		return nil, err
	}
	if account != nil && account.ID == "" {
		return nil, ErrEntityNotFound
	}
	return account, nil
}

// GetAccountFollowers returns the list of accounts following a given account
func (mc *Client) GetAccountFollowers(accountID ActivityID, lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{ID: accountID, Limit: lopt}
	return mc.getMultipleAccountsHelper("followers", o)
}

// GetAccountFollowing returns the list of accounts a given account is following
func (mc *Client) GetAccountFollowing(accountID ActivityID, lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{ID: accountID, Limit: lopt}
	return mc.getMultipleAccountsHelper("following", o)
}

// FollowAccount follows an account
// 'reblogs' can be used to specify if boots should be displayed or hidden.
func (mc *Client) FollowAccount(accountID ActivityID, reblogs *bool) (*Relationship, error) {
	var params apiCallParams
	if reblogs != nil {
		params = make(apiCallParams)
		if *reblogs {
			params["reblogs"] = "true"
		} else {
			params["reblogs"] = "false"
		}
	}
	rel, err := mc.updateRelationship("follow", accountID, params)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// UnfollowAccount unfollows an account
func (mc *Client) UnfollowAccount(accountID ActivityID) (*Relationship, error) {
	rel, err := mc.updateRelationship("unfollow", accountID, nil)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// FollowRemoteAccount follows a remote account
// The parameter 'uri' is a URI (e.g. "username@domain").
func (mc *Client) FollowRemoteAccount(uri string) (*Account, error) {
	if uri == "" {
		return nil, ErrInvalidID
	}

	params := make(apiCallParams)
	params["uri"] = uri

	var account Account
	if err := mc.apiCall("v1/follows", rest.Post, params, nil, nil, &account); err != nil {
		return nil, err
	}
	if account.ID == "" {
		return nil, ErrEntityNotFound
	}
	return &account, nil
}

// BlockAccount blocks an account
func (mc *Client) BlockAccount(accountID ActivityID) (*Relationship, error) {
	rel, err := mc.updateRelationship("block", accountID, nil)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// UnblockAccount unblocks an account
func (mc *Client) UnblockAccount(accountID ActivityID) (*Relationship, error) {
	rel, err := mc.updateRelationship("unblock", accountID, nil)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// MuteAccount mutes an account
// Note that with current Mastodon API, muteNotifications defaults to true
// when it is not provided.
func (mc *Client) MuteAccount(accountID ActivityID, muteNotifications *bool) (*Relationship, error) {
	var params apiCallParams

	if muteNotifications != nil {
		params = make(apiCallParams)
		if *muteNotifications {
			params["notifications"] = "true"
		} else {
			params["notifications"] = "false"
		}
	}

	rel, err := mc.updateRelationship("mute", accountID, params)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// UnmuteAccount unmutes an account
func (mc *Client) UnmuteAccount(accountID ActivityID) (*Relationship, error) {
	rel, err := mc.updateRelationship("unmute", accountID, nil)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrEntityNotFound
	}
	return rel, nil
}

// SearchAccounts returns a list of accounts matching the query string
// The lopt parameter is optional (can be nil) or can be used to set a limit.
func (mc *Client) SearchAccounts(query string, following bool, lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{Q: query, Limit: lopt, Following: following}
	return mc.getMultipleAccountsHelper("search", o)
}

// GetBlockedAccounts returns the list of blocked accounts
// The lopt parameter is optional (can be nil).
func (mc *Client) GetBlockedAccounts(lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{Limit: lopt}
	return mc.getMultipleAccountsHelper("blocks", o)
}

// GetMutedAccounts returns the list of muted accounts
// The lopt parameter is optional (can be nil).
func (mc *Client) GetMutedAccounts(lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{Limit: lopt}
	return mc.getMultipleAccountsHelper("mutes", o)
}

// GetAccountFollowRequests returns the list of follow requests accounts
// The lopt parameter is optional (can be nil).
func (mc *Client) GetAccountFollowRequests(lopt *LimitParams) ([]Account, error) {
	o := &getAccountsOptions{Limit: lopt}
	return mc.getMultipleAccountsHelper("follow_requests", o)
}

// GetAccountRelationships returns a list of relationship entities for the given accounts
func (mc *Client) GetAccountRelationships(accountIDs []ActivityID) ([]Relationship, error) {
	if len(accountIDs) < 1 {
		return nil, ErrInvalidID
	}

	params := make(apiCallParams)
	for i, id := range accountIDs {
		if id == "" {
			return nil, ErrInvalidID
		}
		qID := fmt.Sprintf("[%d]id", i)
		params[qID] = id
	}

	var rl []Relationship
	if err := mc.apiCall("v1/accounts/relationships", rest.Get, params, nil, nil, &rl); err != nil {
		return nil, err
	}
	return rl, nil
}

// GetAccountStatuses returns a list of status entities for the given account
// If onlyMedia is true, returns only statuses that have media attachments.
// If onlyPinned is true, returns only statuses that have been pinned.
// If excludeReplies is true, skip statuses that reply to other statuses.
// If lopt.All is true, several requests will be made until the API server
// has nothing to return.
// If lopt.Limit is set (and not All), several queries can be made until the
// limit is reached.
func (mc *Client) GetAccountStatuses(accountID ActivityID, onlyPinned, onlyMedia, excludeReplies bool, lopt *LimitParams) ([]Status, error) {
	if accountID == "" {
		return nil, ErrInvalidID
	}

	endPoint := "accounts/" + accountID + "/" + "statuses"
	params := make(apiCallParams)
	if onlyMedia {
		params["only_media"] = "true"
	}
	if onlyPinned {
		params["pinned"] = "true"
	}
	if excludeReplies {
		params["exclude_replies"] = "true"
	}

	return mc.getMultipleStatuses(endPoint, params, lopt)
}

// FollowRequestAuthorize authorizes or rejects an account follow-request
func (mc *Client) FollowRequestAuthorize(accountID ActivityID, authorize bool) error {
	endPoint := "follow_requests/reject"
	if authorize {
		endPoint = "follow_requests/authorize"
	}
	_, err := mc.getSingleAccount(endPoint, accountID)
	return err
}

// UpdateAccount updates the connected user's account data
//
// The fields avatar & headerImage are considered as file paths
// and their content will be uploaded.
// Please note that currently Mastodon leaks the avatar file name:
// https://github.com/tootsuite/mastodon/issues/5776
//
// All fields can be nil, in which case they are not updated.
// 'DisplayName' and 'Note' can be set to "" to delete previous values.
// Setting 'Locked' to true means all followers should be approved.
// You can set 'Bot' to true to indicate this is a service (automated) account.
// I'm not sure images can be deleted -- only replaced AFAICS.
func (mc *Client) UpdateAccount(cmdParams UpdateAccountParams) (*Account, error) {
	const endPoint = "accounts/update_credentials"
	params := make(apiCallParams)

	if cmdParams.DisplayName != nil {
		params["display_name"] = *cmdParams.DisplayName
	}
	if cmdParams.Note != nil {
		params["note"] = *cmdParams.Note
	}
	if cmdParams.Locked != nil {
		if *cmdParams.Locked {
			params["locked"] = "true"
		} else {
			params["locked"] = "false"
		}
	}
	if cmdParams.Bot != nil {
		if *cmdParams.Bot {
			params["bot"] = "true"
		} else {
			params["bot"] = "false"
		}
	}
	if cmdParams.FieldsAttributes != nil {
		if len(*cmdParams.FieldsAttributes) > 4 {
			return nil, errors.New("too many fields (max=4)")
		}
		for i, attr := range *cmdParams.FieldsAttributes {
			qName := fmt.Sprintf("fields_attributes[%d][name]", i)
			qValue := fmt.Sprintf("fields_attributes[%d][value]", i)
			params[qName] = attr.Name
			params[qValue] = attr.Value
		}
	}
	if cmdParams.Source != nil {
		s := cmdParams.Source

		if s.Privacy != nil {
			params["source[privacy]"] = *s.Privacy
		}
		if s.Language != nil {
			params["source[language]"] = *s.Language
		}
		if s.Sensitive != nil {
			params["source[sensitive]"] = fmt.Sprintf("%v", *s.Sensitive)
		}
	}

	var err error
	var avatar, headerImage []byte

	avatar, err = readFile(cmdParams.AvatarImagePath)
	if err != nil {
		return nil, err
	}

	headerImage, err = readFile(cmdParams.HeaderImagePath)
	if err != nil {
		return nil, err
	}

	var formBuf bytes.Buffer
	w := multipart.NewWriter(&formBuf)

	if avatar != nil {
		formWriter, err := w.CreateFormFile("avatar", filepath.Base(*cmdParams.AvatarImagePath))
		if err != nil {
			return nil, errors.Wrap(err, "avatar upload")
		}
		formWriter.Write(avatar)
	}
	if headerImage != nil {
		formWriter, err := w.CreateFormFile("header", filepath.Base(*cmdParams.HeaderImagePath))
		if err != nil {
			return nil, errors.Wrap(err, "header upload")
		}
		formWriter.Write(headerImage)
	}

	for k, v := range params {
		fw, err := w.CreateFormField(k)
		if err != nil {
			return nil, errors.Wrapf(err, "form field: %s", k)
		}
		n, err := io.WriteString(fw, v)
		if err != nil {
			return nil, errors.Wrapf(err, "writing field: %s", k)
		}
		if n != len(v) {
			return nil, errors.Wrapf(err, "partial field: %s", k)
		}
	}

	w.Close()

	// Prepare the request
	req, err := mc.prepareRequest("v1/"+endPoint, rest.Patch, params)
	if err != nil {
		return nil, errors.Wrap(err, "prepareRequest failed")
	}
	req.Headers["Content-Type"] = w.FormDataContentType()
	req.Body = formBuf.Bytes()

	// Make API call
	r, err := restAPI(req)
	if err != nil {
		return nil, errors.Wrap(err, "account update failed")
	}

	// Check for error reply
	var errorResult Error
	if err := json.Unmarshal([]byte(r.Body), &errorResult); err == nil {
		// The empty object is not an error
		if errorResult.Text != "" {
			return nil, errors.New(errorResult.Text)
		}
	}

	// Not an error reply; let's unmarshal the data
	var account Account
	if err := json.Unmarshal([]byte(r.Body), &account); err != nil {
		return nil, errors.Wrap(err, "cannot decode API response")
	}
	return &account, nil
}

// readFile is a helper function to read a file's contents.
func readFile(filename *string) ([]byte, error) {
	if filename == nil || *filename == "" {
		return nil, nil
	}

	file, err := os.Open(*filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fStat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fStat.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
