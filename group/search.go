/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package group

import (
	"github.com/ExpediaGroup/flyte-ldap/ldap"
	ldapClient "gopkg.in/ldap.v2"
	"strings"
)

type SearchDetails struct {
	Attributes     []string // i.e. the attributes to be returned by the group, e.g. 'memberOf'
	BaseDn         string
	SearchFilter   string
	GroupAttribute string // the attribute that gives the name of the group from the attribute values, e.g. 'cn'
	SearchTimeout  int
}

type Searcher interface {
	GetGroupsFor(sd *SearchDetails, username string) ([]string, error)
}

type searcher struct {
	client ldap.Client
}

func NewSearcher(client ldap.Client) Searcher {
	return &searcher{client: client}
}

func (searcher *searcher) GetGroupsFor(sd *SearchDetails, username string) ([]string, error) {
	if err := searcher.client.ConnectTls(); err != nil {
		return nil, err
	}
	defer searcher.client.Close()

	searchRequest := ldap.SearchRequest{
		Attributes:    sd.Attributes,
		BaseDn:        sd.BaseDn,
		SearchFilter:  strings.Replace(sd.SearchFilter, "{username}", username, -1),
		SearchTimeout: sd.SearchTimeout,
	}

	searchResults, err := searcher.client.Search(searchRequest)

	if err != nil {
		return nil, err
	}
	return extractUserGroupsFrom(searchResults, sd.GroupAttribute), nil
}

func extractUserGroupsFrom(searchResults *ldapClient.SearchResult, groupAttribute string) []string {
	groups := []string{}
	if len(searchResults.Entries) > 0 {
		for _, attr := range searchResults.Entries[0].Attributes {
			for _, attributeValue := range attr.Values {
				if userGroup := extractUserGroupFrom(attributeValue, groupAttribute); userGroup != "" {
					groups = append(groups, userGroup)
				}
			}
		}
	}
	return groups
}

func extractUserGroupFrom(attributeValue, groupAttribute string) string {
	for _, v := range strings.Split(attributeValue, ",") {
		if strings.HasPrefix(strings.ToUpper(v), strings.ToUpper(groupAttribute)+"=") {
			return v[3:]
		}
	}
	return ""
}
