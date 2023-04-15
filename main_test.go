package main

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	//	"regexp"
)

func TestGenerateRequestID(t *testing.T) {
	msg := GenerateRequestID()
	res := strings.Count(msg, "-")
	if len(msg) != 36 || res != 4 {
		t.Fatalf("GenerateRequestID() result incorrect")
	}
}

func TestCreateNewEntry(t *testing.T) {
	var newEntry = createNewEntry("1234567")
	if newEntry.Count != 0 || strings.Compare(newEntry.Number, "1234567") != 0 {
		t.Fatalf("GenerateRequestID() result incorrect")
	}
}

func TestCreateResponse(t *testing.T) {
	var newEntry IdEntry
	newEntry.RequestId = "testID"
	newEntry.Code = 1234
	var senRes = createResponse(newEntry)
	if senRes.Code != 1234 || strings.Compare(senRes.RequestId, "testID") != 0 {
		t.Fatalf("GenerateRequestID() result incorrect")
	}
}

func TestRenderJSON(t *testing.T) {
	var newEntry IdEntry
	newEntry.RequestId = "testID"
	newEntry.Code = 1234
	var w http.ResponseWriter
	var test = renderJSON(w, newEntry)
	if len(test) < 1 {
		t.Fatalf("GenerateRequestID() result incorrect")
	}
}

func TestGenerateCheckCode(t *testing.T) {
	code, _ := strconv.Atoi(GenerateCheckCode())
	if code < 1 {
		t.Fatalf("GenerateRequestID() result incorrect")
	}
}
