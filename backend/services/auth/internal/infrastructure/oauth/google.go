package oauth

import "golang.org/x/oauth2"

func NewGoogle(origin string, credentials Credentials) *Provider {
	return &Provider{
		name:     "google",
		endpoint: "https://openidconnect.googleapis.com/v1/userinfo",
		config: oauth2.Config{
			ClientID:     credentials.ClientID,
			ClientSecret: credentials.ClientSecret,
			RedirectURL:  origin + "/api/v1/auth/callback/google",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL:  "https://oauth2.googleapis.com/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{"openid", "profile", "email"},
		},
	}
}
