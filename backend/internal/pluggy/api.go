package pluggy

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		return nil, fmt.Errorf("erro ao forçar atualização do item: status %d", resp.StatusCode)
	}

	var res Item
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
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
