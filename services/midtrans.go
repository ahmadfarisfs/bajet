package services

import (
	"fmt"
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

func (r *MidtransReceiver) HandleNotificationCallback()

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
