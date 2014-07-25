package soramame

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var JST, _ = time.LoadLocation("Asia/Tokyo")

type Observation struct {
	Time time.Time
	PM25 int
}

type Result struct {
	Code         string
	Observations []Observation
}

func parseRow(row *goquery.Selection) (*Observation, error) {
	columns := make([]string, 20)
	row.Find("td").Each(
		func(i int, selection *goquery.Selection) {
			columns[i] = selection.Text()
		},
	)
	y, err := strconv.Atoi(columns[0])
	if err != nil {
		return nil, err
	}
	m, err := strconv.Atoi(columns[1])
	if err != nil {
		return nil, err
	}
	d, err := strconv.Atoi(columns[2])
	if err != nil {
		return nil, err
	}
	h, err := strconv.Atoi(columns[3])
	if err != nil {
		return nil, err
	}
	observedTime := time.Date(y, time.Month(m), d, h, 0, 0, 0, JST)

	pm25, err := strconv.Atoi(columns[14])
	if err != nil {
		return nil, err
	}

	observation := Observation{
		Time: observedTime,
		PM25: pm25,
	}

	return &observation, nil
}

func Fetch(code string) (*Result, error) {
	now := time.Now().In(JST)
	timeString := now.Format("2006010215")
	url := fmt.Sprintf(
		"http://soramame.taiki.go.jp/DataListHyou.php?MstCode=%s&Time=%s",
		code,
		timeString,
	)

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	observations := make([]Observation, 0)
	doc.Find("table.hyoMenu tr").EachWithBreak(
		func(_ int, row *goquery.Selection) bool {
			observation, err := parseRow(row)
			if err != nil {
				return false
			}
			observations = append(observations, *observation)
			return true
		},
	)

	result := Result{
		Code:         code,
		Observations: observations,
	}
	return &result, nil
}
