package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"cambio/pkg/model"

	"github.com/google/go-cmp/cmp"
)

func TestInvalidJSON(t *testing.T) {
	invalidJSON := []string{
		`totally not json`,
		`{"key": "value", badKey: 6`,
	}
	errors := []string{
		`malformed json at position 2`,
		`malformed json at position 18`,
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestInvalidStructure(t *testing.T) {
	invalidJSON := []string{
		`{"diagnosisKeys": 42}`,
		`{"diagnosisKeys": ["41", 42]}`,
		`{"appPackageName": 4.5}`,
		`{"country": ["us", "ca"]}`,
		`{"badField": "doesn't exist"}`,
	}
	errors := []string{
		`invalid value diagnosisKeys at position 20`,
		`invalid value diagnosisKeys at position 27`,
		`invalid value appPackageName at position 22`,
		`invalid value country at position 13`,
		`unknown field "badField"`,
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestValidPublisMessage(t *testing.T) {
	json := `{"diagnosisKeys": ["ABC", "DEF", "123"],
    "appPackageName": "com.google.android.awesome",
    "country": "us",
    "platform": "android",
    "verificationPayload": "foo"}`

	body := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	got := &model.Publish{}
	err, code := unmarshal(w, r, got)
	if err != nil {
		t.Fatalf("unexpected err, %v", err)
	}
	if code != http.StatusOK {
		t.Errorf("unmarshal wanted %v response code, got %v", http.StatusOK, code)
	}

	want := &model.Publish{
		Keys:           []string{"ABC", "DEF", "123"},
		AppPackageName: "com.google.android.awesome",
		Country:        "us",
		Platform:       "android",
		Verification:   "foo",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%v", diff)
	}
}

func unmarshalTestHelper(t *testing.T, payloads []string, errors []string, expCode int) {
	for i, testStr := range payloads {
		body := ioutil.NopCloser(bytes.NewReader([]byte(testStr)))
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("content-type", "application/json")

		w := httptest.NewRecorder()
		data := &model.Publish{}
		err, code := unmarshal(w, r, data)
		if code != expCode {
			t.Errorf("unmarshal wanted %v response code, got %v", expCode, code)
		}
		if errors[i] == "" {
			// No error expected for this test, bad if we got one.
			if err != nil {
				t.Errorf("expected no error for `%v`, got: %v", testStr, err)
			}
		} else {
			if err == nil {
				t.Errorf("wanted error '%v', got nil", errors[i])
			} else if err.Error() != errors[i] {
				t.Errorf("expected error '%v', got: %v", errors[i], err)
			}
		}
	}
}