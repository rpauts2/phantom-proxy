package vishing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// VishingClient клиент для Vishing сервиса
type VishingClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// CallRequest запрос на звонок
type CallRequest struct {
	PhoneNumber  string                 `json:"phone_number"`
	VoiceProfile string                 `json:"voice_profile"`
	Scenario     string                 `json:"scenario"`
	CustomData   map[string]interface{} `json:"custom_data,omitempty"`
}

// CallResponse ответ на звонок
type CallResponse struct {
	Success      bool   `json:"success"`
	CallID       string `json:"call_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	RecordingURL string `json:"recording_url,omitempty"`
}

// NewVishingClient создаёт новый Vishing клиент
func NewVishingClient(baseURL string, logger *zap.Logger) *VishingClient {
	return &VishingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger,
	}
}

// MakeCall совершает звонок
func (v *VishingClient) MakeCall(ctx context.Context, req CallRequest) (*CallResponse, error) {
	v.logger.Info("Making vishing call",
		zap.String("phone", req.PhoneNumber),
		zap.String("voice", req.VoiceProfile),
		zap.String("scenario", req.Scenario))
	
	// Сериализация запроса
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// HTTP запрос
	url := fmt.Sprintf("%s/api/v1/vishing/call", v.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := v.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Чтение ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Парсинг ответа
	var callResp CallResponse
	if err := json.Unmarshal(body, &callResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if !callResp.Success {
		return nil, fmt.Errorf("call failed: %s", callResp.Message)
	}
	
	v.logger.Info("Vishing call initiated",
		zap.String("call_id", callResp.CallID))
	
	return &callResp, nil
}

// GetCallStatus получает статус звонка
func (v *VishingClient) GetCallStatus(ctx context.Context, callID string) (map[string]interface{}, error) {
	v.logger.Debug("Getting call status",
		zap.String("call_id", callID))
	
	url := fmt.Sprintf("%s/api/v1/vishing/call/%s", v.baseURL, callID)
	
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	
	return result, nil
}

// RegisterVoice регистрирует голосовой профиль
func (v *VishingClient) RegisterVoice(ctx context.Context, name string, referenceAudio string, language string) error {
	v.logger.Info("Registering voice",
		zap.String("name", name),
		zap.String("audio", referenceAudio))
	
	type Request struct {
		Name     string `json:"name"`
		Audio    string `json:"reference_audio"`
		Language string `json:"language"`
	}
	
	jsonReq, _ := json.Marshal(Request{
		Name:     name,
		Audio:    referenceAudio,
		Language: language,
	})
	
	url := fmt.Sprintf("%s/api/v1/vishing/voice", v.baseURL)
	
	resp, err := v.httpClient.Post(url, "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GenerateScenario генерирует сценарий через LLM
func (v *VishingClient) GenerateScenario(ctx context.Context, targetService string, goal string) (map[string]interface{}, error) {
	v.logger.Info("Generating scenario",
		zap.String("service", targetService),
		zap.String("goal", goal))
	
	type Request struct {
		TargetService string `json:"target_service"`
		Goal          string `json:"goal"`
	}
	
	jsonReq, _ := json.Marshal(Request{
		TargetService: targetService,
		Goal:          goal,
	})
	
	url := fmt.Sprintf("%s/api/v1/vishing/generate-scenario", v.baseURL)
	
	resp, err := v.httpClient.Post(url, "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// HealthCheck проверяет здоровье сервиса
func (v *VishingClient) HealthCheck(ctx context.Context) (map[string]string, error) {
	url := fmt.Sprintf("%s/health", v.baseURL)
	
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}
