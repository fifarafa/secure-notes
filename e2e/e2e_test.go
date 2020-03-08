package e2e_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CreateNoteResponse struct {
	Id string `json:"id"`
}

type GetNoteResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	TTL  int    `json:"ttl"`
}

func Test_CreateAndGetNote(t *testing.T) {
	createNoteResp := createNote(t)
	gotNote := getNote(t, createNoteResp.Id)

	assert.Equal(t, createNoteResp.Id, gotNote.ID)
	assert.Equal(t, "Hello World", gotNote.Text)
}

func createNote(t *testing.T) CreateNoteResponse {
	createNoteBody := `{
		"text": "Hello World",
		"lifeTimeSeconds": 360000,
		"password": "mySecretPassword"
	}`

	resp, err := http.Post("https://yyosn5pkag.execute-api.us-east-1.amazonaws.com/dev/notes", "application/json", strings.NewReader(createNoteBody))
	if err != nil {
		t.Fatalf("POST note: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read create note response: %v", err)
	}

	var r CreateNoteResponse
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		t.Fatalf("unmarshal create note response: %v", err)
	}
	return r
}

func getNote(t *testing.T, noteID string) GetNoteResponse {
	req, err := http.NewRequest(http.MethodGet, "https://yyosn5pkag.execute-api.us-east-1.amazonaws.com/dev/notes/"+noteID, http.NoBody)
	if err != nil {
		t.Fatalf("create GET note request: %v", err)
	}
	req.Header.Add("password", "mySecretPassword")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get note: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read get note reponse: %v", err)
	}

	var r GetNoteResponse
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		t.Fatalf("unmarshal get note response: %v", err)
	}
	return r
}
