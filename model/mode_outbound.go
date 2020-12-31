package model

import "time"

type Ifpcfg struct {
	Group     Group     `json:"group"`
	Machine   Machine   `json:"machine"`
	Parameter Parameter `json:"parameter"`
	Ts        time.Time `json:"ts"`
}

//put tag "-" if you dont want that value
type Group struct {
	Items []struct {
		ID   string `json:"id"`
		Item struct {
			ID          string    `json:"_id"`
			Parent      string    `json:"parent"`
			Depth       int       `json:"depth"`
			Name        string    `json:"name"`
			Description string    `json:"-"`
			TimeZone    string    `json:"timeZone"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
		} `json:"item,omitempty"`
	} `json:"items"`
	TotalCount int `json:"totalCount"`
}

type Machine struct {
	Items []struct {
		ID   string `json:"id"`
		Item struct {
			ID          string    `json:"_id"`
			Group       string    `json:"group"` //值是groupId
			Name        string    `json:"name"`
			Description string    `json:"-"`
			Index       int       `json:"index"`
			ImageURL    string    `json:"-"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
		} `json:"item"`
	} `json:"items"`
	TotalCount int `json:"totalCount"`
}

type Parameter struct {
	Items []struct {
		ID   string `json:"id"`
		Item struct {
			ID          string    `json:"_id"`
			Machine     string    `json:"machine"`
			Name        string    `json:"name"`
			Description string    `json:"-"`
			TagID       string    `json:"tagId"`
			ValueType   string    `json:"valueType"`
			CreatedAt   time.Time `json:"createdAt"`
			UpdatedAt   time.Time `json:"updatedAt"`
		} `json:"item"`
	} `json:"items"`
	TotalCount int `json:"totalCount"`
}
