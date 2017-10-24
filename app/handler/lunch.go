package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	nows "github.com/jinzhu/now"
	"github.com/joshua1b/Plate/app/model"
	_ "github.com/lib/pq"
)

type Scope interface {
	Message(bool, time.Time, *sql.DB) string
	Name() string
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

type Response struct {
	Text string `json:"text"`
}

const (
	NoData           = "아직 데이터가 없어..."
	NotText          = "나는 글자 밖에 못 읽어!"
	CannotUnderstand = "뭐라는 거지... 미안, 내가 좀 멍청해."
)

const formatText string = "20060102"

var Scopes = []Scope{
	Today{name: "오늘"},
	Tomorrow{name: "내일"},
	NextWeek{name: "다음주"},
	ThisWeek{name: "이번주"},
	ThisMonth{name: "이번달"},
	NextMonth{name: "다음달"},
}

var Expected = []string{"맛있", "급식", "점심"}

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
	ok, delicious, date := parseContent(message.Content)

	switch {
	case message.Type != "text":
		response.Text = NotText
	case message.Content == "도와줘":
		response.Text = help
	case ok && (date != ""):
		if delicious {
			response.Text = getResponseText(db, date, true)
		} else {
			response.Text = getResponseText(db, date, false)
		}
	default:
		response.Text = CannotUnderstand
	}
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

func getResponseText(db *sql.DB, scope string, delicious bool) string {
	loc, _ := time.LoadLocation("Asia/Seoul")
	now := time.Now().In(loc)
	for _, s := range Scopes {
		if scope == s.Name() {
			return s.Message(delicious, now, db)
		}
	}
	return CannotUnderstand
}

func TodayFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	text := "오늘은 " + f + " 나온다."
	return text
}

func TodayDeliciousFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	text := "오늘은 " + f + " 나온다!!!!!!"
	return text
}

func TomorrowFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	text := "내일은 " + f + "나온다."
	return text
}

func TomorrowDeliciousFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	text := "내일은 " + f + "나온다!!!!!!!"
	return text
}

func ThisWeekFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "이번주에는 이런게 있어."
}

func ThisWeekDeliciousFoodsMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "다음주에는 이런게 있어."
}

func JoinWithComma(lunches []model.Lunch) string {
	var str string
	for _, lunch := range lunches {
		names := make([]string, len(lunch.Foods))
		for _, food := range lunch.Foods {
			names = append(names, food.Name)
		}
		str += lunch.Date + "\n" + strings.Join(names, "\n")
	}
	return str
}

func getLunches(db *sql.DB, startDate string, endDate string) ([]model.Lunch, error) {
	return []model.Lunch{
		model.Lunch{
			Foods: []model.Food{model.Food{"Food"}},
		},
	}, nil
}

func getDeliciousFoods(db *sql.DB, startDate string, endDate string) ([]model.Lunch, error) {
	return []model.Lunch{
		model.Lunch{
			Foods: []model.DeliciousFood{model.DeliciousFood{"DeliciousFood"}},
		},
	}, nil
}

func getFood(db *sql.DB, date string) ([]model.Food, error) {
	return []model.Food{model.Food{Name: "Test"}}, nil
}

func getDeliciousFood(db *sql.DB, date string) ([]model.DeliciousFood, error) {
	return []model.DeliciousFood{model.DeliciousFood{Name: "Test"}}, nil
}

func (t Today) Message(delicious bool, now time.Time, db *sql.DB) string {
	today := now.Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFood(db, today)
		if err != nil {
			return NoData
		}
		return TodayDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFood(db, today)
	if err != nil {
		return NoData
	}
	return TodayFoodsMessage(lunches)
}

func (t Today) Name() string {
	return t.name
}

func (to Tomorrow) Message(delicious bool, now time.Time, db *sql.DB) string {
	tomorrow := now.AddDate(0, 0, 1).Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFood(db, tomorrow)
		if err != nil {
			return NoData
		}
		return TomorrowDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFood(db, tomorrow)
	if err != nil {
		return NoData
	}
	return TomorrowFoodsMessage(lunches)
}

func (to Tomorrow) Name() string {
	return to.name
}

func (tw ThisWeek) Message(delicious bool, now time.Time, db *sql.DB) string {
	n := nows.New(now)
	beginningOfThisWeek := n.BeginningOfWeek().Format(formatText)
	endOfThisWeek := n.EndOfWeek().Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFoods(db,
			beginningOfThisWeek,
			endOfThisWeek,
		)
		if err != nil {
			return NoData
		}
		return ThisWeekDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFoods(db,
		beginningOfThisWeek,
		endOfThisWeek,
	)
	if err != nil {
		return NoData
	}
	return ThisWeekFoodsMessage(lunches)
}

func (tw ThisWeek) Name() string {
	return tw.name
}

func (nw NextWeek) Message(delicious bool, now time.Time, db *sql.DB) string {
	n := nows.New(now.AddDate(0, 0, 7))
	beginningOfNextWeek := n.BeginningOfWeek().Format(formatText)
	endOfNextWeek := n.EndOfWeek().Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFoods(db,
			beginningOfNextWeek,
			endOfNextWeek,
		)
		if err != nil {
			return NoData
		}
		return NextWeekDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFoods(db,
		beginningOfNextWeek,
		endOfNextWeek,
	)
	if err != nil {
		return NoData
	}
	return NextWeekFoodsMessage(lunches)
}

func (nw NextWeek) Name() string {
	return nw.name
}

func (tm ThisMonth) Message(delicious bool, now time.Time, db *sql.DB) string {
	n := nows.New(now)
	beginningOfThisMonth := n.BeginningOfMonth().Format(formatText)
	endOfThisMonth := n.EndOfMonth().Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFoods(db,
			beginningOfThisMonth,
			endOfThisMonth,
		)
		if err != nil {
			return NoData
		}
		return ThisMonthDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFoods(db,
		beginningOfThisMonth,
		beginningOfThisMonth,
	)
	if err != nil {
		return NoData
	}
	return ThisMonthFoodsMessage(lunches)
}

func (tm ThisMonth) Name() string {
	return tm.name
}

func (nm NextMonth) Message(delicious bool, now time.Time, db *sql.DB) string {
	n := nows.New(now)
	beginningOfNextMonth := n.BeginningOfMonth().Format(formatText)
	endOfNextMonth := n.EndOfMonth().Format(formatText)
	if delicious {
		deliciousFoods, err := getDeliciousFoods(db,
			beginningOfNextMonth,
			endOfNextMonth,
		)
		if err != nil {
			return NoData
		}
		return NextMonthDeliciousFoodsMessage(deliciousFoods)
	}
	lunches, err := getFoods(db,
		beginningOfNextMonth,
		beginningOfNextMonth,
	)
	if err != nil {
		return NoData
	}
	return NextMonthFoodsMessage(lunches)
}

func (nm NextMonth) Name() string {
	return nm.name
}
