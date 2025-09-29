package session

import (
	"time"

	domainSession "zpwoot/internal/domain/session"
)

type ProxyConfig struct {
	Type     string `json:"type" example:"http"` // http, socks5
	Host     string `json:"host" example:"proxy.example.com"`
	Port     int    `json:"port" example:"8080"`
	Username string `json:"username,omitempty" example:"proxyuser"`
	Password string `json:"password,omitempty" example:"proxypass123"`
} //@name ProxyConfig

type CreateSessionRequest struct {
	Name        string       `json:"name" validate:"required,min=3,max=50" example:"my-session"`
	QrCode      bool         `json:"qrCode" example:"false"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name CreateSessionRequest

type CreateSessionResponse struct {
	ID          string       `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	Name        string       `json:"name" example:"my-session"`
	IsConnected bool         `json:"isConnected" example:"false"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	QrCode      string       `json:"qrCode,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
	Code        string       `json:"code,omitempty" example:"2@abc123..."`
	CreatedAt   time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} //@name CreateSessionResponse

type UpdateSessionRequest struct {
	Name        *string      `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Session Name"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name UpdateSessionRequest

type ListSessionsRequest struct {
	IsConnected *bool   `json:"isConnected,omitempty" query:"isConnected" example:"true"`
	DeviceJid   *string `json:"deviceJid,omitempty" query:"deviceJid" example:"5511999999999@s.Wameow.net"`
	Limit       int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset      int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0" example:"0"`
} //@name ListSessionsRequest

type ListSessionsResponse struct {
	Sessions []SessionInfoResponse `json:"sessions"`
	Total    int                   `json:"total" example:"10"`
	Limit    int                   `json:"limit" example:"20"`
	Offset   int                   `json:"offset" example:"0"`
} //@name ListSessionsResponse

type SessionInfoResponse struct {
	Session    *SessionResponse    `json:"session"`
	DeviceInfo *DeviceInfoResponse `json:"deviceInfo,omitempty"`
} //@name SessionInfoResponse

type SessionResponse struct {
	ID              string       `json:"id" example:"session-123"`
	Name            string       `json:"name" example:"my-Wameow-session"`
	DeviceJid       string       `json:"deviceJid,omitempty" example:"5511999999999@s.Wameow.net"`
	IsConnected     bool         `json:"isConnected" example:"false"`
	ConnectionError *string      `json:"connectionError,omitempty" example:"Connection timeout"`
	ProxyConfig     *ProxyConfig `json:"proxyConfig,omitempty"`
	CreatedAt       time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time    `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
	ConnectedAt     *time.Time   `json:"connectedAt,omitempty" example:"2024-01-01T00:00:30Z"`
} //@name SessionResponse

type DeviceInfoResponse struct {
	Platform    string `json:"platform" example:"android"`
	DeviceModel string `json:"deviceModel" example:"Samsung Galaxy S21"`
	OSVersion   string `json:"osVersion" example:"11"`
	AppVersion  string `json:"appVersion" example:"2.21.4.18"`
} //@name DeviceInfoResponse

type PairPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,e164" example:"+5511987654321"`
} //@name PairPhoneRequest

type QRCodeResponse struct {
	QRCode      string    `json:"qrCode" example:"2@abc123def456..."`
	QRCodeImage string    `json:"qrCodeImage,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="`
	ExpiresAt   time.Time `json:"expiresAt" example:"2024-01-01T00:01:00Z"`
	Timeout     int       `json:"timeoutSeconds" example:"60"`
} //@name QRCodeResponse

type SetProxyRequest struct {
	ProxyConfig ProxyConfig `json:"proxyConfig"`
} //@name SetProxyRequest

type ProxyResponse struct {
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} //@name ProxyResponse

type ConnectSessionResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Session connection initiated successfully"`
	QrCode  string `json:"qrCode,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
	Code    string `json:"code,omitempty" example:"2@abc123..."`
} //@name ConnectSessionResponse

func (r *CreateSessionRequest) ToCreateSessionRequest() *domainSession.CreateSessionRequest {
	var proxyConfig *domainSession.ProxyConfig
	if r.ProxyConfig != nil {
		proxyConfig = &domainSession.ProxyConfig{
			Type:     r.ProxyConfig.Type,
			Host:     r.ProxyConfig.Host,
			Port:     r.ProxyConfig.Port,
			Username: r.ProxyConfig.Username,
			Password: r.ProxyConfig.Password,
		}
	}
	return &domainSession.CreateSessionRequest{
		Name:        r.Name,
		QrCode:      r.QrCode,
		ProxyConfig: proxyConfig,
	}
}

func FromSession(s *domainSession.Session) *SessionResponse {
	var proxyConfig *ProxyConfig
	if s.ProxyConfig != nil {
		proxyConfig = &ProxyConfig{
			Type:     s.ProxyConfig.Type,
			Host:     s.ProxyConfig.Host,
			Port:     s.ProxyConfig.Port,
			Username: s.ProxyConfig.Username,
			Password: s.ProxyConfig.Password,
		}
	}

	response := &SessionResponse{
		ID:              s.ID.String(),
		Name:            s.Name,
		IsConnected:     s.IsConnected,
		ConnectionError: s.ConnectionError,
		ProxyConfig:     proxyConfig,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		ConnectedAt:     s.ConnectedAt,
	}

	if s.DeviceJid != "" {
		response.DeviceJid = s.DeviceJid
	}

	return response
}

func FromSessionInfo(si *domainSession.SessionInfo) *SessionInfoResponse {
	response := &SessionInfoResponse{}

	if si.Session != nil {
		response.Session = FromSession(si.Session)
	}

	if si.DeviceInfo != nil {
		response.DeviceInfo = &DeviceInfoResponse{
			Platform:    si.DeviceInfo.Platform,
			DeviceModel: si.DeviceInfo.DeviceModel,
			OSVersion:   si.DeviceInfo.OSVersion,
			AppVersion:  si.DeviceInfo.AppVersion,
		}
	}

	return response
}

func FromQRCodeResponse(qr *domainSession.QRCodeResponse) *QRCodeResponse {
	return &QRCodeResponse{
		QRCode:      qr.QRCode,
		QRCodeImage: qr.QRCodeImage,
		ExpiresAt:   qr.ExpiresAt,
		Timeout:     qr.Timeout,
	}
}
