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

package main

import (
	"testing"
	"net/url"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigVal_shouldReturnEnvironmentConfigValue(t *testing.T) {
	env = &mockEnvironment{
		values: map[string]string{
			"LDAP_URL": "my.ldap.com:123",
		},
	}

	ldapUrl := configVal("LDAP_URL")

	if ldapUrl != "my.ldap.com:123" {
		t.Errorf("Value returned wrong. Expect: '%s'. Actual '%s'", "my.ldap.com:123", ldapUrl)
	}
}

func TestConfigVal_shouldLogFatalIfConfigValueNotSet(t *testing.T) {
	loggertest.Init(loggertest.LogLevelInfo)
	defer loggertest.Reset()

	defer func() {
		if r := recover(); r != nil {
			logMessages := loggertest.GetLogMessages()
			require.Len(t, logMessages, 1)
			assert.Contains(t, logMessages[0].RawMessage, "Config value \"LDAP_URL\" must be set")
		}
	}()

	configVal("LDAP_URL")
}

func TestOptionalConfigVal_shouldReturnEnvironmentConfigValue(t *testing.T) {
	env = &mockEnvironment{
		values: map[string]string{
			"LDAP_URL": "my.ldap.com:123",
		},
	}

	ldapUrl := optionalConfigVal("LDAP_URL", "ldap.default.com:1234")

	if ldapUrl != "my.ldap.com:123" {
		t.Errorf("Value returned wrong. Expect: '%s'. Actual '%s'", "my.ldap.com:123", ldapUrl)
	}
}

func TestOptionalConfigVal_shouldReturnDefaultConfigValueWhenEnvironmentConfigNotSet(t *testing.T) {
	setEmptyEnvironment()

	ldapUrl := optionalConfigVal("LDAP_URL", "ldap.default.com:1234")

	if ldapUrl != "ldap.default.com:1234" {
		t.Errorf("Value returned wrong. Expect: '%s'. Actual '%s'", "ldap.default.com:1234", ldapUrl)
	}
}

func TestCreateUrl_shouldCreateUrlFromStringRepresentation(t *testing.T) {
	strUrl := "http://www.something.com"
	var url *url.URL

	url = createURL(strUrl)

	if url.Scheme != "http" {
		t.Errorf("Scheme is wrong. Should be: 'http', is: '%s'.", url.Scheme)
	}
	if url.Host != "www.something.com" {
		t.Errorf("Host is wrong. Should be: 'www.something.com', is: '%s'.", url.Host)
	}
}

func TestCreateUrl_shouldLogFatalIfStringUrlCannotBeParsed(t *testing.T) {
	loggertest.Init(loggertest.LogLevelInfo)
	defer loggertest.Reset()

	defer func() {
		if r := recover(); r != nil {
			logMessages := loggertest.GetLogMessages()
			require.Len(t, logMessages, 1)
			assert.Contains(t, logMessages[0].RawMessage, "Cannot parse url: '://hello' error: 'parse ://hello: missing protocol scheme'")
		}
	}()

	createURL("://hello")
}

func setEmptyEnvironment() {
	env = &mockEnvironment{}
}

type mockEnvironment struct {
	values map[string]string
}

func (m *mockEnvironment) getValueFor(name string) string {
	return m.values[name]
}
