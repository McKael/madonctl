/*
Copyright 2017-2018 Mikael Berthe

Licensed under the MIT license.  Please see the LICENSE file is this directory.
*/

package madon

import (
	"github.com/sendgrid/rest"
)

// GetBlockedDomains returns the current user blocked domains
// If lopt.All is true, several requests will be made until the API server
// has nothing to return.
func (mc *Client) GetBlockedDomains(lopt *LimitParams) ([]DomainName, error) {
	const endPoint = "domain_blocks"
	var links apiLinks
	var domains []DomainName
	if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, lopt, &links, &domains); err != nil {
		return nil, err
	}
	if lopt != nil { // Fetch more pages to reach our limit
		for (lopt.All || lopt.Limit > len(domains)) && links.next != nil {
			domainSlice := []DomainName{}
			newlopt := links.next
			links = apiLinks{}
			if err := mc.apiCall("v1/"+endPoint, rest.Get, nil, newlopt, &links, &domainSlice); err != nil {
				return nil, err
			}
			domains = append(domains, domainSlice...)
		}
	}
	return domains, nil
}

// BlockDomain blocks the specified domain
func (mc *Client) BlockDomain(domain DomainName) error {
	const endPoint = "domain_blocks"
	params := make(apiCallParams)
	params["domain"] = string(domain)
	return mc.apiCall("v1/"+endPoint, rest.Post, params, nil, nil, nil)
}

// UnblockDomain unblocks the specified domain
func (mc *Client) UnblockDomain(domain DomainName) error {
	const endPoint = "domain_blocks"
	params := make(apiCallParams)
	params["domain"] = string(domain)
	return mc.apiCall("v1/"+endPoint, rest.Delete, params, nil, nil, nil)
}
