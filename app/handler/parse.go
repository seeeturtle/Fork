package handler

import (
	"regexp"
	"strings"
	"time"
)

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
				t, _ := time.Parse("2006년1월2일", time.Now().In(loc).Format("2006년")+d)
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
