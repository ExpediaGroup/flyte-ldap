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
	"crypto/tls"
	"fmt"
	"gopkg.in/ldap.v2"
)

type Client interface {
	Connect() error
	ConnectTls(insecureSkipVerify bool) error
	Search(sr SearchRequest) (*ldap.SearchResult, error)
	Close()
}

type ldapClient struct {
	bindUsername  string
	bindPassword  string
	ldapServerUrl string
	ldapSearcher  ldapSearcher
}

type SearchRequest struct {
	Attributes    []string // i.e. the attributes to be returned by the group, e.g. 'memberOf'
	BaseDn        string
	SearchFilter  string
	SearchTimeout int
}

type ldapSearcher interface {
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close()
}

func NewClient(bindUsername, bindPassword, ldapServerUrl string) Client {
	return &ldapClient{
		bindUsername:  bindUsername,
		bindPassword:  bindPassword,
		ldapServerUrl: ldapServerUrl,
	}
}

func (c *ldapClient) Connect() error {
	ldapConn, err := ldap.Dial("tcp", c.ldapServerUrl)
	if err != nil {
		return fmt.Errorf("Cannot connect to LDAP: %v", err)
	}

	err = ldapConn.Bind(c.bindUsername, c.bindPassword)
	if err != nil {
		ldapConn.Close()
		return fmt.Errorf("Cannot bind to LDAP: %v", err)
	}

	c.ldapSearcher = ldapConn

	return nil
}

func (c *ldapClient) ConnectTls(insecureSkipVerify bool) error {
	ldapConn, err := ldap.DialTLS("tcp", c.ldapServerUrl, &tls.Config{InsecureSkipVerify: insecureSkipVerify})
	if err != nil {
		return fmt.Errorf("Cannot connect to LDAP: %v", err)
	}

	err = ldapConn.Bind(c.bindUsername, c.bindPassword)
	if err != nil {
		ldapConn.Close()
		return fmt.Errorf("Cannot bind to LDAP: %v", err)
	}

	c.ldapSearcher = ldapConn

	return nil
}

func (c *ldapClient) Search(sr SearchRequest) (*ldap.SearchResult, error) {
	searchRequest := &ldap.SearchRequest{
		BaseDN:       sr.BaseDn,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    sr.SearchTimeout,
		TypesOnly:    false,
		Filter:       sr.SearchFilter,
		Attributes:   sr.Attributes,
		Controls:     nil,
	}

	searchResults, err := c.ldapSearcher.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP group error: %v", err.Error())
	}
	return searchResults, nil
}

func (c *ldapClient) Close() {
	c.ldapSearcher.Close()
}
