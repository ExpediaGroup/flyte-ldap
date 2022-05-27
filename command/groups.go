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
	"github.com/ExpediaGroup/flyte-ldap/group"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
)

const getGroupsCommandName = "GetGroups"

var getGroupsSuccessEventDef = flyte.EventDef{Name: "GroupsRetrieved"}
var getGroupsErrorEventDef = flyte.EventDef{Name: "GroupsRetrievalError"}

type GetGroupsInput struct {
	UserName string `json:"username"`
}

type userGroupsPayload struct {
	Username   string   `json:"username,omitempty"`
	UserGroups []string `json:"usergroups,omitempty"`
	ErrorText  string   `json:"error,omitempty"`
}

func GetGroupsCommand(searcher group.Searcher, searchDetails *group.SearchDetails) flyte.Command {
	return flyte.Command{
		Name:    getGroupsCommandName,
		Handler: getGroupsHandler(searcher, searchDetails),
		OutputEvents: []flyte.EventDef{
			getGroupsSuccessEventDef,
			getGroupsErrorEventDef,
		},
	}
}

func getGroupsHandler(searcher group.Searcher, searchDetails *group.SearchDetails) flyte.CommandHandler {
	return func(input json.RawMessage) flyte.Event {
		// unmarshall input
		args := GetGroupsInput{}
		if err := json.Unmarshal(input, &args); err != nil {
			return flyte.NewFatalEvent(userGroupsPayload{
				ErrorText: "Json unmarshalling error: " + err.Error(),
			})
		}
		if args.UserName == "" {
			return NewGetGroupsErrorEvent("No Username provided.", "")
		}

		// group search
		userGroups, err := searcher.GetGroupsFor(searchDetails, args.UserName)
		if err != nil {
			logger.Debugf("Got error as %v", err)
			return NewGetGroupsErrorEvent(err.Error(), args.UserName)
		}
		logger.Debugf("Got the user groups as %v", userGroups)

		return flyte.Event{
			EventDef: getGroupsSuccessEventDef,
			Payload: userGroupsPayload{
				UserGroups: userGroups,
				Username:   args.UserName,
			},
		}
	}
}

func NewGetGroupsErrorEvent(errorText, username string) flyte.Event {
	return flyte.Event{
		EventDef: getGroupsErrorEventDef,
		Payload: userGroupsPayload{
			Username:  username,
			ErrorText: errorText,
		},
	}
}
