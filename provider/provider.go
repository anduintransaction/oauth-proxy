package provider

import (
	"net/http"

	"github.com/anduintransaction/oauth-proxy/proxy"
)

type Provider interface {
	RedirectURI(proxy *proxy.Proxy, randomState string) string
	ErrorString(request *http.Request) string
	RequestToken(state *proxy.State, code string) (string, error)
	VerifyUser(state *proxy.State, token string) (*proxy.UserInfo, error)
}

func GetProvider(name string) Provider {
	switch name {
	case "github":
		return &GithubProvider{}
	default:
		return nil
	}
}
