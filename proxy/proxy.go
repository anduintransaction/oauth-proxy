package proxy

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"regexp"

	"github.com/anduintransaction/oauth-proxy/utils"
	"gottb.io/goru/config"
	"gottb.io/goru/log"
)

type whilelist struct {
	method string
	path   *regexp.Regexp
}

type Proxy struct {
	Provider      string   `config:"provider"`
	Scheme        string   `config:"scheme"`
	RedirectURI   string   `config:"redirect_uri"`
	RequestHost   string   `config:"request_host"`
	EndPoint      string   `config:"end_point"`
	PreserveHost  bool     `config:"preserve_host"`
	ClientID      string   `config:"client_id"`
	ClientSecret  string   `config:"client_secret"`
	CallbackURI   string   `config:"callback_uri"`
	Organizations []string `config:"organizations"`
	Teams         []string `config:"teams"`
	Whitelists    []string `config:"whitelists"`
	organizations utils.StringSet
	teams         utils.StringSet
	target        *url.URL
	whitelists    []*whilelist
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

func (p *Proxy) IsWhiteList(method, path string) bool {
	for _, w := range p.whitelists {
		if w.method != "ANY" && w.method != method {
			continue
		}
		path = strings.TrimRight(path, "/")
		if path == "" {
			path = "/"
		}
		matched := w.path.MatchString(path)
		if matched {
			return true
		}
	}
	return false
}

func (p *Proxy) createReverseProxy() {
	p.reverseProxy = &httputil.ReverseProxy{
		Director: p.transformRequest,
	}
}

func (p *Proxy) transformRequest(req *http.Request) {
	req.URL.Scheme = p.target.Scheme
	req.URL.Host = p.target.Host
	req.URL.Path = p.singleJoiningSlash(p.target.Path, req.URL.Path)
	targetQuery := p.target.RawQuery
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
	if !p.PreserveHost {
		req.Header.Set("Host", p.target.Host)
		req.Host = p.target.Host
	}
}

func (p *Proxy) singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
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
		proxy.whitelists = []*whilelist{}
		for _, wl := range proxy.Whitelists {
			w := &whilelist{}
			pieces := strings.SplitN(wl, ":", 2)
			if len(pieces) == 1 {
				w.method = "ANY"
				w.path, err = regexp.Compile("^" + pieces[0] + "$")
			} else {
				w.method = strings.ToUpper(pieces[0])
				w.path, err = regexp.Compile("^" + pieces[1] + "$")
			}
			if err != nil {
				return err
			}
			proxy.whitelists = append(proxy.whitelists, w)
		}
		proxy.createReverseProxy()
		proxyMap[proxy.RequestHost] = proxy
		log.Debug(proxy)
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
