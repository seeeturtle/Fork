package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joshua1b/SchoolMeal/app/handler"
)

type Response struct {
	Text string `json:"text"`
}

type Test struct {
	Value    []string
	Expected string
}

func TestCreateMessage(t *testing.T) {
	tests := []Test{
		Test{[]string{"text", "오늘 맛있는 급식"}, "오늘은 Test 나와요!\n빨리 먹고 싶네요!"},
		Test{[]string{"photo", "Test"}, "문자가 아닙니다."},
		Test{[]string{"text", "내일 맛있는 급식"}, ""},
	}

	for _, test := range tests {
		str := fmt.Sprintf(`{"user_key":"encypted", "type":"%s", "content":"%s"}`, test.Value[0], test.Value[1])
		jsonStr := []byte(str)
		req, err := http.NewRequest("POST", "/message", bytes.NewBuffer(jsonStr))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h := http.HandlerFunc(CreateMessage)
		h.ServeHTTP(rr, req)

		var response Response
		decoder := json.NewDecoder(rr.Body)
		if err := decoder.Decode(&response); err != nil {
			t.Fatal(err)
		}
		if response.Text != test.Expected {
			t.Errorf("handler returned unexpected text: got %v want %v", response.Text, test.Expected)
		}
	}
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	handler.CreateMessage(&sql.DB{}, w, r)
}
