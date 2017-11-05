package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jinzhu/now"
	"github.com/joshua1b/Fork/app/model"
	"github.com/sirupsen/logrus"
)

type Scope interface {
	Beginning() string
	End() string
	Name() string
	FoodMessage([]model.Lunch) string
	DeliciousFoodMessage([]model.Lunch) string
}

type Today struct {
	name string
}

type Tomorrow struct {
	name string
}

type ThisWeek struct {
	name string
}

type NextWeek struct {
	name string
}

type ThisMonth struct {
	name string
}

type NextMonth struct {
	name string
}

type Response map[string]map[string]string

type Message struct {
	UserKey string `json:"user_key"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

const (
	NoData           = "급식이 없어"
	NotText          = "나는 글자 밖에 못 읽어!"
	CannotUnderstand = "뭐라는 거지... 미안, 내가 좀 멍청해."
)

const timeForm string = "20060102"

const LocForm string = "Asia/Seoul"

var Scopes = []Scope{
	Today{name: "오늘"},
	Tomorrow{name: "내일"},
	NextWeek{name: "다음주"},
	ThisWeek{name: "이번주"},
	ThisMonth{name: "이번달"},
	NextMonth{name: "다음달"},
}

var (
	Expected         = []string{"맛있", "급식", "점심"}
	Loc, _           = time.LoadLocation(LocForm)
	m        Message = Message{}
	log              = logrus.New()
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

	가능한 범위는 오늘, 내일, 이번주, 다음주, 이번달, 다음달이야.

	꼭 점심, 급식이라는 단어는 있어야해.

	그리고 맛있는 급식만 원하면 '맛있'이 꼭 안에 있어야해.

	예시 문장으로는
	- 오늘 맛있는 급식 알려줘.
	- 내일 급식 맛있는 게 뭐 있지?
	- 오늘 급식

	문의 기능은 아직... 학교에서 문의해줘.
	`
	ok, delicious, date := parseContent(m.Content)

	switch {
	case m.Type != "text":
		text = NotText
	case m.Content == "도와줘":
		text = help
	case m.Content == "시작!":
		text = "자! 어떤 급식이 궁금하니?"
	case ok && (date != ""):
		if delicious {
			text = getResponseText(date, true)
		} else {
			text = getResponseText(date, false)
		}
	default:
		text = CannotUnderstand
	}
	response["message"] = make(map[string]string)
	response["message"]["text"] = text
	respondJSON(w, http.StatusOK, response)
}

func parseContent(str string) (ok bool, delicious bool, date string) {
	expected := make(map[string]bool, len(Expected))
	for _, e := range Expected {
		expected[e] = false
	}
	for _, s := range Scopes {
		expected[s.Name()] = false
	}
	splitted := strings.Split(str, " ")
	for _, w := range splitted {
		for k := range expected {
			if strings.Contains(w, k) {
				expected[k] = true
			}
		}
	}
	for k, b := range expected {
		if !b {
			continue
		}
		switch k {
		case "점심":
			ok = true
		case "급식":
			ok = true
		case "오늘":
			date = k
		case "내일":
			date = k
		case "다음주":
			date = k
		case "이번주":
			date = k
		case "이번달":
			date = k
		case "다음달":
			date = k
		case "맛있":
			delicious = true
		}
	}
	return
}

func getResponseText(scope string, delicious bool) string {
	for _, s := range Scopes {
		if scope == s.Name() {
			return message(s, delicious)
		}
	}
	log.WithFields(logrus.Fields{
		"user_key": m.UserKey,
		"scope":    scope,
	}).Info("Fields doesn't support")
	return CannotUnderstand
}

func JoinWithComma(lunches []model.Lunch) string {
	var str string
	for _, lunch := range lunches {
		names := []string{}
		for _, food := range lunch.Foods {
			names = append(names, food.Name)
		}
		dateStr := getDateStr(lunch.Date)
		str += dateStr + strings.Join(names, ", ") + getPostposition(names[len(names)-1]) + ",\n"
	}
	return str
}

func getDateStr(date string) string {
	dateTime, _ := time.Parse(timeForm, date)
	dateTime = roundTime(dateTime.In(Loc))
	now := roundTime(time.Now().In(Loc))
	duration := dateTime.Sub(now)
	diffDays := int(duration.Hours() / 24)
	switch diffDays {
	case 0:
		return "오늘은 "
	case 1:
		return "내일은 "
	case 2:
		return "모레는 "
	case 3:
		return "글피는 "
	default:
		return fmt.Sprintf("%d월 %d일은 ", dateTime.Month(), dateTime.Day())
	}
}

func roundTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func getPostposition(str string) string {
	defaultStr := "가"
	s := []rune(str)
	lastCharacter := string(s[len(s)-1])
	r, _ := utf8.DecodeRuneInString(lastCharacter)
	jongSeongCode := (int(r) - 44032) % 28
	if jongSeongCode != 0 {
		defaultStr = "이"
	}
	return defaultStr
}

func message(s Scope, delicious bool) string {
	beginning := s.Beginning()
	end := s.End()
	if delicious {
		deliciousLunches, err := model.Lunches.GetDelicious(beginning, end)
		if err != nil {
			log.WithFields(logrus.Fields{
				"user_key": m.UserKey,
				"error":    err,
			}).Warn("error from getting lunches")
			if beginning == end {
				return getDateStr(beginning) + " " + NoData
			}
			return NoData
		}
		return s.DeliciousFoodMessage(deliciousLunches)
	}
	lunches, err := model.Lunches.Get(beginning, end)
	if err != nil {
		log.WithFields(logrus.Fields{
			"user_key": m.UserKey,
			"error":    err,
		}).Warn("error from getting lunches")
		if beginning == end {
			return getDateStr(beginning) + " " + NoData
		}
		return NoData
	}
	return s.FoodMessage(lunches)
}

func (t Today) Beginning() string {
	return time.Now().In(Loc).Format(timeForm)
}

func (t Today) End() string {
	return time.Now().In(Loc).Format(timeForm)
}

func (t Today) Name() string {
	return t.name
}

func (t Today) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당"
}

func (t Today) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!!"
}

func (to Tomorrow) Beginning() string {
	return time.Now().In(Loc).AddDate(0, 0, 1).Format(timeForm)
}

func (to Tomorrow) End() string {
	return time.Now().In(Loc).AddDate(0, 0, 1).Format(timeForm)
}

func (to Tomorrow) Name() string {
	return to.name
}

func (to Tomorrow) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당"
}

func (to Tomorrow) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!!"
}

func (tw ThisWeek) Beginning() string {
	return now.BeginningOfWeek().In(Loc).Format(timeForm)
}

func (tw ThisWeek) End() string {
	return now.EndOfWeek().In(Loc).Format(timeForm)
}

func (tw ThisWeek) Name() string {
	return tw.name
}

func (tw ThisWeek) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당. 이번주에는 그럭저럭 하네."
}

func (tw ThisWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!! 이번주에 맛있는게 많다!"
}

func (nw NextWeek) Beginning() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 0, 7))
	return n.BeginningOfWeek().In(Loc).Format(timeForm)
}

func (nw NextWeek) End() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 0, 7))
	return n.EndOfWeek().In(Loc).Format(timeForm)
}

func (nw NextWeek) Name() string {
	return nw.name
}

func (nw NextWeek) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당. 다음주에는 이런게 나온데."
}

func (nm NextWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!! 다음주가 기대된다!!!"
}

func (tm ThisMonth) Beginning() string {
	return now.BeginningOfMonth().In(Loc).Format(timeForm)
}

func (tm ThisMonth) End() string {
	return now.EndOfMonth().In(Loc).Format(timeForm)
}

func (tm ThisMonth) Name() string {
	return tm.name
}

func (tm ThisMonth) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당. 이번달 급식표임"
}

func (tm ThisMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당. 이런건 별표 다섯개 밑줄 쫙쫙!"
}

func (nm NextMonth) Beginning() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 1, 0))
	return n.BeginningOfMonth().In(Loc).Format(timeForm)
}

func (nm NextMonth) End() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 1, 0))
	return n.EndOfMonth().In(Loc).Format(timeForm)
}

func (nm NextMonth) Name() string {
	return nm.name
}

func (nm NextMonth) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당."
}

func (nm NextMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온당. 다음달 기대되지?"
}
