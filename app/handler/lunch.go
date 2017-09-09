package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"database/sql"

	"github.com/joshua1b/SchoolMeal/app/model"
	_ "github.com/lib/pq"
)

type Response struct {
	Text string `json:"text"`
}

func GetKeyboard(w http.ResponseWriter, r *http.Request) {
	keyboard := struct {
		Type    string   `json:"type"`
		Buttons []string `json:"buttons"`
	}{
		"buttons",
		[]string{"오늘 급식", "내일 급식"},
	}
	respondJSON(w, http.StatusOK, keyboard)
}

func CreateMessage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var message struct {
		UserKey string
		Type    string
		Content string
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	var response Response

	commands := `
	[범위] [맛있는] 급식
	[범위]-   오늘, 내일, 다음주, 이번주, 이번달, 다음달 중
	하나 선택. 기본값은 오늘.
	[맛있는]- 있을때에는 급식중 맛있는 급식만 보여준다.
	기본값은 전체 보여주기.
	ex. 오늘 맛있는 급식

	도움말
	이 도움말 출력.
	`
	scopes := []string{"오늘, 내일, 다음주, 이번주, 이번달, 다음달"}
	patternScope := strings.Join(scopes, "|")
	pattern := fmt.Sprintf("(%s)(맛있는)?급식", patternScope)
	filterd := strings.Replace(strings.TrimSpace(message.Content), " ", "", -1)
	matched, _ := regexp.MatchString(pattern, filterd)

	switch {
	case message.Type != "text":
		response.Text = "문자가 아닙니다."
	case message.Content == "도움말":
		response.Text = commands
	case matched == true:
		re := regexp.MustCompile(patternScope)
		delicious, _ := regexp.MatchString(fmt.Sprintf("(%s)(맛있는)급식", patternScope), filterd)
		scope := re.FindString(response.Text)
		if delicious == true {
			getResponse(db, &response, scope, true)
		} else {
			getResponse(db, &response, scope, false)
		}
	default:
		response.Text = "제대로 된 명령어가 아닙니다."
	}
	respondJSON(w, http.StatusOK, response)
}

func getResponse(db *sql.DB, res *Response, scope string, delicious bool) {
	switch scope {
	case "오늘":
		t := time.Now()
		loc, _ := time.LoadLocation("Asia/Seoul")
		t = t.In(loc)
		if delicious {
			delicious_foods, err := model.GetDeliciousFoods(db, &t)
			if err != nil {
				res.Text = "데이터가 없습니다."
			}
			dels := strings.Join(delicious_foods, "\n")
			res.Text = "오늘 맛있는 급식은\n" + dels + "\n입니다."
			return
		}
		foods, err := model.GetFoods(db, &t)
		if err != nil {
			res.Text = "데이터가 없습니다."
			return
		}
		f := strings.Join(foods, "\n")
		res.Text = "오늘 급식은\n" + f + "\n입니다."
		return
	}
}
