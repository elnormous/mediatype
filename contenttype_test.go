package contenttype

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestGetMediaType(t *testing.T) {
	testCases := []struct {
		header     string
		result     MediaType
		parameters Parameters
	}{
		{"", MediaType{}, Parameters{}},
		{"application/json", MediaType{"application", "json"}, Parameters{}},
		{"*/*", MediaType{"*", "*"}, Parameters{}},
		{"Application/JSON", MediaType{"application", "json"}, Parameters{}},
		{" application/json ", MediaType{"application", "json"}, Parameters{}},
		{"Application/XML;charset=utf-8", MediaType{"application", "xml"}, Parameters{"charset": "utf-8"}},
		{"application/xml;foo=bar ", MediaType{"application", "xml"}, Parameters{"foo": "bar"}},
		{"application/xml ; foo=bar ", MediaType{"application", "xml"}, Parameters{"foo": "bar"}},
		{"application/xml;foo=\"bar\" ", MediaType{"application", "xml"}, Parameters{"foo": "bar"}},
		{"application/xml;foo=\"\" ", MediaType{"application", "xml"}, Parameters{"foo": ""}},
		{"application/xml;foo=\"\\\"b\" ", MediaType{"application", "xml"}, Parameters{"foo": "\"b"}},
		{"application/xml;foo=\"\\\"B\" ", MediaType{"application", "xml"}, Parameters{"foo": "\"b"}},
		{"a/b+c;a=b;c=d", MediaType{"a", "b+c"}, Parameters{"a": "b", "c": "d"}},
		{"a/b;A=B", MediaType{"a", "b"}, Parameters{"a": "b"}},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Content-Type", testCase.header)
		}

		result, parameters, mediaTypeError := GetMediaType(request)
		if mediaTypeError != nil {
			t.Errorf("Unexpected error for %s: %s", testCase.header, mediaTypeError.Error())
		} else if result != testCase.result {
			t.Errorf("Invalid content type, got %s, exptected %s", result, testCase.result)
		} else if !reflect.DeepEqual(parameters, testCase.parameters) {
			t.Errorf("Wrong parameters, got %v, expected %v", fmt.Sprint(parameters), fmt.Sprint(testCase.parameters))
		}
	}
}

func TestGetMediaTypeErrors(t *testing.T) {
	testCases := []struct {
		header string
		err    error
	}{
		{"Application", InvalidMediaTypeError},
		{"/Application", InvalidMediaTypeError},
		{"Application/", InvalidMediaTypeError},
		{"Application/JSON/test", InvalidMediaTypeError},
		{"application/xml;=bar ", InvalidParameterError},
		{"application/xml; =bar ", InvalidParameterError},
		{"application/xml;foo= ", InvalidParameterError},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Content-Type", testCase.header)
		}

		_, _, mediaTypeError := GetMediaType(request)
		if mediaTypeError == nil {
			t.Errorf("Expected an error for %s", testCase.header)
		} else if testCase.err != mediaTypeError {
			t.Errorf("Unexpected error \"%s\", expected \"%s\"", mediaTypeError.Error(), testCase.err.Error())
		}
	}
}

func TestGetAcceptableMediaType(t *testing.T) {
	testCases := []struct {
		header              string
		availableMediaTypes []MediaType
		result              MediaType
		parameters          Parameters
	}{
		{"", []MediaType{{"application", "json"}}, MediaType{"application", "json"}, Parameters{}},
		{"application/json", []MediaType{{"application", "json"}}, MediaType{"application", "json"}, Parameters{}},
		{"Application/Json", []MediaType{{"application", "json"}}, MediaType{"application", "json"}, Parameters{}},
		{"text/plain,application/xml", []MediaType{{"text", "plain"}}, MediaType{"text", "plain"}, Parameters{}},
		{"text/plain,application/xml", []MediaType{{"application", "xml"}}, MediaType{"application", "xml"}, Parameters{}},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Accept", testCase.header)
		}

		result, parameters, mediaTypeError := GetAcceptableMediaType(request, testCase.availableMediaTypes)

		if mediaTypeError != nil {
			t.Errorf("Unexpected error for %s: %s", testCase.header, mediaTypeError.Error())
		} else if result != testCase.result {
			t.Errorf("Invalid content type, got %s, exptected %s", result, testCase.result)
		} else if !reflect.DeepEqual(parameters, testCase.parameters) {
			t.Errorf("Wrong parameters, got %v, expected %v", fmt.Sprint(parameters), fmt.Sprint(testCase.parameters))
		}
	}
}

func TestGetAcceptableMediaTypeErrors(t *testing.T) {
	testCases := []struct {
		header              string
		availableMediaTypes []MediaType
		err                 error
	}{
		{"", []MediaType{}, NoAvailableTypeGivenError},
		{"application/xml", []MediaType{{"application", "json"}}, NoAcceptableTypeFoundError},
		{"application/xml/", []MediaType{{"application", "json"}}, InvalidMediaRangeError},
		{"application/xml,", []MediaType{{"application", "json"}}, InvalidMediaTypeError},
		{"/xml", []MediaType{{"application", "json"}}, InvalidMediaTypeError},
		{"application/,", []MediaType{{"application", "json"}}, InvalidMediaTypeError},
	}

	for _, testCase := range testCases {
		request, requestError := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if requestError != nil {
			log.Fatal(requestError)
		}

		if len(testCase.header) > 0 {
			request.Header.Set("Accept", testCase.header)
		}

		_, _, mediaTypeError := GetAcceptableMediaType(request, testCase.availableMediaTypes)
		if mediaTypeError == nil {
			t.Errorf("Expected an error for %s", testCase.header)
		} else if testCase.err != mediaTypeError {
			t.Errorf("Unexpected error \"%s\", expected \"%s\"", mediaTypeError.Error(), testCase.err.Error())
		}
	}
}
