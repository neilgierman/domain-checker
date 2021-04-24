package main

import (
	"context"
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
	a.Initialize(a.appCfg.Database.Host,a.appCfg.Database.Port,a.appCfg.Database.Database + "-test")
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
	body, _ := resp.Body.ReadString('\n')
	if body != "unknown" {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", body)
	}
}

func TestBounceAndReport(t *testing.T) {
	clearCollection()
	createExampleRecord()
	req, _ := http.NewRequest("PUT", "/events/example.com/bounced", nil)
	resp := executeRequest(req)
	checkResponseCode(t, http.StatusAccepted, resp.Code)
	time.Sleep(2 * time.Second)
	req, _ = http.NewRequest("GET", "/domains/example.com", nil)
	resp = executeRequest(req)
	checkResponseCode(t, http.StatusOK, resp.Code)
	body, _ := resp.Body.ReadString('\n')
	if body != "not catch-all" {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", body)
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
	body, _ := resp.Body.ReadString('\n')
	if body != "catch-all" {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", body)
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
	body, _ := resp.Body.ReadString('\n')
	if body != "unknown" {
		t.Errorf("Expected %s does not match actual %s\n", "unknown", body)
	}
}