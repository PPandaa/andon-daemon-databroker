package model

import (
	"strings"
	"time"

	"databroker/pkg/desk"
)

type Ifpcfg struct {
	Group     Group     `json:"group,omitempty"`
	Machine   Machine   `json:"machine,omitempty"`
	Parameter Parameter `json:"parameter,omitempty"`
	Ts        time.Time `json:"ts,omitempty"`
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
	TotalCount float64 `json:"totalCount"`
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
	TotalCount float64 `json:"totalCount"`
}

type Parameter struct {
	Items []*struct {
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
	} `json:"items,omitempty"`
	TotalCount float64 `json:"totalCount,omitempty"`
}

type Wadata struct {
	ID string `json:"_id"`
	D  struct {
		Ifp struct {
			Val interface{} `json:"Val"`
		} `json:"ifp"`
	} `json:"d"`
	Ts time.Time `json:"ts"`
}

func (w *Wadata) Service(key string) {
	if val := w.D.Ifp.Val; val != nil {
		m := val.(map[string]interface{})
		for k, _ := range m {
			// fmt.Println(k, ":", v)
			ss := strings.Split(k, "_")
			if ss[0] == key {
				// fmt.Println("send status id and value[", ss[1], v, "] to peter func")
				desk.UpdateMachineStatus(ss[1])
			}
		}
	}
}
