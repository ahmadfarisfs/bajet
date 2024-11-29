package services

import (
	"os"
	"testing"
)

func TestMidtransQRIS(t *testing.T) {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		t.Fatal("MIDTRANS_SERVER_KEY environment variable is not set")
	}
	// Create a new MidtransReceiver instance
	receiver := NewMidtransMoneyReceiver(serverKey, true)

	// Create a new QRIS invoice
	resp, err := receiver.CreateDynamicQRISInvoice(10000)
	if err != nil {
		t.Fatalf("Error creating QRIS invoice: %v", err)
	}

	// Check if the response is not nil
	if resp == nil {
		t.Fatalf("Response is nil")
	}
	t.Logf("QRIS invoice created: %v", resp)
}
