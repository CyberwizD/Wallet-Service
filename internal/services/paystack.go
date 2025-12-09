package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PaystackService wraps Paystack HTTP calls.
type PaystackService struct {
	secretKey string
	baseURL   string
	client    *http.Client
}

type paystackInitRequest struct {
	Amount    int64  `json:"amount"`
	Email     string `json:"email"`
	Reference string `json:"reference"`
}

type paystackInitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// NewPaystackService constructs a PaystackService.
func NewPaystackService(secretKey, baseURL string) *PaystackService {
	return &PaystackService{
		secretKey: secretKey,
		// Default base URL is production API.
		baseURL: baseURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// InitializeTransaction requests a Paystack checkout URL.
func (p *PaystackService) InitializeTransaction(amount int64, email, reference string) (string, error) {
	reqBody := paystackInitRequest{
		Amount:    amount,
		Email:     email,
		Reference: reference,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transaction/initialize", p.baseURL), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.secretKey))
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var parsed paystackInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if !parsed.Status {
		return "", fmt.Errorf("paystack init failed: %s", parsed.Message)
	}
	return parsed.Data.AuthorizationURL, nil
}

// VerifySignature checks the Paystack webhook signature using the secret key.
func (p *PaystackService) VerifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}
	mac := hmac.New(sha512.New, []byte(p.secretKey))
	mac.Write(body)
	expected := mac.Sum(nil)
	decoded, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(decoded, expected)
}
