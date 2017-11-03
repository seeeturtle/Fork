package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	nows "github.com/jinzhu/now"
	"github.com/joshua1b/Fork/app/model"
)

type Scope interface {
	Beginning(time.Time) string
	End(time.Time) string
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

type Response struct {
	Text string `json:"text"`
}

const (
	NoData           = "아직 데이터가 없어..."
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

var Expected = []string{"맛있", "급식", "점심"}

var Loc, _ = time.LoadLocation(LocForm)

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
			response.Text = getResponseText(date, true)
		} else {
			response.Text = getResponseText(date, false)
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

func getResponseText(scope string, delicious bool) string {
	now := time.Now().In(Loc)
	for _, s := range Scopes {
		if scope == s.Name() {
			return Message(s, delicious, now)
		}
	}
	return CannotUnderstand
}

func JoinWithComma(lunches []model.Lunch) string {
	var str string
	for _, lunch := range lunches {
		names := make([]string, len(lunch.Foods))
		for _, food := range lunch.Foods {
			names = append(names, food.Name)
		}
		str += lunch.Date + "\n" + strings.Join(names, "\n") + "\n"
	}
	return str
}

func Message(s Scope, delicious bool, now time.Time) string {
	beginning := s.Beginning(now)
	end := s.End(now)
	if delicious {
		deliciousLunches, err := model.Lunches.GetDelicious(beginning, end)
		if err != nil {
			return NoData
		}
		return s.DeliciousFoodMessage(deliciousLunches)
	}
	lunches, err := model.Lunches.Get(beginning, end)
	if err != nil {
		return NoData
	}
	return s.FoodMessage(lunches)
}

func (t Today) Beginning(now time.Time) string {
	return now.Format(timeForm)
}

func (t Today) End(now time.Time) string {
	return now.Format(timeForm)
}

func (t Today) Name() string {
	return t.name
}

func (t Today) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "오늘은 이런게 나온당"
}

func (t Today) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "친구들 출격!!!!"
}

func (to Tomorrow) Beginning(now time.Time) string {
	return now.AddDate(0, 0, 1).Format(timeForm)
}

func (to Tomorrow) End(now time.Time) string {
	return now.AddDate(0, 0, 1).Format(timeForm)
}

func (to Tomorrow) Name() string {
	return to.name
}

func (to Tomorrow) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "내일 이런 거 먹을 수 있넹"
}

func (to Tomorrow) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "내일은 어쩔 수 없이 학교를 가야겠네"
}

func (tw ThisWeek) Beginning(now time.Time) string {
	n := nows.New(now)
	return n.BeginningOfWeek().In(Loc).Format(timeForm)
}

func (tw ThisWeek) End(now time.Time) string {
	n := nows.New(now)
	return n.EndOfWeek().In(Loc).Format(timeForm)
}

func (tw ThisWeek) Name() string {
	return tw.name
}

func (tw ThisWeek) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "이번주에는 이런게 나온당"
}

func (tw ThisWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "우워어ㅓㅓ 출격!!!"
}

func (nw NextWeek) Beginning(now time.Time) string {
	n := nows.New(now.AddDate(0, 0, 7))
	return n.BeginningOfWeek().In(Loc).Format(timeForm)
}

func (nw NextWeek) End(now time.Time) string {
	n := nows.New(now.AddDate(0, 0, 7))
	return n.EndOfWeek().In(Loc).Format(timeForm)
}

func (nw NextWeek) Name() string {
	return nw.name
}

func (nw NextWeek) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "다음주는 이런게 나온당..."
}

func (nm NextWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "와아아ㅏㅏ아 역시 급식이 최고야!"
}

func (tm ThisMonth) Beginning(now time.Time) string {
	n := nows.New(now)
	return n.BeginningOfMonth().In(Loc).Format(timeForm)
}

func (tm ThisMonth) End(now time.Time) string {
	n := nows.New(now)
	return n.EndOfMonth().In(Loc).Format(timeForm)
}

func (tm ThisMonth) Name() string {
	return tm.name
}

func (tm ThisMonth) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "이번달에는 이런게 있넹"
}

func (tm ThisMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "이번달에는 이런게 있네!!! 가자!!!"
}

func (nm NextMonth) Beginning(now time.Time) string {
	n := nows.New(now.AddDate(0, 1, 0))
	return n.BeginningOfMonth().In(Loc).Format(timeForm)
}

func (nm NextMonth) End(now time.Time) string {
	n := nows.New(now.AddDate(0, 1, 0))
	return n.EndOfMonth().In(Loc).Format(timeForm)
}

func (nm NextMonth) Name() string {
	return nm.name
}

func (nm NextMonth) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "다음달에는 이런게 있넹"
}

func (nm NextMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "다음달에는 이런게 있네!!! 다음달은 도대체 언제 오는 거야!!!!"
}
