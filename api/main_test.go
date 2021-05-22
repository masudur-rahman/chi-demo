package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type testData struct {
	method     string
	url        string
	status     int
	middleware func(next http.Handler) http.Handler
	handler    http.HandlerFunc
	path       string
	body       io.Reader
	response   string
}

func init() {
	StartXormEngine()
	CreateInitialEngineerProfile()
}

func TestShowAllEngineers(t *testing.T) {

	test := []testData{
		{
			"GET",
			"/engineers",
			200,
			nil,
			ShowAllEngineers,
			"/engineers",
			nil,
			`[{"username":"masud","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1},{"username":"fahim","firstname":"Fahim","lastname":"Abrar","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1},{"username":"tahsin","firstname":"Tahsin","lastname":"Rahman","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1},{"username":"jenny","firstname":"Jannatul","lastname":"Ferdows","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}]`,
		},
	}

	for _, data := range test {
		runTest(data, t)
	}

}

func TestShowSingleEngineer(t *testing.T) {
	test := []testData{
		{
			"GET",
			"/engineers/masud",
			200,
			UserCtx,
			ShowSingleEngineer,
			"/engineers/{username}",
			nil,
			`{"username":"masud","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}`,
		},
		{
			"GET",
			"/engineers/fahim",
			200,
			UserCtx,
			ShowSingleEngineer,
			"/engineers/{username}",
			nil,
			`{"username":"fahim","firstname":"Fahim","lastname":"Abrar","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}`,
		},
		{
			"GET",
			"/engineers/tahsin",
			200,
			UserCtx,
			ShowSingleEngineer,
			"/engineers/{username}",
			nil,
			`{"username":"tahsin","firstname":"Tahsin","lastname":"Rahman","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}`,
		},
		{
			"GET",
			"/engineers/jenny",
			200,
			UserCtx,
			ShowSingleEngineer,
			"/engineers/{username}",
			nil,
			`{"username":"jenny","firstname":"Jannatul","lastname":"Ferdows","city":"Chittagong","division":"Chittagong","position":"Software Engineer","CreatedAt":"2019-03-20T18:17:07+06:00","UpdatedAt":"2019-03-20T18:17:07+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}`,
		},
		{
			"GET",
			"/engineers/abcd",
			404,
			UserCtx,
			ShowSingleEngineer,
			"/engineers/{username}",
			nil,
			`"user:abcd not found"`,
		},
	}

	for _, data := range test {
		runTest(data, t)

	}

}

func TestAddNewEngineer(t *testing.T) {
	test := []testData{
		{
			"POST",
			"/engineers",
			409,
			UserCtx,
			AddNewEngineer,
			"/engineers",
			strings.NewReader(`{"username":"masud","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer"}`),
			`409 - username already exists`,
		},
		{
			"POST",
			"/engineers",
			201,
			UserCtx,
			AddNewEngineer,
			"/engineers",
			strings.NewReader(`{"username":"masudur","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer"}`),
			`{"username":"masudur","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer","CreatedAt":"2019-03-20T19:16:49.9130127+06:00","UpdatedAt":"2019-03-20T19:16:49.913024478+06:00","DeletedAt":"0001-01-01T00:00:00Z","Version":1}`,
		},
	}

	for _, data := range test {
		runTest(data, t)
	}

}

func TestUpdateEngineerProfile(t *testing.T) {
	test := []testData{
		{
			"POST",
			"/engineers/masud",
			405,
			UserCtx,
			UpdateEngineerProfile,
			"/engineers/{username}",
			strings.NewReader(`{"username":"masudd","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer"}`),
			`405 - Username can't be changed`,
		},
		{
			"POST",
			"/engineers/masudd",
			404,
			UserCtx,
			UpdateEngineerProfile,
			"/engineers/{username}",
			strings.NewReader(`{"username":"masudd","firstname":"Masudur","lastname":"Rahman","city":"Madaripur","division":"Dhaka","position":"Software Engineer"}`),
			`404 - Content Not Found`,
		},
		{
			"POST",
			"/engineers/masud",
			200,
			UserCtx,
			UpdateEngineerProfile,
			"/engineers/{username}",
			strings.NewReader(`{"username":"masud","firstname":"Masudur","lastname":"Rahman","city":"M","division":"D","position":"Software Engineer"}`),
			`201 - Updated successfully`,
		},
	}

	for _, data := range test {
		runTest(data, t)
	}
}

func TestDeleteEngineer(t *testing.T) {
	test := []testData{
		{
			"DELETE",
			"/engineers/masudur",
			200,
			UserCtx,
			DeleteEngineer,
			"/engineers/{username}",
			nil,
			`200 - Deleted Successfully`,
		},
		{
			"DELETE",
			"/engineers/hello",
			404,
			UserCtx,
			DeleteEngineer,
			"/engineers/{username}",
			nil,
			`404 - Content Not Found`,
		},
	}
	for _, data := range test {
		runTest(data, t)
	}
}

func runTest(test testData, t *testing.T) {
	req, err := http.NewRequest(test.method, test.url, test.body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bWFzdWQ6cGFzcw==")
	responseRecorder := httptest.NewRecorder()

	r := chi.NewRouter()

	r.Route(test.path, func(r chi.Router) {
		if test.middleware != nil {
			r.Use(test.middleware)
		}
		r.Method(test.method, "/", test.handler)
	})

	r.ServeHTTP(responseRecorder, req)

	if status := responseRecorder.Code; status != test.status {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, test.status)
	}

}
