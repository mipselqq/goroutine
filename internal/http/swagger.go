package http

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SwaggerAutoAuthOnLoginScript(loginPath string) string {
	return `
		function(response) {
			try {
				const isLoginRequest = response?.url?.includes("` + loginPath + `")
				if (!isLoginRequest || response?.status !== 200) {
					return response
				}
				const payload = response?.body
				const token = payload?.token

				if (!token) {
					throw new Error("Failed to auto-authorize API key on login: no token in response")
				}
				window.ui.preauthorizeApiKey("BearerAuth", "Bearer " + token);
			} catch (e) {
				console.error("Failed to auto-authorize API key on login", e)
			}

			return response;
		}
	`
}

func NewSwaggerHandler(basePath, loginPath string) http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL(basePath+"doc.json"),
		httpSwagger.PersistAuthorization(true), // Doesn't work with auto-authorize at /login
		httpSwagger.UIConfig(map[string]string{
			"responseInterceptor": SwaggerAutoAuthOnLoginScript(loginPath),
		}),
	)
}
