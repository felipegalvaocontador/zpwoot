package contracts

import (
	"time"

	"zpwoot/internal/core/session"
)

type CreateSessionRequest struct {
	Name        string       `json:"name" validate:"required,min=3,max=50" example:"my-session"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	QRCode      bool         `json:"qrCode" example:"false"`
} // @name CreateSessionRequest

type ListSessionsRequest struct {
	IsConnected *bool   `json:"isConnected,omitempty" query:"isConnected" example:"true"`
	DeviceJID   *string `json:"deviceJid,omitempty" query:"deviceJid" example:"5511999999999@s.whatsapp.net"`
	Limit       int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset      int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0" example:"0"`
} // @name ListSessionsRequest

type SetProxyRequest struct {
	ProxyConfig ProxyConfig `json:"proxyConfig" validate:"required"`
} // @name SetProxyRequest

type PairPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,e164" example:"+5511999999999"`
} // @name PairPhoneRequest

type CreateSessionResponse struct {
	ID          string       `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	Name        string       `json:"name" example:"my-session"`
	IsConnected bool         `json:"isConnected" example:"false"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	QRCode      string       `json:"qrCode,omitempty" example:"2@abc123..."`
	QRCodeImage string       `json:"qrCodeImage,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
	CreatedAt   time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} // @name CreateSessionResponse

type SessionResponse struct {
	ID              string       `json:"id" example:"session-123"`
	Name            string       `json:"name" example:"my-whatsapp-session"`
	DeviceJID       string       `json:"deviceJid,omitempty" example:"5511999999999@s.whatsapp.net"`
	IsConnected     bool         `json:"isConnected" example:"false"`
	ConnectionError *string      `json:"connectionError,omitempty" example:"Connection timeout"`
	ProxyConfig     *ProxyConfig `json:"proxyConfig,omitempty"`
	CreatedAt       time.Time    `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time    `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
	ConnectedAt     *time.Time   `json:"connectedAt,omitempty" example:"2024-01-01T00:00:30Z"`
} // @name SessionResponse

type SessionInfoResponse struct {
	Session    *SessionResponse    `json:"session"`
	DeviceInfo *DeviceInfoResponse `json:"deviceInfo,omitempty"`
} // @name SessionInfoResponse

type ListSessionsResponse struct {
	Sessions []SessionInfoResponse `json:"sessions"`
	Total    int                   `json:"total" example:"10"`
	Limit    int                   `json:"limit" example:"20"`
	Offset   int                   `json:"offset" example:"0"`
} // @name ListSessionsResponse

type ConnectSessionResponse struct {
	Success     bool   `json:"success" example:"true"`
	Message     string `json:"message" example:"Session connection initiated successfully"`
	QRCode      string `json:"qrCode,omitempty" example:"2@abc123..."`
	QRCodeImage string `json:"qrCodeImage,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
} // @name ConnectSessionResponse

type QRCodeResponse struct {
	QRCode      string    `json:"qrCode" example:"2@abc123def456..." description:"Raw QR code string"`
	QRCodeImage string    `json:"qrCodeImage,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..." description:"Base64 encoded QR code image"`
	ExpiresAt   time.Time `json:"expiresAt" example:"2024-01-01T00:01:00Z"`
	Timeout     int       `json:"timeoutSeconds" example:"60"`
} // @name QRCodeResponse

type ProxyResponse struct {
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
} // @name ProxyResponse

type SessionStatsResponse struct {
	Total     int `json:"total" example:"10"`
	Connected int `json:"connected" example:"3"`
	Offline   int `json:"offline" example:"7"`
} // @name SessionStatsResponse

type ProxyConfig struct {
	Type     string `json:"type" validate:"required,oneof=http socks5" example:"http"`
	Host     string `json:"host" validate:"required,hostname_rfc1123" example:"proxy.example.com"`
	Port     int    `json:"port" validate:"required,min=1,max=65535" example:"8080"`
	Username string `json:"username,omitempty" example:"proxyuser"`
	Password string `json:"password,omitempty" example:"proxypass123"`
} // @name ProxyConfig

type DeviceInfoResponse struct {
	Platform    string `json:"platform" example:"android"`
	DeviceModel string `json:"deviceModel" example:"Samsung Galaxy S21"`
	OSVersion   string `json:"osVersion" example:"11"`
	AppVersion  string `json:"appVersion" example:"2.21.4.18"`
} // @name DeviceInfoResponse

func (r *CreateSessionRequest) ToCreateSessionRequest() *session.CreateSessionRequest {
	req := &session.CreateSessionRequest{
		Name:        r.Name,
		AutoConnect: r.QRCode,
	}

	if r.ProxyConfig != nil {
		req.ProxyConfig = &session.ProxyConfig{
			Type:     r.ProxyConfig.Type,
			Host:     r.ProxyConfig.Host,
			Port:     r.ProxyConfig.Port,
			Username: r.ProxyConfig.Username,
			Password: r.ProxyConfig.Password,
		}
	}

	return req
}

func FromSession(s *session.Session) *SessionResponse {
	response := &SessionResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		IsConnected: s.IsConnected,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}

	if s.DeviceJID != nil {
		response.DeviceJID = *s.DeviceJID
	}

	if s.ConnectionError != nil {
		response.ConnectionError = s.ConnectionError
	}

	if s.ConnectedAt != nil {
		response.ConnectedAt = s.ConnectedAt
	}

	if s.ProxyConfig != nil {
		response.ProxyConfig = &ProxyConfig{
			Type:     s.ProxyConfig.Type,
			Host:     s.ProxyConfig.Host,
			Port:     s.ProxyConfig.Port,
			Username: s.ProxyConfig.Username,
			Password: s.ProxyConfig.Password,
		}
	}

	return response
}

func FromQRCodeResponse(qr *session.QRCodeResponse) *QRCodeResponse {
	return &QRCodeResponse{
		QRCode:    qr.QRCode,
		ExpiresAt: qr.ExpiresAt,
		Timeout:   qr.Timeout,
	}
}
