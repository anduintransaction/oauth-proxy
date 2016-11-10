package proxy

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/anduintransaction/oauth-proxy/utils"
	"gottb.io/goru/config"
)

type Proxy struct {
	Provider      string   `config:"provider"`
	Scheme        string   `config:"scheme"`
	RedirectURI   string   `config:"redirect_uri"`
	RequestHost   string   `config:"request_host"`
	EndPoint      string   `config:"end_point"`
	ClientID      string   `config:"client_id"`
	ClientSecret  string   `config:"client_secret"`
	CallbackURI   string   `config:"callback_uri"`
	Organizations []string `config:"organizations"`
	Teams         []string `config:"teams"`
	organizations utils.StringSet
	teams         utils.StringSet
	target        *url.URL
	reverseProxy  *httputil.ReverseProxy
}

func (p *Proxy) HasOrg(org string) bool {
	return p.organizations.Has(org)
}

func (p *Proxy) HasTeam(team string) bool {
	return p.teams.Has(team)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) createReverseProxy() {
	p.reverseProxy = httputil.NewSingleHostReverseProxy(p.target)
}

var Config struct {
	Provider      string `config:"provider"`
	ClientID      string `config:"client_id"`
	ClientSecret  string `config:"client_secret"`
	CallbackURI   string `config:"callback_uri"`
	StateTimeout  int    `config:"state_timeout"`
	CookieTimeout int    `config:"cookie_timeout"`
	CookieName    string `config:"cookie_name"`
	CheckVersion  bool   `config:"check_version"`
	Version       int64
}

var proxies []*Proxy
var proxyMap map[string]*Proxy

func Start(config *config.Config) error {
	oauthConfig, err := config.Get("oauth")
	if err != nil {
		return err
	}
	err = oauthConfig.Unmarshal(&Config)
	if err != nil {
		return err
	}

	proxyConfig, err := config.Get("proxy")
	if err != nil {
		return err
	}
	err = proxyConfig.Unmarshal(&proxies)
	if err != nil {
		return err
	}
	proxyMap = make(map[string]*Proxy)
	for _, proxy := range proxies {
		if proxy.Provider == "" {
			proxy.Provider = Config.Provider
		}
		if proxy.ClientID == "" {
			proxy.ClientID = Config.ClientID
		}
		if proxy.ClientSecret == "" {
			proxy.ClientSecret = Config.ClientSecret
		}
		if proxy.CallbackURI == "" {
			proxy.CallbackURI = Config.CallbackURI
		}
		proxy.organizations = utils.NewStringSet(proxy.Organizations)
		proxy.teams = utils.NewStringSet(proxy.Teams)
		proxy.target, err = url.Parse(proxy.EndPoint)
		if err != nil {
			return err
		}
		proxy.createReverseProxy()
		proxyMap[proxy.RequestHost] = proxy
	}

	rand.Seed(time.Now().UnixNano())
	Config.Version = rand.Int63()
	defaultStateMap = newStateMap(Config.StateTimeout)
	return nil
}

func Stop(config *config.Config) error {
	defaultStateMap.quit()
	return nil
}

func GetProxy(requestHost string) *Proxy {
	return proxyMap[requestHost]
}
