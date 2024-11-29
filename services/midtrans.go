package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type MidtransReceiver struct {
	client *coreapi.Client
}

func NewMidtransMoneyReceiver(serverKey string, isSandbox bool) *MidtransReceiver {
	//Initiate client for Midtrans CoreAPI
	var c = coreapi.Client{}
	if isSandbox {
		c.New(serverKey, midtrans.Sandbox)
	} else {
		c.New(serverKey, midtrans.Production)
	}
	return &MidtransReceiver{
		client: &c,
	}
}

func (r *MidtransReceiver) GetNotificationCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// 1. Initialize empty map
		var notificationPayload map[string]interface{}

		// 2. Parse JSON request body and use it to set json to payload
		err := json.NewDecoder(req.Body).Decode(&notificationPayload)
		if err != nil {
			// do something on error when decode
			return
		}

		// 3. Get order-id from payload
		orderId, exists := notificationPayload["order_id"].(string)
		if !exists {
			// do something when key `order_id` not found
			return
		}

		// 4. Check transaction to Midtrans with param orderId
		transactionStatusResp, e := r.client.CheckTransaction(orderId)
		if e != nil {
			http.Error(w, e.GetMessage(), http.StatusInternalServerError)
			return
		} else {
			if transactionStatusResp != nil {
				// 5. Do set transaction status based on response from check transaction status
				if transactionStatusResp.TransactionStatus == "capture" {
					if transactionStatusResp.FraudStatus == "challenge" {
						// TODO set transaction status on your database to 'challenge'
						// e.g: 'Payment status challenged. Please take action on your Merchant Administration Portal
					} else if transactionStatusResp.FraudStatus == "accept" {
						// TODO set transaction status on your database to 'success'
					}
				} else if transactionStatusResp.TransactionStatus == "settlement" {
					// TODO set transaction status on your database to 'success'
				} else if transactionStatusResp.TransactionStatus == "deny" {
					// TODO you can ignore 'deny', because most of the time it allows payment retries
					// and later can become success
				} else if transactionStatusResp.TransactionStatus == "cancel" || transactionStatusResp.TransactionStatus == "expire" {
					// TODO set transaction status on your database to 'failure'
				} else if transactionStatusResp.TransactionStatus == "pending" {
					// TODO set transaction status on your database to 'pending' / waiting payment
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("ok"))
	}
}

func (r *MidtransReceiver) CreateDynamicQRISInvoice(amount int) (*coreapi.ChargeResponse, error) {
	req := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  "odr-" + fmt.Sprint(time.Now().Unix()),
			GrossAmt: int64(amount),
		},
		Qris: &coreapi.QrisDetails{
			Acquirer: "gopay",
		},
	}

	resp, err := r.client.ChargeTransaction(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
