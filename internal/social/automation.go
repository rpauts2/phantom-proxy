package social

import (
	"bytes"
	"context"
	"html/template"
	"sync"
	"time"
)

// TargetProfile for personalized phishing
type TargetProfile struct {
	Email       string
	FullName    string
	Company     string
	Role        string
	Interests   []string
	OSINTData   map[string]string
}

// EmailTemplate for campaign
type EmailTemplate struct {
	Subject    string
	BodyHTML   string
	BodyText   string
	Variables  []string // {{.Name}}, {{.Company}}, etc.
}

// CampaignConfig for mass mailing
type CampaignConfig struct {
	Name         string
	Template     *EmailTemplate
	Targets      []TargetProfile
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	FromAddress  string
	FromName     string
	RateLimit    int           // emails per minute
	DelayBetween time.Duration
	MaxPerDay    int
}

// CampaignResult single send result
type CampaignResult struct {
	TargetEmail string
	Sent        bool
	Error       string
	SentAt      time.Time
}

// AutomationEngine for SE campaigns
type AutomationEngine struct {
	config *CampaignConfig
	mu     sync.Mutex
}

// NewAutomationEngine creates SE automation engine
func NewAutomationEngine(cfg *CampaignConfig) *AutomationEngine {
	return &AutomationEngine{config: cfg}
}

// RenderTemplate applies template to profile
func (e *AutomationEngine) RenderTemplate(tpl *EmailTemplate, profile TargetProfile) (subject, body string, err error) {
	data := map[string]interface{}{
		"Name":     profile.FullName,
		"Company":  profile.Company,
		"Role":     profile.Role,
		"Email":    profile.Email,
		"Interests": profile.Interests,
	}
	for k, v := range profile.OSINTData {
		data[k] = v
	}

	t := template.Must(template.New("subject").Parse(tpl.Subject))
	var subjBuf bytes.Buffer
	if err := t.Execute(&subjBuf, data); err != nil {
		return "", "", err
	}
	subject = subjBuf.String()

	t2 := template.Must(template.New("body").Parse(tpl.BodyHTML))
	var bodyBuf bytes.Buffer
	if err := t2.Execute(&bodyBuf, data); err != nil {
		return "", "", err
	}
	body = bodyBuf.String()
	return subject, body, nil
}

// SendCampaign sends emails with rate limiting
func (e *AutomationEngine) SendCampaign(ctx context.Context) ([]CampaignResult, error) {
	var results []CampaignResult
	delay := e.config.DelayBetween
	if delay == 0 {
		delay = time.Minute / time.Duration(e.config.RateLimit)
	}
	for _, target := range e.config.Targets {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
		e.mu.Lock()
		subject, body, err := e.RenderTemplate(e.config.Template, target)
		e.mu.Unlock()
		if err != nil {
			results = append(results, CampaignResult{TargetEmail: target.Email, Error: err.Error()})
			continue
		}
		// В production: smtp.SendMail(...)
		_ = subject
		_ = body
		results = append(results, CampaignResult{
			TargetEmail: target.Email,
			Sent:        true,
			SentAt:      time.Now(),
		})
		time.Sleep(delay)
	}
	return results, nil
}

// ProfilingData for target profiling (без скрапинга — данные вводятся оператором)
type ProfilingData struct {
	TargetID   string
	LinkedIn   string
	SocialLinks []string
	JobTitle   string
	Company    string
	Skills     []string
	ParsedAt   time.Time
}

// ProfileBuilder builds profile from OSINT (структуры — реализация через внешние инструменты)
type ProfileBuilder struct{}

// BuildProfile creates profile from provided data
func (p *ProfileBuilder) BuildProfile(data *ProfilingData) TargetProfile {
	return TargetProfile{
		Email:    "",
		FullName: data.JobTitle,
		Company:  data.Company,
		Role:     data.JobTitle,
		Interests: data.Skills,
		OSINTData: map[string]string{
			"linkedin": data.LinkedIn,
		},
	}
}
