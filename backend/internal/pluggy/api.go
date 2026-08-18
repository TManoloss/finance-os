package pluggy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetAccounts busca todas as contas vinculadas a um item (conexão).
func (c *Client) GetAccounts(itemID string) ([]Account, error) {
	path := fmt.Sprintf("/accounts?itemId=%s", itemID)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar contas: status %d", resp.StatusCode)
	}

	var res struct {
		Results []Account `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Results, nil
}

// GetItem busca detalhes de uma conexão (item).
func (c *Client) GetItem(itemID string) (*Item, error) {
	path := fmt.Sprintf("/items/%s", itemID)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar item: status %d", resp.StatusCode)
	}

	var res Item
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ForceUpdateItem força a atualização de um item na Pluggy.
func (c *Client) ForceUpdateItem(itemID string) (*Item, error) {
	path := fmt.Sprintf("/items/%s", itemID)
	// PATCH /items/{id} aciona uma sincronização manual
	resp, err := c.doRequest(http.MethodPatch, path, []byte("{}"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, responseError("Pluggy recusou a atualização", resp)
	}

	var res Item
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func responseError(operation string, resp *http.Response) error {
	var payload struct {
		Code    string `json:"codeDescription"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	details := make([]string, 0, 2)
	if payload.Code != "" {
		details = append(details, payload.Code)
	}
	if payload.Message != "" {
		details = append(details, payload.Message)
	}
	detail := strings.Join(details, ": ")
	if detail == "" {
		return fmt.Errorf("%s: status %d", operation, resp.StatusCode)
	}
	return fmt.Errorf("%s: status %d (%s)", operation, resp.StatusCode, detail)
}

// DeleteItem revoga o consentimento e remove a conexão na Pluggy.
func (c *Client) DeleteItem(itemID string) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/items/%s", itemID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("erro ao excluir item: status %d", resp.StatusCode)
	}
	return nil
}

// GetConnector busca detalhes de um conector (instituição).
func (c *Client) GetConnector(connectorID int) (*Connector, error) {
	path := fmt.Sprintf("/connectors/%d", connectorID)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar conector: status %d", resp.StatusCode)
	}

	var res Connector
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetTransactions busca TODAS as transações de uma conta com filtros de data,
// percorrendo automaticamente todas as páginas da API da Pluggy.
func (c *Client) GetTransactions(accountID string, from, to string) ([]Transaction, error) {
	var allTransactions []Transaction
	page := 1
	pageSize := 500 // máximo suportado pela Pluggy

	for {
		path := fmt.Sprintf("/transactions?accountId=%s&from=%s&to=%s&page=%d&pageSize=%d",
			accountID, from, to, page, pageSize)

		resp, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("erro ao buscar transações (página %d): status %d", page, resp.StatusCode)
		}

		var res struct {
			Results    []Transaction `json:"results"`
			Total      int           `json:"total"`
			TotalPages int           `json:"totalPages"`
			Page       int           `json:"page"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		allTransactions = append(allTransactions, res.Results...)

		// Sai do loop se estamos na última página
		if page >= res.TotalPages || len(res.Results) == 0 {
			break
		}
		page++
	}

	return allTransactions, nil
}
