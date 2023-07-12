package mockghauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/dosquad/mock-oauth-test-server/internal/staticsrc"
)

type GitHubOAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type GitHubOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GitHubAPIUser struct {
	Login                   string            `json:"login"`
	ID                      int               `json:"id"`
	NodeID                  string            `json:"node_id"`
	AvatarURL               string            `json:"avatar_url"`
	GravatarID              string            `json:"gravatar_id"`
	URL                     string            `json:"url"`
	HTMLURL                 string            `json:"html_url"`
	FollowersURL            string            `json:"followers_url"`
	FollowingURL            string            `json:"following_url"`
	GistsURL                string            `json:"gists_url"`
	StarredURL              string            `json:"starred_url"`
	SubscriptionsURL        string            `json:"subscriptions_url"`
	OrganizationsURL        string            `json:"organizations_url"`
	ReposURL                string            `json:"repos_url"`
	EventsURL               string            `json:"events_url"`
	ReceivedEventsURL       string            `json:"received_events_url"`
	Type                    string            `json:"type"`
	SiteAdmin               bool              `json:"site_admin"`
	Name                    string            `json:"name"`
	Company                 string            `json:"company"`
	Blog                    string            `json:"blog"`
	Location                string            `json:"location"`
	Email                   string            `json:"email"`
	Hireable                bool              `json:"hireable"`
	Bio                     string            `json:"bio"`
	TwitterUsername         string            `json:"twitter_username"`
	PublicRepos             int               `json:"public_repos"`
	PublicGists             int               `json:"public_gists"`
	Followers               int               `json:"followers"`
	Following               int               `json:"following"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
	PrivateGists            int               `json:"private_gists"`
	TotalPrivateRepos       int               `json:"total_private_repos"`
	OwnedPrivateRepos       int               `json:"owned_private_repos"`
	DiskUsage               int               `json:"disk_usage"`
	Collaborators           int               `json:"collaborators"`
	TwoFactorAuthentication bool              `json:"two_factor_authentication"`
	Plan                    GitHubAPIUserPlan `json:"plan"`
}

func urlMustResolve(baseURL *url.URL, relativePath string) *url.URL {
	relURL, err := url.Parse(relativePath)
	if err != nil {
		log.Panicf("unable to parse URL: %s", err)
	}

	return baseURL.ResolveReference(relURL)
}

func timeMustParseDef(value string) time.Time {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		log.Panicf("unable to parse time(%s): %s", value, err)
	}

	return ts
}

func DefaultGitHubAPIUser(baseURL *url.URL) (*GitHubAPIUser, error) {
	var user GitHubAPIUser
	buf, bufErr := staticsrc.Content.ReadFile("api_v3_user.json")
	if bufErr != nil {
		return nil, bufErr
	}

	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&user); err != nil {
		return nil, err
	}

	user.AvatarURL = urlMustResolve(baseURL, "/images/error/octocat_happy.gif").String()
	user.URL = urlMustResolve(baseURL, "/api/v3/users/octocat").String()
	user.HTMLURL = urlMustResolve(baseURL, "/octocat").String()
	user.FollowersURL = urlMustResolve(baseURL, "/api/v3/users/octocat/followers").String()
	user.FollowingURL = urlMustResolve(baseURL, "/api/v3/users/octocat/following{/other_user}").String()
	user.GistsURL = urlMustResolve(baseURL, "/api/v3/users/octocat/gists{/gist_id}").String()
	user.StarredURL = urlMustResolve(baseURL, "/api/v3/users/octocat/starred{/owner}{/repo}").String()
	user.SubscriptionsURL = urlMustResolve(baseURL, "/api/v3/users/octocat/subscriptions").String()
	user.OrganizationsURL = urlMustResolve(baseURL, "/api/v3/users/octocat/orgs").String()
	user.ReposURL = urlMustResolve(baseURL, "/api/v3/users/octocat/repos").String()
	user.EventsURL = urlMustResolve(baseURL, "/api/v3/users/octocat/events{/privacy}").String()
	user.ReceivedEventsURL = urlMustResolve(baseURL, "/api/v3/users/octocat/received_events").String()
	user.CreatedAt = timeMustParseDef("2008-01-14T04:33:35Z")
	user.UpdatedAt = timeMustParseDef("2008-01-14T04:33:35Z")

	return &user, nil
}

type GitHubAPIUserPlan struct {
	Name          string `json:"name"`
	Space         int    `json:"space"`
	PrivateRepos  int    `json:"private_repos"`
	Collaborators int    `json:"collaborators"`
}

type GitHubAPIError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

func UnauthorizedGitHubAPIError(err ...any) *GitHubAPIError {
	out := &GitHubAPIError{
		Message:          "Must authenticate to access this API.",
		DocumentationURL: "https://docs.github.com/enterprise-server@3.8/rest",
	}
	if len(err) > 0 {
		switch v := err[0].(type) {
		case error:
			out.Message = v.Error()
		case string:
			out.Message = v
		case fmt.Stringer:
			out.Message = v.String()
		}
	}

	return out
}
