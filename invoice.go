package sn

import (
	"encoding/json"
	"fmt"
	"time"
)

type Invoice struct {
	Id                int       `json:"id,string"`
	Hash              string    `json:"hash"`
	Hmac              string    `json:"hmac"`
	Bolt11            string    `json:"bolt11"`
	CreatedAt         time.Time `json:"createdAt"`
	ExpiresAt         time.Time `json:"expiresAt"`
	Cancelled         bool      `json:"cancelled"`
	ConfirmedAt       time.Time `json:"confirmedAt"`
	SatsReceived      int       `json:"satsReceived"`
	SatsRequested     int       `json:"satsRequested"`
	Comment           string    `json:"nostr"`
	IsHeld            int       `json:"isHeld"`
	ConfirmedPreimage string    `json:"confirmedPreimage"`
}

type CreateInvoiceArgs struct {
	Amount      int
	ExpireSecs  int
	HodlInvoice bool
}

type CreateInvoiceResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		CreateInvoice Invoice `json:"createInvoice"`
	} `json:"data"`
}

func (c *Client) CreateInvoice(args *CreateInvoiceArgs) (*Invoice, error) {
	if args == nil {
		args = &CreateInvoiceArgs{}
	}

	body := GqlBody{
		// TODO: add createdAt
		//   when I wrote this code, createdAt returned null but is non-nullable
		//   so I had to remove it.
		Query: `
		mutation createInvoice($amount: Int!, $expireSecs: Int, $hodlInvoice: Boolean) {
			createInvoice(amount: $amount, expireSecs: $expireSecs, hodlInvoice: $hodlInvoice) {
				id
				hash
				hmac
				bolt11
				satsRequested
				satsReceived
				isHeld
				comment
				confirmedPreimage
				expiresAt
				confirmedAt
			}
		}`,
		Variables: map[string]interface{}{
			"amount": args.Amount,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody CreateInvoiceResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding items: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.CreateInvoice, nil
}
