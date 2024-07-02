package sn

import (
	"encoding/json"
	"fmt"
	"time"
)

type Invoice struct {
	Id                int                    `json:"id,string"`
	Hash              string                 `json:"hash"`
	Hmac              string                 `json:"hmac"`
	Bolt11            string                 `json:"bolt11"`
	SatsRequested     int                    `json:"satsRequested"`
	SatsReceived      int                    `json:"satsReceived"`
	Cancelled         bool                   `json:"cancelled"`
	ConfirmedAt       time.Time              `json:"createdAt"`
	ExpiresAt         time.Time              `json:"expiresAt"`
	Nostr             map[string]interface{} `json:"nostr"`
	IsHeld            bool                   `json:"isHeld"`
	Comment           string                 `json:"comment"`
	Lud18Data         map[string]interface{} `json:"lud18Data"`
	ConfirmedPreimage string                 `json:"confirmedPreimage"`
	ActionState       string                 `json:"actionState"`
	ActionType        string                 `json:"actionType"`
}

type PaymentMethod string

const (
	PaymentMethodFeeCredits  PaymentMethod = "FEE_CREDIT"
	PaymentMethodOptimistic  PaymentMethod = "OPTIMISTIC"
	PaymentMethodPessimistic PaymentMethod = "PESSIMISTIC"
)

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
