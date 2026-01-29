package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Deployment representa la estructura de un deployment en la API
type Deployment struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	TemplateID    string                 `json:"template_id"`
	DesktopName   string                 `json:"desktop_name"`
	Visible       bool                   `json:"visible"`
	Allowed       map[string]interface{} `json:"allowed"`
	VCPUs         int64                  `json:"vcpus,omitempty"`
	Memory        float64                `json:"memory,omitempty"`
	Interfaces    []string               `json:"interfaces,omitempty"`
	GuestProps    map[string]interface{} `json:"guest_properties,omitempty"`
	Image         map[string]interface{} `json:"image,omitempty"`
	UserPerms     []string               `json:"user_permissions,omitempty"`
}

// DeploymentInfo representa la información completa de un deployment
type DeploymentInfo struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DesktopName     string                 `json:"desktop_name"`
	Visible         bool                   `json:"visible"`
	TemplateID      string                 `json:"template"`
	Allowed         map[string]interface{} `json:"allowed"`
	TotalDesktops   int                    `json:"totalDesktops"`
	VisibleDesktops int                    `json:"visibleDesktops"`
	StartedDesktops int                    `json:"startedDesktops"`
	CreatingDesktops int                   `json:"creatingDesktops"`
}

// CreateDeployment crea un nuevo deployment
func (c *Client) CreateDeployment(
	name string,
	description string,
	templateID string,
	desktopName string,
	visible bool,
	allowed map[string]interface{},
	vcpus *int64,
	memory *float64,
	interfaces []string,
	guestProperties map[string]interface{},
	image map[string]interface{},
	userPermissions []string,
) (string, error) {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployments", c.HostURL)

	// Obtener información del template para construir payload completo
	template, err := c.GetTemplateInfo(templateID)
	if err != nil {
		return "", fmt.Errorf("error obteniendo información del template: %w", err)
	}

	templateHardware, _ := template["hardware"].(map[string]interface{})
	templateGuestProps, _ := template["guest_properties"].(map[string]interface{})
	templateImage, _ := template["image"].(map[string]interface{})

	// Construir el payload básico
	payload := map[string]interface{}{
		"name":         name,
		"description":  description,
		"template_id":  templateID,
		"desktop_name": desktopName,
		"visible":      visible,
		"allowed":      allowed,
	}

	// Agregar user_permissions
	if len(userPermissions) > 0 {
		payload["user_permissions"] = userPermissions
	} else {
		payload["user_permissions"] = []string{}
	}

	// Construir hardware desde el template y aplicar override si se especifica
	hardware := make(map[string]interface{})
	
	// Copiar campos del template
	if boot_order, ok := templateHardware["boot_order"]; ok {
		hardware["boot_order"] = boot_order
	}
	if disk_bus, ok := templateHardware["disk_bus"]; ok {
		hardware["disk_bus"] = disk_bus
	}
	if disks, ok := templateHardware["disks"]; ok {
		hardware["disks"] = disks
	}
	if floppies, ok := templateHardware["floppies"]; ok {
		hardware["floppies"] = floppies
	}
	if isos, ok := templateHardware["isos"]; ok {
		hardware["isos"] = isos
	}
	
	// Siempre incluir videos - requerido por la API
	if videos, ok := templateHardware["videos"]; ok {
		hardware["videos"] = videos
	} else if video, ok := templateHardware["video"]; ok {
		hardware["videos"] = video
	} else {
		hardware["videos"] = []string{"default"}
	}

	// vcpus y memory: usar valores especificados o del template
	if vcpus != nil {
		hardware["vcpus"] = int(*vcpus)
	} else if templateVCPUs, ok := templateHardware["vcpus"].(float64); ok {
		hardware["vcpus"] = int(templateVCPUs)
	} else {
		hardware["vcpus"] = 2
	}
	
	if memory != nil {
		hardware["memory"] = int(*memory)
	} else if templateMemory, ok := templateHardware["memory"].(float64); ok {
		// Convertir de KiB a GB
		hardware["memory"] = int(templateMemory / 1024 / 1024)
	} else {
		hardware["memory"] = 2
	}
	
	// interfaces: usar valores especificados o del template
	if len(interfaces) > 0 {
		hardware["interfaces"] = interfaces
	} else {
		hardware["interfaces"] = []string{"default", "wireguard"}
	}
	
	// reservables: siempre ["None"] para que funcione correctamente
	hardware["reservables"] = map[string]interface{}{"vgpus": []string{"None"}}
	payload["hardware"] = hardware
	
	// guest_properties: combinar valores del template con los especificados
	finalGuestProps := make(map[string]interface{})
	
	// Primero copiar del template si existe
	if templateGuestProps != nil {
		for k, v := range templateGuestProps {
			finalGuestProps[k] = v
		}
	}
	
	// Luego sobrescribir/añadir con valores especificados
	if len(guestProperties) > 0 {
		for k, v := range guestProperties {
			finalGuestProps[k] = v
		}
	}
	
	payload["guest_properties"] = finalGuestProps

	// image: usar valores especificados o del template
	if len(image) > 0 {
		payload["image"] = image
	} else if templateImage != nil {
		payload["image"] = templateImage
	} else {
		payload["image"] = map[string]interface{}{"type": "user"}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error codificando JSON: %w", err)
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creando la petición POST: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error ejecutando POST: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("error creando deployment (status %d): %s", res.StatusCode, string(body))
	}

	// Parsear la respuesta para obtener el ID
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error parseando respuesta JSON: %w", err)
	}

	deploymentID, ok := response["id"].(string)
	if !ok {
		return "", fmt.Errorf("no se encontró el ID en la respuesta: %s", string(body))
	}

	return deploymentID, nil
}

// GetDeployment obtiene la información de un deployment
func (c *Client) GetDeployment(deploymentID string) (*DeploymentInfo, error) {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployment/%s", c.HostURL, deploymentID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando la petición GET: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando GET: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("deployment not found")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error obteniendo deployment (status %d): %s", res.StatusCode, string(body))
	}

	// Parsear la respuesta
	var deployment DeploymentInfo
	if err := json.Unmarshal(body, &deployment); err != nil {
		return nil, fmt.Errorf("error parseando respuesta JSON: %w", err)
	}

	return &deployment, nil
}

// GetDeploymentInfo obtiene información detallada de un deployment para edición
func (c *Client) GetDeploymentInfo(deploymentID string) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployment/info/%s", c.HostURL, deploymentID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando la petición GET: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando GET: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("deployment not found")
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error obteniendo deployment info (status %d): %s", res.StatusCode, string(body))
	}

	// Parsear la respuesta
	var deploymentInfo map[string]interface{}
	if err := json.Unmarshal(body, &deploymentInfo); err != nil {
		return nil, fmt.Errorf("error parseando respuesta JSON: %w", err)
	}

	return deploymentInfo, nil
}

// UpdateDeployment actualiza un deployment existente
func (c *Client) UpdateDeployment(deploymentID string, updateData map[string]interface{}) error {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployment/%s", c.HostURL, deploymentID)

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("error codificando JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando la petición PUT: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando PUT: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error actualizando deployment (status %d): %s", res.StatusCode, string(body))
	}

	return nil
}

// DeleteDeployment elimina un deployment
func (c *Client) DeleteDeployment(deploymentID string, permanent bool) error {
	permanentStr := "false"
	if permanent {
		permanentStr = "true"
	}
	reqURL := fmt.Sprintf("https://%s/api/v3/deployments/%s/%s", c.HostURL, deploymentID, permanentStr)

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creando la petición DELETE: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando DELETE: %w", err)
	}
	defer res.Body.Close()

	// Leer el body para obtener información de error si es necesario
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	// Considerar éxito los códigos 200, 204 (No Content) y 404 (ya no existe)
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("error eliminando deployment (status %d): %s", res.StatusCode, string(body))
}

// StartDeployment inicia todos los desktops de un deployment
func (c *Client) StartDeployment(deploymentID string) error {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployments/start/%s", c.HostURL, deploymentID)

	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creando la petición PUT: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando PUT: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error iniciando deployment (status %d): %s", res.StatusCode, string(body))
	}

	return nil
}

// StopDeployment detiene todos los desktops de un deployment
func (c *Client) StopDeployment(deploymentID string) error {
	reqURL := fmt.Sprintf("https://%s/api/v3/deployments/stop/%s", c.HostURL, deploymentID)

	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creando la petición PUT: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando PUT: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error deteniendo deployment (status %d): %s", res.StatusCode, string(body))
	}

	return nil
}

// GetTemplateInfo obtiene información del template necesaria para crear deployments
func (c *Client) GetTemplateInfo(templateID string) (map[string]interface{}, error) {
	// Usar el endpoint de templates normal
	reqURL := fmt.Sprintf("https://%s/api/v3/template/%s", c.HostURL, templateID)

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
		return nil, fmt.Errorf("error obteniendo template (status %d): %s", res.StatusCode, string(body))
	}

	var template map[string]interface{}
	if err := json.Unmarshal(body, &template); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	return template, nil
}
