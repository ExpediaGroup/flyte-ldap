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

package command

import (
	"encoding/json"
	"errors"
	"github.com/ExpediaGroup/flyte-ldap/group"
	"reflect"
	"strings"
	"testing"
)

func TestGetGroupsCommand_shouldReturnGroupsUserIsAMemberOf(t *testing.T) {
	mockSearcher := &mockSearcher{
		groupsToReturn: func(sd *group.SearchDetails, username string) ([]string, error) {
			return []string{"group1", "group2"}, nil
		},
	}

	command := GetGroupsCommand(mockSearcher, someSearchDetails())
	event := command.Handler(json.RawMessage(`{"username": "carlos"}`))

	if event.EventDef != getGroupsSuccessEventDef {
		t.Errorf("EventDef is wrong! EventDef: %v", event.EventDef)
	}
	payload := event.Payload.(userGroupsPayload)
	if payload.Username != "carlos" {
		t.Errorf("Username is wrong! Username: %v", payload.Username)
	}
	if payload.UserGroups[0] != "group1" || payload.UserGroups[1] != "group2" {
		t.Errorf("The groups returned are wrong! Usergroups: %v", payload.UserGroups)
	}
}

func TestGetGroupsCommand_shouldPassSearchDetailsDirectlyToTheSearcherWithoutModification(t *testing.T) {
	var searchDetailsPassedToSearcher *group.SearchDetails
	mockSearcher := &mockSearcher{
		groupsToReturn: func(sd *group.SearchDetails, username string) ([]string, error) {
			searchDetailsPassedToSearcher = sd
			return []string{"group1", "group2"}, nil
		},
	}
	searchDetails := &group.SearchDetails{
		Attributes:     []string{"memberOf"},
		BaseDn:         "cn=blah-blah,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,",
		SearchFilter:   "(mailNickname=%s)",
		GroupAttribute: "cn",
		SearchTimeout:  20,
	}

	command := GetGroupsCommand(mockSearcher, searchDetails)
	command.Handler(json.RawMessage(`{"username": "carlos"}`))

	if !reflect.DeepEqual(searchDetailsPassedToSearcher, searchDetails) {
		t.Errorf("SearchDetails passed to searcher does not match the search details passed into GetGroupsFor. Expected: '%+v', actual: '%+v'", searchDetails, searchDetailsPassedToSearcher)
	}
}

func TestGetGroupsCommand_shouldReturnFatalErrorEventForJsonUnmarshallingError(t *testing.T) {
	mockSearcher := &mockSearcher{}
	command := GetGroupsCommand(mockSearcher, someSearchDetails())
	event := command.Handler(json.RawMessage(`{"dodgy-json`))

	if event.EventDef.Name != "FATAL" {
		t.Errorf("EventDef is wrong! EventDef: %v", event.EventDef)
	}
	payload := event.Payload.(userGroupsPayload)
	if !strings.Contains(payload.ErrorText, "Json unmarshalling error: ") {
		t.Errorf("Error text is wrong! Error text: %v", payload.ErrorText)
	}
}

func TestGetGroupsCommand_shouldReturnErrorEventIfUsernameNotProvided(t *testing.T) {
	mockSearcher := &mockSearcher{}
	command := GetGroupsCommand(mockSearcher, someSearchDetails())
	event := command.Handler(json.RawMessage(`{}`))

	if event.EventDef != getGroupsErrorEventDef {
		t.Errorf("EventDef is wrong! EventDef: %v", event.EventDef)
	}
	payload := event.Payload.(userGroupsPayload)
	if payload.ErrorText != "No Username provided." {
		t.Errorf("Error text is wrong! Error text: %v", payload.ErrorText)
	}
}

func TestGetGroupsCommand_shouldReturnErrorEventIfClientReturnsError(t *testing.T) {
	mockSearcher := &mockSearcher{
		groupsToReturn: func(sd *group.SearchDetails, username string) ([]string, error) {
			return nil, errors.New("Search went wrong!!")
		},
	}

	command := GetGroupsCommand(mockSearcher, someSearchDetails())
	event := command.Handler(json.RawMessage(`{"username": "carlos"}`))

	if event.EventDef != getGroupsErrorEventDef {
		t.Errorf("EventDef is wrong! EventDef: %v", event.EventDef)
	}
	payload := event.Payload.(userGroupsPayload)
	if payload.Username != "carlos" {
		t.Errorf("Username is wrong! Username: %v", payload.Username)
	}
	if payload.ErrorText != "Search went wrong!!" {
		t.Errorf("Error text is wrong! Error text: %v", payload.ErrorText)
	}
}

func someSearchDetails() *group.SearchDetails {
	return &group.SearchDetails{
		Attributes:     []string{"memberOf"},
		BaseDn:         "cn=blah-blah,OU=User Policies,OU=All Users,DC=FAE,DC=CORPORATE,",
		SearchFilter:   "(mailNickname={username})",
		GroupAttribute: "cn",
		SearchTimeout:  20,
	}
}

type mockSearcher struct {
	groupsToReturn func(*group.SearchDetails, string) ([]string, error)
}

func (c *mockSearcher) GetGroupsFor(sd *group.SearchDetails, username string) ([]string, error) {
	return c.groupsToReturn(sd, username)
}
