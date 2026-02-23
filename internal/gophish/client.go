package gophish

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client GoPhish API клиент
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Config конфигурация GoPhish
type Config struct {
	APIKey  string
	BaseURL string
	SkipVerify bool
}

// NewClient создает новый GoPhish клиент
func NewClient(config *Config) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipVerify},
	}
	
	return &Client{
		APIKey: config.APIKey,
		BaseURL: config.BaseURL,
		HTTPClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// Campaign кампания GoPhish
type Campaign struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CreatedDate time.Time `json:"created_date"`
	SendByDate  time.Time `json:"send_by_date"`
	Complete    bool      `json:"complete"`
	Results     []Result `json:"results"`
	Page        string    `json:"page"`
	Template    string    `json:"template"`
	URL         string    `json:"url"`
}

// Result результат кампании
type Result struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	Time      time.Time `json:"time"`
}

// Group группа пользователей
type Group struct {
	ID        int64    `json:"id"`
	Name      string    `json:"name"`
	Modified  time.Time `json:"modified"`
	Targets   []Target `json:"targets"`
}

// Target цель
type Target struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

// Template email шаблон
type Template struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	Text        string    `json:"text"`
	HTML        string    `json:"html"`
	ModifiedDate time.Time `json:"modified_date"`
}

// LandingPage лендинг страница
type LandingPage struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	HTML        string    `json:"html"`
	ModifiedDate time.Time `json:"modified_date"`
}

// Profile профиль отправки
type Profile struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Host      string  `json:"host"`
	From      string  `json:"from"`
	ReplyTo   string  `json:"reply_to"`
}

// API Response
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Campaigns получение списка кампаний
func (c *Client) Campaigns() ([]Campaign, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/campaigns/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var campaigns []Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaigns); err != nil {
		return nil, err
	}

	return campaigns, nil
}

// Campaign получение кампании по ID
func (c *Client) Campaign(id int64) (*Campaign, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/campaigns/%d", c.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var campaign Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaign); err != nil {
		return nil, err
	}

	return &campaign, nil
}

// CreateCampaign создание кампании
type CreateCampaignRequest struct {
	Name        string   `json:"name"`
	Page        int64    `json:"page"`
	Template    int64    `json:"template"`
	URL         string   `json:"url"`
	Group       int64    `json:"group"`
	SendByDate string   `json:"send_by_date"`
	Smtp        int64    `json:"smtp"`
}

func (c *Client) CreateCampaign(req *CreateCampaignRequest) (*Campaign, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/campaigns/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var campaign Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaign); err != nil {
		return nil, err
	}

	return &campaign, nil
}

// DeleteCampaign удаление кампании
func (c *Client) DeleteCampaign(id int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/campaigns/%d", c.BaseURL, id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Groups получение списка групп
func (c *Client) Groups() ([]Group, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/groups/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// CreateGroup создание группы
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

func (c *Client) CreateGroup(req *CreateGroupRequest) (*Group, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/groups/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// Templates получение списка шаблонов
func (c *Client) Templates() ([]Template, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/templates/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var templates []Template
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// Pages получение списка лендингов
func (c *Client) Pages() ([]LandingPage, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/pages/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pages []LandingPage
	if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
		return nil, err
	}

	return pages, nil
}

// Profiles получение списка профилей отправки
func (c *Client) Profiles() ([]Profile, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/smtp/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

// Summary получение сводки
type Summary struct {
	Campaigns   int `json:"campaigns"`
	Results      int `json:"results"`
	Groups       int `json:"groups"`
	Templates    int `json:"templates"`
	Pages        int `json:"pages"`
	Profiles     int `json:"profiles"`
}

func (c *Client) Summary() (*Summary, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/summary/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var summary Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}
