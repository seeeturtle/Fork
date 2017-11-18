package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
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

type Day struct {
	name string
	date string
}

type Today struct {
	name string
}

type Tomorrow struct {
	name string
}

type Nextomorrow struct {
	name string
}

type Threemorrow struct {
	name string
}

type ThisWeek struct {
	name string
}

type WeekAfterNext struct {
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
	Error            = "오류가 났어ㅠㅠ"
	NoData           = "급식은 없어"
	NotText          = "나는 글자 밖에 못 읽어!"
	CannotUnderstand = "뭐라는 거지... 미안, 내가 좀 멍청해."
)

const timeForm string = "20060102"

const LocForm string = "Asia/Seoul"

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

var Scopes = []Scope{
	&Day{name: "날짜"},
	Today{name: "오늘"},
	Tomorrow{name: "내일"},
	Nextomorrow{name: "모레"},
	Threemorrow{name: "글피"},
	WeekAfterNext{name: "다다음주"},
	NextWeek{name: "다음주"},
	ThisWeek{name: "이번주"},
	ThisMonth{name: "이번달"},
	NextMonth{name: "다음달"},
}

var (
	Loc, _ = time.LoadLocation(LocForm)
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

	문의 기능은 아직... 학교에서 문의해줘.
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
	case ok && similar && !delicious:
		text = date + ` 급식을 원하는거야? 그러면 "` + date + ` 급식" 이라고 말해줘.`
	case ok && similar && delicious:
		text = date + ` 맛있는 급식을 원하는거야? 그러면 "` + date + ` 맛있는 급식" 이라고 말해줘.`
	case ok && (date != ""):
		text = getResponseText(date, delicious)
	case ok && (date == ""):
		text = "언제 급식을 원하는 거야?"
	default:
		text = CannotUnderstand
	}
	response["message"] = make(map[string]string)
	response["message"]["text"] = text
	respondJSON(w, http.StatusOK, response)
}

func parseContent(str string) (ok, delicious, similar, slang bool, date string) {
	splitted := strings.Split(str, " ")
	re := regexp.MustCompile(`[\d]+월[\d]+일`)
	for _, w := range splitted {
		if w == "" {
			continue
		}
		d := re.FindString(w)
		switch {
		case d != "":
			if date == "" {
				t, _ := time.Parse("2006년1월2일", time.Now().In(Loc).Format("2006년")+d)
				date = "날짜" + t.Format(timeForm)
			}
		case strings.Contains(w, "오늘"):
			if date == "" {
				date = "오늘"
			}
		case strings.Contains(w, "내일"):
			if date == "" {
				date = "내일"
			}
		case strings.Contains(w, "모레"):
			if date == "" {
				date = "모레"
			}
		case strings.Contains(w, "글피"):
			if date == "" {
				date = "글피"
			}
		case strings.Contains(w, "이번주"):
			if date == "" {
				date = "이번주"
			}
		case strings.Contains(w, "다다음주"):
			if date == "" {
				date = "다다음주"
			}
		case strings.Contains(w, "다음주"):
			if date == "" {
				date = "다음주"
			}
		case strings.Contains(w, "이번달"):
			if date == "" {
				date = "이번달"
			}
		case strings.Contains(w, "다음달"):
			if date == "" {
				date = "다음달"
			}
		case similarity([]rune("오늘"), []rune(w)) >= 0.5:
			if date == "" {
				date = "오늘"
				similar = true
			}
		case similarity([]rune("내일"), []rune(w)) >= 0.5:
			if date == "" {
				date = "내일"
				similar = true
			}
		case similarity([]rune("모레"), []rune(w)) >= 0.42:
			if date == "" {
				date = "모레"
				similar = true
			}
		case similarity([]rune("글피"), []rune(w)) >= 0.5:
			if date == "" {
				date = "글피"
				similar = true
			}
		case similarity([]rune("이번주"), []rune(w)) >= 0.5:
			if date == "" {
				date = "이번주"
				similar = true
			}
		case similarity([]rune("다음주"), []rune(w)) >= 0.5:
			if date == "" {
				date = "다음주"
				similar = true
			}
		case similarity([]rune("다다음주"), []rune(w)) >= 0.5:
			if date == "" {
				date = "다다음주"
				similar = true
			}
		case similarity([]rune("이번달"), []rune(w)) >= 0.5:
			if date == "" {
				date = "이번달"
				similar = true
			}
		case similarity([]rune("다음달"), []rune(w)) >= 0.5:
			if date == "" {
				date = "다음달"
				similar = true
			}
		}
		switch {
		case strings.Contains(w, "급식"):
			ok = true
		case strings.Contains(w, "점심"):
			ok = true
		}
		if strings.Contains(w, "맛있") {
			delicious = true
		}
		if slangSimilarity(w) >= 0.38 {
			slang = true
		}
	}
	return
}

func similarity(a, b []rune) float64 {
	intersection := make([]int, 0)
	union := make([]int, 0)
	var longer, shorter []int
	sliceA := make([]int, 0)
	for _, r := range a {
		sliceA = append(sliceA, seperate(r)...)
	}
	sliceA = cutByTwo(sliceA)
	sliceB := make([]int, 0)
	for _, r := range b {
		sliceB = append(sliceB, seperate(r)...)
	}
	sliceB = cutByTwo(sliceB)
	if len(sliceA) >= len(sliceB) {
		longer = sliceA
		shorter = sliceB
	} else {
		longer = sliceB
		shorter = sliceA
	}
	for _, i := range shorter {
		if !inIntSlice(i, union) {
			union = append(union, i)
		}
		for _, j := range longer {
			if !inIntSlice(j, union) {
				union = append(union, j)
			}
			if i == j && !inIntSlice(i, intersection) {
				intersection = append(intersection, i)
			}
		}
	}
	intersectionLen := len(intersection)
	unionLen := len(union)
	return float64(intersectionLen) / float64(unionLen)
}

func inIntSlice(a int, b []int) bool {
	for _, i := range b {
		if a == i {
			return true
		}
	}
	return false
}

func seperate(a rune) []int {
	var slice []int
	code := int(a) - 44032
	jongSeong := code % 28
	jungSeong := ((code - jongSeong) / 28) % 21
	choSeong := ((code-jongSeong)/28 - jungSeong) / 21
	slice = append(slice, choSeong)
	slice = append(slice, jungSeong)
	if jongSeong != 0 {
		slice = append(slice, jongSeong)
	}

	return slice
}

func cutByTwo(a []int) []int {
	var result []int
	result = append(result, (a[0]+3)*30)
	for index, i := range a {
		if index == len(a)-1 {
			result = append(result, 100*(i+3))
			continue
		}
		result = append(result, (i+3)*a[index+1])
	}
	return result
}

func slangSimilarity(str string) float64 {
	var preprocessed string
	var similaritys []float64
	re := regexp.MustCompile("[^가-힣]")
	preprocessed = re.ReplaceAllString(str, "")
	for _, s := range slangs {
		similaritys = append(similaritys, similarity([]rune(s), []rune(preprocessed)))
	}
	var max float64
	for _, s := range similaritys {
		if s >= max {
			max = s
		}
	}
	return max
}

func getResponseText(scope string, delicious bool) string {
	for _, s := range Scopes {
		if strings.Contains(scope, s.Name()) {
			if s.Name() == "날짜" {
				s.(*Day).date = string([]rune(scope)[2:])
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

func JoinWithComma(lunches []model.Lunch) string {
	var str string
	if len(lunches) > 1 {
		for i, lunch := range lunches {
			names := []string{}
			for _, food := range lunch.Foods {
				names = append(names, food.Name)
			}
			dateStr := getDatesStr(lunch.Date)
			str += dateStr + strings.Join(names, ", ") + getPostposition(names[len(names)-1])
			if i != len(lunches)-1 {
				str += ",\n"
			} else {
				str += " "
			}
		}
	} else if len(lunches) == 1 {
		names := []string{}
		for _, food := range lunches[0].Foods {
			names = append(names, food.Name)
		}
		dateStr := getDateStr(lunches[0].Date)
		str = dateStr + strings.Join(names, ", ") + getPostposition(names[len(names)-1])
	}
	return str
}

func getDatesStr(date string) string {
	var weekDays [7]string = [7]string{
		"일요일",
		"월요일",
		"화요일",
		"수요일",
		"목요일",
		"금요일",
		"토요일",
	}
	dateTime, _ := time.Parse(timeForm, date)
	dateTime = roundTime(dateTime.In(Loc))
	n := time.Now().In(Loc)
	diffWeeks := int(now.New(dateTime).BeginningOfWeek().Sub(now.New(n).BeginningOfWeek())) / (int(time.Hour) * 24 * 7)
	weekDay := weekDays[int(dateTime.Weekday())]
	switch {
	case diffWeeks == -3:
		return "3주전 " + weekDay + "은 "
	case diffWeeks == -2:
		return "저저번주 " + weekDay + "은 "
	case diffWeeks == -1:
		return "저번주 " + weekDay + "은 "
	case diffWeeks == 0:
		return "이번주 " + weekDay + "은"
	case diffWeeks == 1:
		return "다음주 " + weekDay + "은 "
	case diffWeeks == 2:
		return "다다음주 " + weekDay + "은 "
	case diffWeeks == 3:
		return "3주후 " + weekDay + "은 "
	default:
		return fmt.Sprintf("%d월 %d일은 ", dateTime.Month(), dateTime.Day())
	}
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
		if len(deliciousLunches) == 0 {
			switch s.(type) {
			case *Day:
				dateTime, _ := time.Parse(timeForm, s.(*Day).date)
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
			dateTime, _ := time.Parse(timeForm, s.(*Day).date)
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

func (d *Day) Beginning() string {
	return d.date
}

func (d *Day) End() string {
	return d.date
}

func (d Day) Name() string {
	return d.name
}

func (d Day) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다"
}

func (d Day) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!"
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
	return f + "나온다"
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
	return f + "나온다"
}

func (to Tomorrow) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!!"
}

func (nt Nextomorrow) Beginning() string {
	return time.Now().AddDate(0, 0, 2).In(Loc).Format(timeForm)
}

func (nt Nextomorrow) End() string {
	return time.Now().AddDate(0, 0, 2).In(Loc).Format(timeForm)
}

func (nt Nextomorrow) Name() string {
	return nt.name
}

func (nt Nextomorrow) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다"
}

func (nt Nextomorrow) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!"
}

func (tm Threemorrow) Beginning() string {
	return time.Now().AddDate(0, 0, 3).In(Loc).Format(timeForm)
}

func (tm Threemorrow) End() string {
	return time.Now().AddDate(0, 0, 3).In(Loc).Format(timeForm)
}

func (tm Threemorrow) Name() string {
	return tm.name
}

func (tm Threemorrow) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다"
}

func (tm Threemorrow) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!"
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
	return f + "나온다. 이번주에는 그럭저럭 하네."
}

func (tw ThisWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!! 이번주에 맛있는게 많다!"
}

func (wn WeekAfterNext) Beginning() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 0, 14))
	return n.BeginningOfWeek().In(Loc).Format(timeForm)
}

func (wn WeekAfterNext) End() string {
	n := now.New(time.Now().In(Loc).AddDate(0, 0, 14))
	return n.EndOfWeek().In(Loc).Format(timeForm)
}

func (wn WeekAfterNext) Name() string {
	return wn.name
}

func (wn WeekAfterNext) FoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다"
}

func (wn WeekAfterNext) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!"
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
	return f + "나온다. 괜찮은데?"
}

func (nm NextWeek) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다!!!!"
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
	return f + "나온다.\n이번달 급식임"
}

func (tm ThisMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다.\n기억해둬"
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
	return f + "나온다.\n다음달 급식임."
}

func (nm NextMonth) DeliciousFoodMessage(lunches []model.Lunch) string {
	f := JoinWithComma(lunches)
	return f + "나온다.\n다음달 급식도 괜찮은 듯."
}
