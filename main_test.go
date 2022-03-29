package main

import (
	"fmt"
	"log"
	"testing"
)

// Attributes received on new version creation
// map[dataFormat:JSON_API_V1 eventType:SECRET_VERSION_ADD secretId:projects/827585297303/secrets/test-secret timestamp:2022-03-27T14:31:37.964139-07:00 versionId:projects/827585297303/secrets/test-secret/versions/2]

// received on a new version creation
const examplePayload = `{
  "name": "projects/827585297303/secrets/test-secret/versions/1",
  "createTime": "2022-03-27T21:14:28.798342Z",
  "state": "ENABLED",
  "replicationStatus": {
    "automatic": {}
  },
  "etag": "\"15db39ae618386\""
}`

const exampleDisable = `{
  "name": "projects/827585297303/secrets/test-secret/versions/1",
  "createTime": "2022-03-27T21:14:28.798342Z",
  "state": "DISABLED",
  "replicationStatus": {
    "automatic": {}
  },
  "etag": "\"15db39bde84736\""
}`

// {"name":"projects/827585297303/secrets/test-secret/versions/1","createTime":"2022-03-27T21:14:28.798342Z","destroyTime":"2022-03-27T21:22:02.044927824Z","state":"DESTROYED","replicationStatus":{"automatic":{}},"etag":"\"15db39c9669094\""}

const bodyOnDeletion = `{
  "name": "projects/827585297303/secrets/test-secret",
  "replication": {
    "automatic": {}
  },
  "createTime": "2022-03-27T21:10:23.332858Z",
  "topics": [
    {
      "name": "projects/fir-project-1fc35/topics/secrets.events"
    }
  ],
  "etag": "\"15db399fc001fa\""
}`

func TestSomething(t *testing.T) {
	var fullSecretID = "projects/827585297303/secrets/test-secret"

	var projectNumber int
	var secretID string
	_, err := fmt.Sscanf(fullSecretID, "projects/%d/secrets/%s",
		&projectNumber, &secretID)
	if err != nil {
		t.Fatalf("%v", err)
	}
	log.Println("projectNumber =", projectNumber)
	log.Println("secretID =", secretID)

}
