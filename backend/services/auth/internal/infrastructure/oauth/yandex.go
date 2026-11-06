package oauth

import "golang.org/x/oauth2"

func NewYandex(origin string, credentials Credentials) *Provider {
	return &Provider{
		name:     "yandex",
		endpoint: "https://login.yandex.ru/info?format=json",
		config: oauth2.Config{
			ClientID:     credentials.ClientID,
			ClientSecret: credentials.ClientSecret,
			RedirectURL:  origin + "/api/v1/auth/callback/yandex",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://oauth.yandex.ru/authorize",
				TokenURL:  "https://oauth.yandex.ru/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{"login:info", "login:email"},
		},
	}
}
