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

// GetTransactions busca transações de uma conta com filtros de data.
func (c *Client) GetTransactions(accountID string, from, to string) ([]Transaction, error) {
	path := fmt.Sprintf("/transactions?accountId=%s&from=%s&to=%s", accountID, from, to)
	
	// A Pluggy usa paginação, para esta micro-tarefa inicial vamos focar na primeira página
	// Nas próximas evoluções podemos implementar o scroll completo
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar transações: status %d", resp.StatusCode)
	}

	var res struct {
		Results []Transaction `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Results, nil
}
