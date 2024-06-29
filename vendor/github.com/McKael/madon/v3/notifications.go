/*
Copyright 2017-2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"fmt"

	"github.com/sendgrid/rest"
)

// GetNotifications returns the list of the user's notifications
// excludeTypes is an array of notifications to exclude ("follow", "favourite",
// "reblog", "mention").  It can be nil.
// If lopt.All is true, several requests will be made until the API server
// has nothing to return.
// If lopt.Limit is set (and not All), several queries can be made until the
// limit is reached.
func (mc *Client) GetNotifications(excludeTypes []string, lopt *LimitParams) ([]Notification, error) {
	var notifications []Notification
	var links apiLinks
	var params apiCallParams

	if len(excludeTypes) > 0 {
		params = make(apiCallParams)
		for i, eType := range excludeTypes {
			qID := fmt.Sprintf("[%d]exclude_types", i)
			params[qID] = eType
		}
	}

	if err := mc.apiCall("v1/notifications", rest.Get, params, lopt, &links, &notifications); err != nil {
		return nil, err
	}
	if lopt != nil { // Fetch more pages to reach our limit
		for (lopt.All || lopt.Limit > len(notifications)) && links.next != nil {
			notifSlice := []Notification{}
			newlopt := links.next
			links = apiLinks{}
			if err := mc.apiCall("v1/notifications", rest.Get, nil, newlopt, &links, &notifSlice); err != nil {
				return nil, err
			}
			notifications = append(notifications, notifSlice...)
		}
	}
	return notifications, nil
}

// GetNotification returns a notification
// The returned notification can be nil if there is an error or if the
// requested notification does not exist.
func (mc *Client) GetNotification(notificationID ActivityID) (*Notification, error) {
	if notificationID == "" {
		return nil, ErrInvalidID
	}

	var endPoint = "notifications/" + notificationID
	var notification Notification
	if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, nil, nil, &notification); err != nil {
		return nil, err
	}
	if notification.ID == "" {
		return nil, ErrEntityNotFound
	}
	return &notification, nil
}

// DismissNotification deletes a notification
func (mc *Client) DismissNotification(notificationID ActivityID) error {
	if notificationID == "" {
		return ErrInvalidID
	}

	endPoint := "notifications/dismiss"
	params := apiCallParams{"id": notificationID}
	err := mc.apiCall("v1/"+endPoint, rest.Post, params, nil, nil, &Notification{})
	return err
}

// ClearNotifications deletes all notifications from the Mastodon server for
// the authenticated user
func (mc *Client) ClearNotifications() error {
	err := mc.apiCall("v1/notifications/clear", rest.Post, nil, nil, nil, &Notification{})
	return err
}
