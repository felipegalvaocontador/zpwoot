package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "ZPWoot API Support",
            "url": "https://github.com/zpwoot/zpwoot",
            "email": "support@zpwoot.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {},
    "definitions": {},
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "API Key authentication. Use 'Bearer YOUR_API_KEY' or just 'YOUR_API_KEY'"
        },
        "ApiKeyHeader": {
            "type": "apiKey",
            "name": "X-API-Key",
            "in": "header",
            "description": "API Key authentication via X-API-Key header"
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "2.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "ZPWoot WhatsApp API",
	Description:      "A comprehensive WhatsApp Business API built with Go. Provides endpoints for session management, messaging, contacts, groups, media handling, and integrations with Chatwoot.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
