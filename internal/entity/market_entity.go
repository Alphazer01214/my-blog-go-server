package entity

import "time"

type IndexItem struct {
	Code     string `json:"code"`
	QtCode   string `json:"qtcode"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Zxj      string `json:"zxj"`
	Zdf      string `json:"zdf"`
	State    string `json:"state"`
}

type IndexData struct {
	Common  []IndexItem `json:"common"`
	America []IndexItem `json:"america"`
	Europe  []IndexItem `json:"europe"`
	Asia    []IndexItem `json:"asia"`
	Other   []IndexItem `json:"other"`
}

type IndexRawResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data IndexData `json:"data"`
}

type HistoryPoint struct {
	Time    time.Time   `json:"time"`
	Common  []IndexItem `json:"common"`
	America []IndexItem `json:"america"`
	Europe  []IndexItem `json:"europe"`
	Asia    []IndexItem `json:"asia"`
	Other   []IndexItem `json:"other"`
}
