package adapters

import (
    "encoding/json"
    "errors"
    "net/http"
)

type BankAPIClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewBankAPIClient(baseURL string) *BankAPIClient {
    return &BankAPIClient{
        baseURL:    baseURL,
        httpClient: &http.Client{},
    }
}

func (client *BankAPIClient) FetchBankStatements(accountID string) ([]BankStatement, error) {
    req, err := http.NewRequest("GET", client.baseURL+"/statements?account_id="+accountID, nil)
    if err != nil {
        return nil, err
    }

    resp, err := client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("failed to fetch bank statements: " + resp.Status)
    }

    var statements []BankStatement
    if err := json.NewDecoder(resp.Body).Decode(&statements); err != nil {
        return nil, err
    }

    return statements, nil
}

type BankStatement struct {
    ID        string  `json:"id"`
    Amount    float64 `json:"amount"`
    Date      string  `json:"date"`
    TransactionType string `json:"transaction_type"`
}