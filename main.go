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
	"github.com/ExpediaGroup/flyte-ldap/command"
	"github.com/ExpediaGroup/flyte-ldap/group"
	"github.com/ExpediaGroup/flyte-ldap/ldap"
	"github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var env environment = &osEnvironment{}

type environment interface {
	getValueFor(name string) string
}

type osEnvironment struct{}

func (o *osEnvironment) getValueFor(name string) string {
	return os.Getenv(name)
}

func main() {
	searchTimeout, err := strconv.Atoi(optionalConfigVal("SEARCH_TIMEOUT_IN_SECONDS", "20"))
	if err != nil {
		logger.Fatalf("LDAP group timeout '%v' not convertible to an integer. Error: %v", configVal("SEARCH_TIMEOUT_IN_SECONDS"), err)
	}
	tlsEnabledFlag, err := strconv.ParseBool(configVal("ENABLE_TLS"))
	if err != nil {
		logger.Fatalf("TLS enabled flag is not provided '%v' Error: %v", configVal("ENABLE_TLS"), err)
	}
	insecureSkipVerify, err := strconv.ParseBool(configVal("INSECURE_SKIP_VERIFY"))
	if err != nil {
		logger.Fatalf("InsecureSkipVerify flag is not provided '%v' Error: %v", configVal("INSECURE_SKIP_VERIFY"), err)
	}

	lc := ldap.NewClient(configVal("BIND_USERNAME"), configVal("BIND_PASSWORD"), configVal("LDAP_URL"))
	searcher := group.NewSearcher(lc)
	searchDetails := &group.SearchDetails{
		Attributes:         strings.Split(configVal("ATTRIBUTES"), ","),
		BaseDn:             configVal("BASE_DN"),
		SearchFilter:       configVal("SEARCH_FILTER"),
		SearchTimeout:      searchTimeout,
		GroupAttribute:     configVal("GROUP_ATTRIBUTE"),
		EnableTLS:          tlsEnabledFlag,
		InsecureSkipVerify: insecureSkipVerify,
	}

	packDef := flyte.PackDef{
		Name: "ldap",
		Commands: []flyte.Command{
			command.GetGroupsCommand(searcher, searchDetails),
		},
		HelpURL: createURL("https://github.com/ExpediaGroup/flyte-ldap/blob/master/README.md"),
	}

	pack := flyte.NewPack(packDef, client.NewClient(createURL(configVal("FLYTE_API_URL")), 10*time.Second))

	pack.Start()

	blockForever()
}

func configVal(k string) string {
	v := env.getValueFor(k)
	if v == "" {
		logger.Fatalf("Config value %q must be set", k)
	}
	return v
}

func optionalConfigVal(k string, defaultVal string) string {
	v := env.getValueFor(k)
	if v == "" {
		return defaultVal
	}
	return v
}

func createURL(u string) *url.URL {
	url, err := url.Parse(u)
	if err != nil {
		logger.Fatalf("Cannot parse url: '%s' error: '%s'", u, err.Error())
	}
	return url
}

func blockForever() {
	select {}
}
