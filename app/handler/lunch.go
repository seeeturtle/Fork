package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/joshua1b/Fork/app/model"
	"github.com/sirupsen/logrus"
)

type Response map[string]map[string]string

type Message struct {
	UserKey string `json:"user_key"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

const (
	Error            = "오류가 났어ㅠㅠ"
	NoData           = "급식은 없어"
	NotText          = "나는 글자 밖에 못 읽어!"
	CannotUnderstand = "뭐라는 거지... 미안, 내가 좀 멍청해."
)

var slangs = []string{
	"개",
	"걸레",
	"년",
	"놈",
	"느금마",
	"닥쳐",
	"등신",
	"또라이",
	"미친",
	"멍청",
	"병신",
	"새끼",
	"썅",
	"시발",
	"씨발",
	"씨팔",
	"씨발",
	"썖",
	"씹",
	"에미",
	"애미",
	"애비",
	"에비",
	"염병",
	"옘병",
	"좆",
	"좃",
	"좇",
	"지랄",
	"창",
	"호로",
	"후레",
	"호구",
	"후장",
}

var (
	loc, _ = time.LoadLocation("Asia/Seoul")
	m      = Message{}
	log    = logrus.New()
)

func GetKeyboard(w http.ResponseWriter, r *http.Request) {
	keyboard := struct {
		Type    string   `json:"type"`
		Buttons []string `json:"buttons"`
	}{
		"buttons",
		[]string{"도와줘", "시작!"},
	}
	respondJSON(w, http.StatusOK, keyboard)
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	now.FirstDayMonday = true
	log.Out = os.Stdout

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	response := make(Response)
	var text string

	help := `
	언제 급식을 원하는 건지,
	맛있는 급식을 원하는 건지 모든 급식을 원하는 건지 알려줘야해.

	가능한 범위는 오늘, 내일, 모레, 글피, 이번주, 다음주, 다다음주, 이번달, 다음달이야.

	꼭 점심, 급식이라는 단어는 있어야해.

	그리고 맛있는 급식만 원하면 '맛있'이 꼭 안에 있어야해.

	예시 문장으로는
	- 오늘 맛있는 급식 알려줘.
	- 내일 급식 맛있는 게 뭐 있지?
	- 오늘 급식
	
	아직 문의 기능이 없어.
	`
	ok, delicious, similar, slang, date := parseContent(m.Content)

	switch {
	case m.Type != "text":
		text = NotText
	case m.Content == "도와줘":
		text = help
	case m.Content == "시작!":
		text = "자! 어떤 급식이 궁금하니?"
	case slang:
		text = "내가 아무리 멍청해도 욕은 알아들어!"
	case ok && (date != "") && !similar:
		text = getResponseText(date, delicious)
	case date != "" && !delicious:
		text = date + ` 급식을 원하는거야? 그러면 "` + date + ` 급식" 이라고 말해줘.`
	case date != "" && delicious:
		text = date + ` 맛있는 급식을 원하는거야? 그러면 "` + date + ` 맛있는 급식" 이라고 말해줘.`
	case ok && (date == ""):
		text = "언제 급식을 원하는 거야?"
	default:
		text = CannotUnderstand
	}
	response["message"] = make(map[string]string)
	response["message"]["text"] = text
	respondJSON(w, http.StatusOK, response)
}

func getResponseText(scope string, delicious bool) string {
	for _, s := range Scopes {
		if strings.Contains(scope, s.Name()) {
			if s.Name() == "날짜" {
				dayTime, _ := time.Parse("20060102", scope)
				s.(*Day).date = dayTime.In(loc)
			}
			return message(s, delicious)
		}
	}
	log.WithFields(logrus.Fields{
		"user_key": m.UserKey,
		"scope":    scope,
	}).Info("Fields doesn't support")
	return CannotUnderstand
}

func message(s Scope, delicious bool) string {
	beginning := s.Beginning()
	end := s.End()
	if delicious {
		deliciousLunches, err := model.Lunches.GetDelicious(beginning, end)
		if len(deliciousLunches) == 0 {
			switch s.(type) {
			case *Day:
				dateTime := s.(*Day).date
				return fmt.Sprintf("%d월 %d일 ", dateTime.Month(), dateTime.Day()) + NoData
			default:
				return s.Name() + " " + NoData
			}
		}
		if err != nil {
			log.WithFields(logrus.Fields{
				"user_key": m.UserKey,
				"error":    err,
			}).Warn("error from getting lunches")
			return Error
		}
		return s.DeliciousFoodMessage(deliciousLunches)
	}
	lunches, err := model.Lunches.Get(beginning, end)
	if len(lunches) == 0 {
		switch s.(type) {
		case *Day:
			dateTime := s.(*Day).date
			return fmt.Sprintf("%d월 %d일 ", dateTime.Month(), dateTime.Day()) + NoData
		default:
			return s.Name() + " " + NoData
		}
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"user_key": m.UserKey,
			"error":    err,
		}).Warn("error from getting lunches")
		return Error
	}
	return s.FoodMessage(lunches)
}
