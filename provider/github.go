package provider

import (
	"encoding/json"
	"net/http"
	"net/url"

	"gottb.io/goru/errors"
	"gottb.io/goru/log"

	"github.com/anduintransaction/oauth-proxy/proxy"
	"github.com/anduintransaction/oauth-proxy/utils"
)

const (
	githubDefaultRedirectURI     = "https://github.com/login/oauth/authorize"
	githubDefaultTokenRequestURI = "https://github.com/login/oauth/access_token"
	githubDefaultAPIURI          = "https://api.github.com"
)

type GithubProvider struct {
}

func (p *GithubProvider) RedirectURI(proxy *proxy.Proxy, randomState string) string {
	v := url.Values{}
	v.Add("client_id", proxy.ClientID)
	v.Add("redirect_uri", proxy.CallbackURI)
	v.Add("scope", "user:email,read:org")
	v.Add("state", randomState)
	v.Add("allow_signup", "false")
	return githubDefaultRedirectURI + "?" + v.Encode()
}

func (p *GithubProvider) ErrorString(request *http.Request) string {
	return request.URL.Query().Get("error_description")
}

func (p *GithubProvider) RequestToken(state *proxy.State, code string) (string, error) {
	tokenRequest := &struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
		State        string `json:"state"`
	}{
		ClientID:     state.Proxy.ClientID,
		ClientSecret: state.Proxy.ClientSecret,
		Code:         code,
		RedirectURI:  state.Proxy.RedirectURI,
		State:        state.Name,
	}
	statusCode, responseContent, err := utils.HTTPRequestJSON("POST", githubDefaultTokenRequestURI, tokenRequest, nil)
	if err != nil {
		return "", err
	}
	if statusCode >= 300 {
		return "", errors.Errorf("invalid status code: %d", statusCode)
	}
	log.Infof("Get response for state %s: %s", state.Name, string(responseContent))
	tokenResponse := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = json.Unmarshal(responseContent, &tokenResponse)
	if err != nil {
		return "", errors.Wrap(err)
	}
	if tokenResponse.AccessToken == "" {
		return "", errors.Errorf("invalid token response: %s", string(responseContent))
	}
	return tokenResponse.AccessToken, nil
}

func (p *GithubProvider) VerifyUser(state *proxy.State, token string) (*proxy.UserInfo, error) {
	user, err := p.getUserInfo(token)
	if err != nil {
		return nil, err
	}
	hasOrg, err := p.verifyOrg(state, token)
	if err != nil {
		return nil, err
	}
	if !hasOrg {
		return nil, errors.Errorf("no suitable organization")
	}
	hasTeam, err := p.verifyTeam(state, token)
	if err != nil {
		return nil, err
	}
	if !hasTeam {
		return nil, errors.Errorf("no suitable team")
	}
	return user, nil
}

func (p *GithubProvider) getUserInfo(token string) (*proxy.UserInfo, error) {
	headers := map[string]string{
		"Authorization": "token " + token,
	}
	statusCode, responseContent, err := utils.HTTPRequestJSON("GET", githubDefaultAPIURI+"/user", "", headers)
	if err != nil {
		return nil, err
	}
	if statusCode >= 300 {
		log.Errorf("Invalid status code %d for token %s", statusCode, token)
		return nil, errors.Errorf("invalid status code: %d", statusCode)
	}
	user := &proxy.UserInfo{}
	err = json.Unmarshal(responseContent, user)
	if err != nil {
		log.Errorf("Cannot decode json: %s", string(responseContent))
		return nil, errors.Wrap(err)
	}
	log.Infof("User found for token %s: %s - %s", token, user.Name, user.Email)
	return user, nil
}

func (p *GithubProvider) verifyOrg(state *proxy.State, token string) (bool, error) {
	headers := map[string]string{
		"Authorization": "token " + token,
	}
	statusCode, responseContent, err := utils.HTTPRequestJSON("GET", githubDefaultAPIURI+"/user/orgs", "", headers)
	if err != nil {
		return false, err
	}
	if statusCode >= 300 {
		log.Errorf("Invalid status code %d for token %s", statusCode, token)
		return false, errors.Errorf("invalid status code: %d", statusCode)
	}
	orgResponse := []struct {
		Login string `json:"login"`
	}{}
	err = json.Unmarshal(responseContent, &orgResponse)
	if err != nil {
		log.Errorf("Cannot decode json: %s", string(responseContent))
		return false, errors.Wrap(err)
	}
	log.Infof("Organizations of %s: %v", token, orgResponse)
	hasOrg := false
	for _, org := range orgResponse {
		if state.Proxy.HasOrg(org.Login) {
			log.Infof("Found organization: %s", org.Login)
			hasOrg = true
			break
		}
	}
	return hasOrg, nil
}

func (p *GithubProvider) verifyTeam(state *proxy.State, token string) (bool, error) {
	hasTeam := true
	if len(state.Proxy.Teams) > 0 {
		hasTeam = false
		headers := map[string]string{
			"Authorization": "token " + token,
		}
		statusCode, responseContent, err := utils.HTTPRequestJSON("GET", githubDefaultAPIURI+"/user/teams", "", headers)
		if err != nil {
			return false, err
		}
		if statusCode >= 300 {
			log.Errorf("Invalid status code %d for token %s", statusCode, token)
			return false, errors.Errorf("invalid status code: %d", statusCode)
		}
		teamResponse := []*struct {
			Name string `json:"name"`
		}{}
		err = json.Unmarshal(responseContent, &teamResponse)
		if err != nil {
			log.Errorf("Cannot decode json: %s", string(responseContent))
			return false, errors.Wrap(err)
		}
		log.Infof("Teams of %s: %v", token, teamResponse)
		for _, team := range teamResponse {
			if state.Proxy.HasTeam(team.Name) {
				log.Infof("Found team: %s", team.Name)
				hasTeam = true
				break
			}
		}
	}
	return hasTeam, nil
}
