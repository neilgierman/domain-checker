package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var a App

func TestMain(m *testing.M) {
	a.loadConfig()
	err := a.Initialize(a.appCfg.Database.Host,a.appCfg.Database.Port,a.appCfg.Database.Database + "-test")
	if err != nil {
		log.Fatal(err)
	}
	go a.databaseBatchWriter()
	go a.queueProcessor()
	code := m.Run()
	clearCollection()
	os.Exit(code)
}

func clearCollection() {
	ctx , cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	result, err := a.dbConfig.writeDb.Collection("domains").DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(result.DeletedCount)
}

func createExampleRecord() {
	ctx , cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	result, err := a.dbConfig.writeDb.Collection("domains").InsertOne(ctx, bson.M{
		"domainName":"example.com",
		"bouncedCount":0,
		"deliveredCount":1000,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(result.InsertedID)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected %d does not match actual %d\n", expected, actual)
	}
}

func TestBadHost(t *testing.T) {
	// NOTE: a.LoadConfig is already run in TestMain so a.appCfg should already be populated
	err := a.Initialize("badhost.local",a.appCfg.Database.Port,a.appCfg.Database.Database + "-test")
	if err == nil {
		t.Error("Initialize should have failed with call to badhost but didn't.")
	}
}

func TestBadPort(t *testing.T) {
	// NOTE: a.LoadConfig is already run in TestMain so a.appCfg should already be populated
	err := a.Initialize(a.appCfg.Database.Host,"0",a.appCfg.Database.Database + "-test")
	if err == nil {
		t.Error("Initialize should have failed with call to invalid port but didn't.")
	}
}

func TestRedirect(t *testing.T) {
	req, _ := http.NewRequest("GET", "foo", nil)
	resp := executeRequest(req)

	checkResponseCode(t, http.StatusMovedPermanently, resp.Code)
}

func TestNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/foo/", nil)
	resp := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, resp.Code)
}

func TestUnreportedDomain(t *testing.T) {
	clearCollection()
	req, _ := http.NewRequest("GET", "/domains/example.com", nil)
	resp := executeRequest(req)

	checkResponseCode(t, http.StatusOK, resp.Code)
	var domainResult DomainResult;
	if err := json.Unmarshal(resp.Body.Bytes(), &domainResult); err != nil {
		t.Fatal(err)
	}
	if domainResult.Status != Unknown.String() {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", domainResult.Status)
	}
}

func TestBounceAndReport(t *testing.T) {
	clearCollection()
	createExampleRecord()
	req, _ := http.NewRequest("PUT", "/events/example.com/bounced", nil)
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusAccepted, resp.Code)
	time.Sleep(4 * time.Second)
	req, _ = http.NewRequest("GET", "/domains/example.com", nil)
	resp = executeRequest(req)
	checkResponseCode(t, http.StatusOK, resp.Code)
	var domainResult DomainResult;
	if err := json.Unmarshal(resp.Body.Bytes(), &domainResult); err != nil {
		t.Fatal(err)
	}
	if domainResult.Status != NotCatchAll.String() {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", domainResult.Status)
	}
}

func TestNoBounceAndReport(t *testing.T) {
	clearCollection()
	createExampleRecord()
	req, _ := http.NewRequest("PUT", "/events/example.com/delivered", nil)
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusAccepted, resp.Code)
	time.Sleep(2 * time.Second)
	req, _ = http.NewRequest("GET", "/domains/example.com", nil)
	resp = executeRequest(req)
	checkResponseCode(t, http.StatusOK, resp.Code)
	var domainResult DomainResult;
	if err := json.Unmarshal(resp.Body.Bytes(), &domainResult); err != nil {
		t.Fatal(err)
	}
	if domainResult.Status != CatchAll.String() {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", domainResult.Status)
	}
}

func TestFirstEntryAndReport(t *testing.T) {
	clearCollection()
	req, _ := http.NewRequest("PUT", "/events/example.com/delivered", nil)
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusAccepted, resp.Code)
	time.Sleep(2 * time.Second)
	req, _ = http.NewRequest("GET", "/domains/example.com", nil)
	resp = executeRequest(req)
	checkResponseCode(t, http.StatusOK, resp.Code)
	var domainResult DomainResult;
	if err := json.Unmarshal(resp.Body.Bytes(), &domainResult); err != nil {
		t.Fatal(err)
	}
	if domainResult.Status != Unknown.String() {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", domainResult.Status)
	}
}