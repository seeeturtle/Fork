package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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
		[]string{"도움말", "시작하기"},
	}
	respondJSON(w, http.StatusOK, keyboard)
}

func CreateMessage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var message struct {
		UserKey string `json:"user_key"`
		Type    string `json:"type"`
		Content string `json:"content"`
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
	scopes := []string{"오늘", "내일", "다음주", "이번주", "이번달", "다음달"}
	patternScope := strings.Join(scopes, "|")
	pattern := fmt.Sprintf("(%s) (맛있는 )?급식", patternScope)
	matched, _ := regexp.MatchString(pattern, message.Content)

	switch {
	case message.Type != "text":
		response.Text = "문자가 아닙니다."
	case message.Content == "도움말":
		response.Text = commands
	case matched == true:
		delicious, _ := regexp.MatchString(fmt.Sprintf("(%s) 맛있는 급식", patternScope), message.Content)
		scope := strings.Split(message.Content, " ")[0]
		if delicious {
			response.Text = getResponseText(db, scope, true)
		} else {
			response.Text = getResponseText(db, scope, false)
		}
	default:
		response.Text = "제대로 된 명령어가 아닙니다."
	}
	respondJSON(w, http.StatusOK, response)
}

func getResponseText(db *sql.DB, scope string, delicious bool) string {
	switch scope {
	case "오늘":
		loc, _ := time.LoadLocation("Asia/Seoul")
		today := time.Now().In(loc).Format("20060102")
		if delicious {
			var foods []model.DeliciousFood
			foods, _ = getDeliciousFoods(db, today, today)
			return getTodaysDeliciousFoods(foods)
		}
		var foods []model.Food
		foods, _ = getFoods(db, today, today)
		return getTodaysFoods(foods)
	default:
		return ""
	}
}

func getTodaysFoods(foods []model.Food) string {
	names := make([]string, len(foods))
	for _, food := range foods {
		names = append(names, food.Name)
	}
	f := strings.Join(names, ",")
	text := "오늘은 " + f + " 나와요!"
	return text
}

func getTodaysDeliciousFoods(foods []model.DeliciousFood) string {
	names := make([]string, len(foods))
	for _, food := range foods {
		names = append(names, food.Name)
	}
	f := strings.Join(names, ",")
	text := "오늘은 " + f + " 나와요!\n빨리 먹고 싶네요!"
	return text
}

func getFoods(db *sql.DB, startDate string, endDate string) ([]model.Food, error) {
	return []model.Food{model.Food{Name: "Test"}}, nil
}

func getDeliciousFoods(db *sql.DB, startDate string, endDate string) ([]model.DeliciousFood, error) {
	return []model.DeliciousFood{model.DeliciousFood{Name: "Test"}}, nil
}
