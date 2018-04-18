![Build Status](https://travis-ci.org/HotelsDotCom/flyte-ldap.svg?branch=master)


## Overview
The LDAP pack provides the ability to connect to and search directories, e.g. Microsoft Active Directory.

## Build and Run
### Command Line
To build and run from the command line:
* Clone this repo
* Run `dep ensure` (must have [dep](https://github.com/golang/dep) installed )
* Run `go build`
* Run `FLYTE_API_URL=<URL> BIND_USERNAME=<USERNAME> BIND_PASSWORD=<PASSWORD> LDAP_URL=<LDAP_URL> GROUP_ATTRIBUTE=<GROUP_ATTRIBUTE> ATTRIBUTES=<ATTRIBUTES> BASE_DN=<BASE_DN> SEARCH_FILTER=<SEARCH_FILTER> SEARCH_TIMEOUT_IN_SECONDS=<SEARCH_TIMEOUT_IN_SECONDS> ./flyte-ldap`
* All of these environment variables need to be provided with the exception of 'SEARCH_TIMEOUT_IN_SECONDS', which has a default (see main.go).
#### Example
* Run `FLYTE_API_URL='http://myflyteapi.com' BIND_USERNAME='someUsername' BIND_PASSWORD='somePassword' LDAP_URL='my.ldap.com:123' GROUP_ATTRIBUTE='cn' ATTRIBUTES='memberOf' BASE_DN='DC=QQ,DC=WOW,DC=XYZ,DC=com' SEARCH_FILTER='(mailNickname={username})' SEARCH_TIMEOUT_IN_SECONDS='20' ./flyte-ldap`

### Run tests
To run the unit tests:
* go test ./...

### Docker
To build and run from docker
* Run `docker build -t flyte-ldap .`
* Run `docker run -e FLYTE_API_URL=<URL> -e BIND_USERNAME=<USERNAME> -e BIND_PASSWORD=<PASSWORD> -e LDAP_URL=<LDAP_URL> -e GROUP_ATTRIBUTE=<GROUP_ATTRIBUTE> -e ATTRIBUTES=<ATTRIBUTES> -e BASE_DN=<BASE_DN> -e SEARCH_FILTER=<SEARCH_FILTER> -e SEARCH_TIMEOUT_IN_SECONDS=<SEARCH_TIMEOUT_IN_SECONDS> flyte-ldap`
* All of these environment variables need to be provided with the exception of 'SEARCH_TIMEOUT_IN_SECONDS', which has a default (see main.go).
#### Example
* Run `docker run -e FLYTE_API_URL='http://myflyteapi.com' -e BIND_USERNAME='someUsername' -e BIND_PASSWORD='somePassword' -e LDAP_URL='my.ldap.com:123' -e GROUP_ATTRIBUTE='cn' -e ATTRIBUTES='memberOf' -e BASE_DN='DC=QQ,DC=WOW,DC=XYZ,DC=com' -e SEARCH_FILTER='(mailNickname={username})' -e SEARCH_TIMEOUT_IN_SECONDS='20' flyte-ldap`


#### LDAP Attribute explanation
* GROUP_ATTRIBUTE - The attribute that gives the name of the group from the attribute values, e.g. 'cn'
* ATTRIBUTES - The attributes to be returned by the search, e.g. 'memberOf'
* BASE_DN -  This is the point from where a server will search for users
* SEARCH_FILTER - The criteria used to identify entries in search requests. In our example "SEARCH_FILTER='(mailNickname={username})'", the '{username}' will be replaced by the username passed in to the 'GetGroups' command

## Commands
This pack provides the 'GetGroups' command. This command retrieves the groups a user is a member of.
#### Input
The command input requires the 'username' of the user you want to search:
```
"input": {
    "username": "davyjones",
    }
```
#### Output
This command can either return a `GroupsRetrieved` event meaning the directory has been successfully searched or a 
`GroupsRetrievalError` event, meaning there was a problem.
##### GroupsRetrieved event 
This is the success event, it contains the command name, username and groups the user is a member of. It returns them 
in the form:
```
"payload": {
        "commandName": "GetGroups",
        "username": "davyjones",
        "usergroups": ["group1","group2"]
}
```
##### GroupsRetrievalError
This contains the normal output fields plus the error if the command fails:
```
"payload": {
        "commandName": "GetGroups",
        "username": "davyjones",
        "error": "Ldap connection meh."
}
```