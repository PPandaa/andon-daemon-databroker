package model

import "time"

type IfpcfgInlline struct {
	Group struct {
		Items []struct {
			ID   string `json:"id"`
			Item struct {
				ID          string    `json:"_id"`
				Parent      string    `json:"parent"`
				Depth       int       `json:"depth"`
				Name        string    `json:"name"`
				Description string    `json:"description"`
				TimeZone    string    `json:"timeZone"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
			} `json:"item,omitempty"`
		} `json:"items"`
		TotalCount int `json:"totalCount"`
	} `json:"group"`
	Machine struct {
		Items []struct {
			ID   string `json:"id"`
			Item struct {
				ID          string    `json:"_id"`
				Group       string    `json:"group"`
				Name        string    `json:"name"`
				Description string    `json:"description"`
				Index       int       `json:"index"`
				ImageURL    string    `json:"imageUrl"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
			} `json:"item"`
		} `json:"items"`
		TotalCount int `json:"totalCount"`
	} `json:"machine"`
	Wacfg struct {
		D struct {
			Ifp struct {
				UTg struct {
					ALKBAB2ELFREQ5F23Baac16004D00071918F3 struct {
						TID int `json:"TID"`
					} `json:"ALKB_AB2EL:FREQ_5f23baac16004d00071918f3"`
					ALKBAB2ELFREQ5F23Bad816004D00071918F4 struct {
						TID int `json:"TID"`
					} `json:"ALKB_AB2EL:FREQ_5f23bad816004d00071918f4"`
					ALKBAB2LVTN5F3C79Aca57B360006A5E513 struct {
						TID int `json:"TID"`
					} `json:"ALKB_AB2L:VTN_5f3c79aca57b360006a5e513"`
					ALKBAB2LVTN5F3C79D2A57B360006A5E515 struct {
						TID int `json:"TID"`
					} `json:"ALKB_AB2L:VTN_5f3c79d2a57b360006a5e515"`
					ALKBAB2LFREQ5F4741Dc8772726794E29C70 struct {
						TID int `json:"TID"`
					} `json:"ALKB_AB2L:FREQ_5f4741dc8772726794e29c70"`
				} `json:"UTg"`
			} `json:"ifp"`
		} `json:"d"`
		Ts time.Time `json:"ts"`
	} `json:"wacfg"`
	Parameter struct {
		Items []struct {
			ID   string `json:"id"`
			Item struct {
				ID          string    `json:"_id"`
				Machine     string    `json:"machine"`
				Name        string    `json:"name"`
				Description string    `json:"description"`
				TagID       string    `json:"tagId"`
				ValueType   string    `json:"valueType"`
				CreatedAt   time.Time `json:"createdAt"`
				UpdatedAt   time.Time `json:"updatedAt"`
			} `json:"item"`
		} `json:"items"`
		TotalCount int `json:"totalCount"`
	} `json:"parameter"`
	Ts time.Time `json:"ts"`
}
