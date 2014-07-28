package soramame

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/transform"
	"github.com/PuerkitoBio/goquery"
)

var JST, _ = time.LoadLocation("Asia/Tokyo")

type Observation struct {
	Time time.Time
	PM25 int
	// TODO retrieve other metrics
}

type Result struct {
	Station
	Observations []Observation
}

type Station struct {
	Code      string
	Name      string
	Address   string
	Authority string
	Type      string
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

func fetchObservations(code string, timeString string) ([]Observation, error) {
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
	if err != nil {
		return nil, err
	}
	return observations, nil
}

func fetchHeader(code string, timeString string) (*Station, error) {
	url := fmt.Sprintf(
		"http://soramame.taiki.go.jp/DataListTitle.php?MstCode=%s&Time=%s",
		code,
		timeString,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 5)
	doc.Find(
		"table:first-of-type table:first-child tr td.hyoMenu_List",
	).Each(
		func(i int, td *goquery.Selection) {
			columns[i] = td.Text()
		},
	)

	station := Station{
		Code:      columns[0],
		Name:      columns[1],
		Address:   columns[2],
		Authority: columns[3],
		Type:      columns[4],
	}

	return &station, nil
}

func Fetch(code string) (*Result, error) {
	now := time.Now().In(JST)
	timeString := now.Format("2006010215")

	station, err := fetchHeader(code, timeString)
	if err != nil {
		return nil, err
	}

	observations, err := fetchObservations(code, timeString)
	if err != nil {
		return nil, err
	}

	result := Result{
		Station:      *station,
		Observations: observations,
	}
	return &result, nil
}
