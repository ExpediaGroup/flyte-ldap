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
	"errors"
	"github.com/ExpediaGroup/flyte-ldap/ldap"
	ldapClient "gopkg.in/ldap.v2"
	"reflect"
	"testing"
)

func TestClientConnectAndCloseAreCalledWhenCallingGetGroupsFor(t *testing.T) {
	isClientConnectCalled := false
	isClientCloseCalled := false
	mockClient := &mockClient{
		connect: func() error {
			isClientConnectCalled = true
			return nil
		},
		close: func() {
			isClientCloseCalled = true
		},
		search: func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
			return &ldapClient.SearchResult{}, nil
		}}
	searcher := NewSearcher(mockClient)
	searchDetails := &SearchDetails{}

	_, err := searcher.GetGroupsFor(searchDetails, "dave-jones")

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if !isClientConnectCalled {
		t.Error("Client connect has not been called.")
	}
	if !isClientCloseCalled {
		t.Error("Client close has not been called.")
	}
}

func TestErrorIsReturnedIfClientConnectError(t *testing.T) {
	mockClient := &mockClient{
		connect: func() error {
			return errors.New("Meh")
		},
	}
	searcher := NewSearcher(mockClient)
	searchDetails := &SearchDetails{}

	_, err := searcher.GetGroupsFor(searchDetails, "dave-jones")

	if err == nil {
		t.Fatal("Expected error!")
	}
	if err.Error() != "Meh" {
		t.Errorf("Error message not correct. Expected: 'Meh', actual: '%s'", err.Error())
	}
}

func TestSearchShouldReturnsUserGroups(t *testing.T) {
	returnedSearchResults := &ldapClient.SearchResult{
		Entries: []*ldapClient.Entry{
			{
				DN: "cn=dave-jones,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,",
				Attributes: []*ldapClient.EntryAttribute{
					{"memberOf", []string{
						"CN=London team,OU=Distribution Lists,DC=com",
						"OU=Distribution Lists,cn=New York team,DC=com,DC=EDFR,DC=DFER,DC=com",
						"OU=Dave DLs,DC=SWD,DC=DFER,DC=com,cN=Paris team",
						"Cn=Brussels team,OU=John DLs,DC=SWE,DC=DFER,DC=com"}, nil},
					{"uid", []string{"fsdf56sdf54fs645f"}, nil},
					{"description", []string{"Something about Dave"}, nil},
				}},
		},
	}
	mockClient := &mockClient{
		connect: func() error { return nil },
		close:   func() {},
		search: func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
			return returnedSearchResults, nil
		}}
	searcher := NewSearcher(mockClient)
	searchDetails := someSearchDetails()

	userGroups, err := searcher.GetGroupsFor(searchDetails, "dave-jones")

	if err != nil {
		t.Fatalf("Unexpected search error: %s", err.Error())
	}
	if userGroups[0] != "London team" {
		t.Errorf("User group should be 'London team', is: %v", userGroups[0])
	}
	if userGroups[1] != "New York team" {
		t.Errorf("User group should be 'New York team', is: %v", userGroups[1])
	}
	if userGroups[2] != "Paris team" {
		t.Errorf("User group should be 'Paris team', is: %v", userGroups[2])
	}
	if userGroups[3] != "Brussels team" {
		t.Errorf("User group should be 'Brussels team', is: %v", userGroups[3])
	}
}

func TestSearchShouldNotReturnsUserGroupsIfNoSearchResultsAreReturned(t *testing.T) {
	mockClient := &mockClient{
		connect: func() error { return nil },
		close:   func() {},
		search: func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
			return &ldapClient.SearchResult{}, nil
		}}
	searcher := NewSearcher(mockClient)
	searchDetails := someSearchDetails()

	userGroups, err := searcher.GetGroupsFor(searchDetails, "dave-jones")

	if err != nil {
		t.Fatalf("Unexpected search error: %s", err.Error())
	}
	if len(userGroups) > 0 {
		t.Errorf("No user groups should have been returned. User groups returned: %v", userGroups)
	}
}

func TestSearchShouldReturnClientSearchError(t *testing.T) {
	mockClient := &mockClient{
		connect: func() error { return nil },
		close:   func() {},
		search: func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
			return nil, errors.New("Some error")
		}}
	searcher := NewSearcher(mockClient)
	searchDetails := someSearchDetails()

	_, err := searcher.GetGroupsFor(searchDetails, "dave-jones")

	if err == nil {
		t.Error("Search should've returned an error.")
	}
	if err.Error() != "Some error" {
		t.Errorf("Error message should be: 'Some error', is: '%s''", err.Error())
	}
}

func TestCorrectParametersArePassedToClientSearch(t *testing.T) {
	var searchRequest ldap.SearchRequest
	mockClient := &mockClient{
		connect: func() error { return nil },
		close:   func() {},
		search: func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
			searchRequest = sr
			return &ldapClient.SearchResult{}, nil
		}}
	searcher := NewSearcher(mockClient)
	searchDetails := someSearchDetails()

	searcher.GetGroupsFor(searchDetails, "dave-jones")

	if !reflect.DeepEqual(searchRequest.Attributes, []string{"memberOf"}) {
		t.Errorf("Search request attributes is wrong. Should be 'memberOf', is: %s", searchRequest.Attributes)
	}
	if searchRequest.BaseDn != "cn=blah-blah,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE," {
		t.Errorf("Base dn is wrong. Should be 'cn=blah-blah,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,', is: %s", searchRequest.BaseDn)
	}
	if searchRequest.SearchFilter != "(mailNickname=dave-jones)" {
		t.Errorf("Group filter is wrong. Should be '(mailNickname=dave-jones)', is: %s", searchRequest.SearchFilter)
	}
	if searchRequest.SearchTimeout != 20 {
		t.Errorf("Search timeout is wrong. Should be '20', is: %d", searchRequest.SearchTimeout)
	}
}

func TestExtractUserGroupsShouldExtractUserGroupsFromSearchResults(t *testing.T) {
	searchResults := &ldapClient.SearchResult{
		Entries: []*ldapClient.Entry{
			{
				DN: "cn=dave-jones,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,",
				Attributes: []*ldapClient.EntryAttribute{
					{"memberOf", []string{
						"CN=London team,OU=Distribution Lists,DC=com",
						"OU=Distribution Lists,cn=New York team,DC=com,DC=EDFR,DC=DFER,DC=com",
						"OU=Dave DLs,DC=SWD,DC=DFER,DC=com,cN=Paris team",
						"Cn=Brussels team,OU=John DLs,DC=SWE,DC=DFER,DC=com"}, nil},
					{"uid", []string{"fsdf56sdf54fs645f"}, nil},
					{"description", []string{"Something about Dave"}, nil},
				}},
		},
	}

	userGroups := extractUserGroupsFrom(searchResults, "cn")

	if userGroups[0] != "London team" {
		t.Errorf("User group should be 'London team', is: %v", userGroups[0])
	}
	if userGroups[1] != "New York team" {
		t.Errorf("User group should be 'New York team', is: %v", userGroups[1])
	}
	if userGroups[2] != "Paris team" {
		t.Errorf("User group should be 'Paris team', is: %v", userGroups[2])
	}
	if userGroups[3] != "Brussels team" {
		t.Errorf("User group should be 'Brussels team', is: %v", userGroups[3])
	}
}

func TestExtractUserGroupsShouldReturnEmptyStringArrayFromEmptySearchResults(t *testing.T) {
	searchResults := &ldapClient.SearchResult{}

	userGroups := extractUserGroupsFrom(searchResults, "cn")

	if len(userGroups) > 0 {
		t.Errorf("No user groups should have been returned. User groups returned: %v", userGroups)
	}
}

func someSearchDetails() *SearchDetails {
	return &SearchDetails{
		Attributes:     []string{"memberOf"},
		BaseDn:         "cn=blah-blah,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,",
		SearchFilter:   "(mailNickname={username})",
		GroupAttribute: "cn",
		SearchTimeout:  20,
	}
}

type mockClient struct {
	connect func() error
	close   func()
	search  func(sr ldap.SearchRequest) (*ldapClient.SearchResult, error)
}

func (c *mockClient) Connect() error {
	return c.connect()
}

func (c *mockClient) Search(sr ldap.SearchRequest) (*ldapClient.SearchResult, error) {
	return c.search(sr)
}

func (c *mockClient) Close() {
	c.close()
}
