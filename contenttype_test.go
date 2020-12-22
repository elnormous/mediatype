package contenttype

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestGetMediaType(t *testing.T) {
	tables := []struct {
		header     string
		result     string
		parameters map[string]string
		err        error
	}{
		{"", "", map[string]string{}, nil},
		{"application/json", "application/json", map[string]string{}, nil},
		{"*/*", "*/*", map[string]string{}, nil},
		{"Application/JSON", "application/json", map[string]string{}, nil},
		{"Application", "", nil, InvalidContentTypeError},
		{"/Application", "", nil, InvalidContentTypeError},
		{"Application/JSON/test", "", nil, InvalidParameterError},
		{" application/json ", "application/json", map[string]string{}, nil},
		{"Application/XML;charset=utf-8", "application/xml", map[string]string{"charset": "utf-8"}, nil},
		{"application/xml;foo=bar ", "application/xml", map[string]string{"foo": "bar"}, nil},
		{"application/xml ; foo=bar ", "application/xml", map[string]string{"foo": "bar"}, nil},
		{"application/xml;=bar ", "", nil, InvalidParameterError},
		{"application/xml; =bar ", "", nil, InvalidParameterError},
		{"application/xml;foo= ", "", nil, InvalidParameterError},
		{"application/xml;foo=\"bar\" ", "application/xml", map[string]string{"foo": "bar"}, nil},
		{"application/xml;foo=\"\" ", "application/xml", map[string]string{"foo": ""}, nil},
		{"application/xml;foo=\"\\\"b\" ", "application/xml", map[string]string{"foo": "\"b"}, nil},
		{"a/b+c;a=b;c=d", "a/b+c", map[string]string{"a": "b", "c": "d"}, nil},
		{"a/b;A=B", "a/b", map[string]string{"a": "b"}, nil},
	}

	for _, table := range tables {
		request, err := http.NewRequest(http.MethodGet, "http://test.test", nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(table.header) > 0 {
			request.Header.Set("Content-Type", table.header)
		}
		result, parameters, err := GetMediaType(request)
		if table.err != nil {
			if err == nil {
				t.Errorf("Expected an error for %s", table.header)
			} else if table.err != err {
				t.Errorf("Unexpected error \"%s\", expected \"%s\"", err.Error(), table.err.Error())
			}
		} else if table.err == nil && err != nil {
			t.Errorf("Got an unexpected error \"%s\" for %s", err.Error(), table.header)
		} else if result != table.result {
			t.Errorf("Invalid content type, got %s, exptected %s", result, table.result)
		} else if !reflect.DeepEqual(parameters, table.parameters) {

			t.Errorf("Wrong parameters, got %v, expected %v", fmt.Sprint(parameters), fmt.Sprint(table.parameters))
		}
	}
}
