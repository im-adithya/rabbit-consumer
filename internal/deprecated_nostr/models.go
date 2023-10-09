package nostr

import "time"

type Invoice struct {
	ID                       int64             `json:"id"`
	Type                     string            `json:"type"`
	UserLogin                string            `json:"user_login"`
	Amount                   int64             `json:"amount"`
	Fee                      int64             `json:"fee"`
	Balance                  int64             `json:"balance"`
	Memo                     string            `json:"memo"`
	DescriptionHash          string            `json:"description_hash,omitempty"`
	PaymentRequest           string            `json:"payment_request"`
	DestinationPubkeyHex     string            `json:"destination_pubkey_hex"`
	DestinationCustomRecords map[uint64][]byte `json:"custom_records,omitempty"`
	RHash                    string            `json:"r_hash"`
	Preimage                 string            `json:"preimage"`
	Keysend                  bool              `json:"keysend"`
	State                    string            `json:"state"`
	ErrorMessage             string            `json:"error_message,omitempty"`
	CreatedAt                time.Time         `json:"created_at"`
	ExpiresAt                time.Time         `json:"expires_at"`
	UpdatedAt                time.Time         `json:"updated_at"`
	SettledAt                time.Time         `json:"settled_at"`
}
