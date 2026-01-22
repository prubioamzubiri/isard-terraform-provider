package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Group representa un grupo en Isard VDI
type Group struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ParentCategory string   `json:"parent_category"`
	LinkedGroups   []string `json:"linked_groups"`
	Enrollment     map[string]interface{} `json:"enrollment,omitempty"`
}

// GetGroups obtiene la lista de grupos
func (c *Client) GetGroups() ([]Group, error) {
	reqURL := fmt.Sprintf("https://%s/api/v3/admin/groups", c.HostURL)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando petición GET: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando GET: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error obteniendo grupos (status %d): %s", res.StatusCode, string(body))
	}

	var groups []Group
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	return groups, nil
}
