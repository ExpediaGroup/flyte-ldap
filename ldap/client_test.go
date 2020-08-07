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

package ldap

import (
	"errors"
	ldapserver "github.com/nmcclain/ldap"
	"gopkg.in/ldap.v2"
	"log"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	ldapServerUrl         = "localhost:49456"
	bindDistinguishedName = "cn=testy,dc=testers,dc=testz"
	bindPassword          = "work123"
	shouldBind            = ldapserver.LDAPResultSuccess
	shouldNotBind         = ldapserver.LDAPResultInvalidCredentials
)

// test the connection

func TestConnectShouldConnectToLDAPServer(t *testing.T) {
	quit := make(chan bool)
	startLdapServer(quit, shouldBind)
	defer stopLdapServer(quit)
	client := NewClient(bindDistinguishedName, bindPassword, ldapServerUrl)

	err := client.Connect()
	defer client.Close()

	if err != nil {
		t.Fatalf("Unexpected error: '%s'.", err.Error())
	}
}

func TestConnectShouldReturnErrorIfClientCannotConnectToLdap(t *testing.T) {
	// note ldap test server not started

	client := NewClient(bindDistinguishedName, bindPassword, ldapServerUrl)

	err := client.Connect()

	if err == nil {
		t.Fatal("Should've returned error due to ldap connection failure.")
	}
	if !strings.Contains(err.Error(), "Cannot connect to LDAP:") {
		t.Errorf("Error returned is wrong. Error: %s", err.Error())
	}
}

func TestConnectShouldReturnErrorIfClientCannotBindToLdap(t *testing.T) {
	quit := make(chan bool)
	startLdapServer(quit, shouldNotBind)
	defer stopLdapServer(quit)
	client := NewClient(bindDistinguishedName, bindPassword, ldapServerUrl)

	err := client.Connect()

	if err == nil {
		t.Fatal("Should've returned error due to ldap bind failure.")
	}
	if !strings.Contains(err.Error(), "Cannot bind to LDAP:") {
		t.Errorf("Error returned is wrong. Error: %s", err.Error())
	}
}

func startLdapServer(quit chan bool, br ldapserver.LDAPResultCode) {
	go func() {
		s := ldapserver.NewServer()
		s.QuitChannel(quit)

		s.BindFunc(bindDistinguishedName, bindResultOf(br))

		// start the server
		log.Printf("Starting example LDAP server on %s", ldapServerUrl)
		if err := s.ListenAndServe(ldapServerUrl); err != nil {
			log.Fatalf("LDAP Server Failed: %s", err.Error())
		}
	}()

	waitForPort(5)
}

func waitForPort(retries int) {
	if retries < 0 {
		return
	}
	conn, err := net.DialTimeout("tcp", ldapServerUrl, 2*time.Second)
	if err != nil {
		time.Sleep(1 * time.Second)
		waitForPort(retries - 1)
	}
	if conn != nil {
		defer conn.Close()
	}
}

func stopLdapServer(quit chan bool) {
	quit <- true
}

type BinderFunc func(bindDN, bindSimplePw string, conn net.Conn) (ldapserver.LDAPResultCode, error)

func (b BinderFunc) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldapserver.LDAPResultCode, error) {
	return b(bindDN, bindSimplePw, conn)
}

func bindResultOf(br ldapserver.LDAPResultCode) BinderFunc {
	return func(bindDN, bindSimplePw string, conn net.Conn) (ldapserver.LDAPResultCode, error) {
		return br, nil
	}
}

// test the group

func TestSearchShouldCallLdapSearchWithCorrectParametersAndReturnSearchResults(t *testing.T) {
	searchResultsToReturn := &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{"cn=dave-jones,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,", []*ldap.EntryAttribute{
				{"memberOf", []string{
					"CN=London team,OU=Distribution Lists,DC=com",
					"OU=Distribution Lists,cn=New York team,DC=com,DC=EDFR,DC=DFER,DC=com"}, nil},
				{"uid", []string{"fsdf56sdf54fs645f"}, nil},
				{"description", []string{"Something about Dave"}, nil},
			}},
		},
	}
	ldapSearcher := &mockSearcher{returnedSearchResults: searchResultsToReturn}
	client := ldapClient{ldapSearcher: ldapSearcher}
	searchRequest := SearchRequest{
		Attributes:    []string{"memberOf"},
		BaseDn:        "OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE",
		SearchFilter:  "(mailNickname=dave-jones)",
		SearchTimeout: 20,
	}

	results, err := client.Search(searchRequest)

	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Fatal("LDAP group not called!")
	}
	if ldapSearcher.searchRequest.BaseDN != searchRequest.BaseDn {
		t.Errorf("BaseDn passed to the LDAP group is incorrect. Expected: '%s', actual: '%s'.", searchRequest.BaseDn, ldapSearcher.searchRequest.BaseDN)
	}
	if ldapSearcher.searchRequest.TimeLimit != searchRequest.SearchTimeout {
		t.Errorf("Search timeout passed to the LDAP group is incorrect. Expected: '%d', actual: '%d'.", searchRequest.SearchTimeout, ldapSearcher.searchRequest.TimeLimit)
	}
	if ldapSearcher.searchRequest.Filter != searchRequest.SearchFilter {
		t.Errorf("Search filter passed to the LDAP group is incorrect. Expected: '%s', actual: '%s'.", searchRequest.SearchFilter, ldapSearcher.searchRequest.Filter)
	}
	if !reflect.DeepEqual(ldapSearcher.searchRequest.Attributes, searchRequest.Attributes) {
		t.Errorf("Attributes passed to the LDAP group is incorrect. Expected: '%s', actual: '%s'.", searchRequest.Attributes, ldapSearcher.searchRequest.Attributes)
	}
}

func TestShouldReturnErrorIfSearchProblem(t *testing.T) {
	client := ldapClient{ldapSearcher: &mockSearcher{shouldReturnError: true}}
	searchRequest := SearchRequest{
		Attributes:    []string{"memberOf"},
		BaseDn:        "OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE",
		SearchFilter:  "(mailNickname=dave-jones)",
		SearchTimeout: 20,
	}

	_, err := client.Search(searchRequest)

	if err == nil {
		t.Fatal("Should've returned error due to ldap group failure.")
	}
	if !strings.Contains(err.Error(), "LDAP group error:") {
		t.Errorf("Error returned is wrong. Error: %s", err.Error())
	}
}

func TestCloseShouldCallCloseMethodOnTheLdapSearcher(t *testing.T) {
	mockSearcher := &mockSearcher{}
	client := ldapClient{ldapSearcher: mockSearcher}

	client.Close()

	if !mockSearcher.isClosed {
		t.Error("Close method not called!")
	}
}

type mockSearcher struct {
	isClosed              bool
	searchRequest         *ldap.SearchRequest
	returnedSearchResults *ldap.SearchResult
	shouldReturnError     bool
}

func (s *mockSearcher) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	s.searchRequest = searchRequest
	if s.shouldReturnError {
		return nil, errors.New("Some error")
	}
	return s.returnedSearchResults, nil
}

func (s *mockSearcher) Close() {
	s.isClosed = true
}
