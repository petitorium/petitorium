package workspace

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PersistedCookie struct {
	Name     string    `yaml:"name"`
	Value    string    `yaml:"value"`
	Domain   string    `yaml:"domain,omitempty"`
	Path     string    `yaml:"path,omitempty"`
	Expires  time.Time `yaml:"expires,omitempty"`
	Secure   bool      `yaml:"secure"`
	HttpOnly bool      `yaml:"http_only"`
	SameSite string    `yaml:"same_site,omitempty"`
}

func (c *Cookie) toPersistedCookie() PersistedCookie {
	var expires time.Time
	if c.Expires != "" {
		expires, _ = time.Parse(time.RFC3339, c.Expires)
	}
	return PersistedCookie{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		Expires:  expires,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
		SameSite: c.SameSite,
	}
}

func (c *Cookie) fromPersistedCookie(pc PersistedCookie) {
	c.Name = pc.Name
	c.Value = pc.Value
	c.Domain = pc.Domain
	c.Path = pc.Path
	if !pc.Expires.IsZero() {
		c.Expires = pc.Expires.Format(time.RFC3339)
	}
	c.Secure = pc.Secure
	c.HttpOnly = pc.HttpOnly
	c.SameSite = pc.SameSite
}

func (cj *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if cj.Cookies == nil {
		cj.Cookies = []Cookie{}
	}

	for _, cookie := range cookies {
		cj.addCookie(cookie, u)
	}
}

func (cj *CookieJar) addCookie(cookie *http.Cookie, u *url.URL) {
	domain := cookie.Domain
	if domain == "" {
		domain = u.Host
	}

	path := cookie.Path
	if path == "" {
		path = "/"
	}

	secure := cookie.Secure
	httpOnly := cookie.HttpOnly
	sameSite := ""
	switch cookie.SameSite {
	case http.SameSiteNoneMode:
		sameSite = "none"
	case http.SameSiteLaxMode:
		sameSite = "lax"
	case http.SameSiteStrictMode:
		sameSite = "strict"
	}

	expires := ""
	if !cookie.Expires.IsZero() {
		expires = cookie.Expires.Format(time.RFC3339)
	}

	newCookie := Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Domain:   domain,
		Path:     path,
		Expires:  expires,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	}

	for i, c := range cj.Cookies {
		if c.Name == cookie.Name && c.Domain == domain && c.Path == path {
			cj.Cookies[i] = newCookie
			return
		}
	}

	cj.Cookies = append(cj.Cookies, newCookie)
}

func (cj *CookieJar) GetCookies(u *url.URL) []*http.Cookie {
	var result []*http.Cookie
	host := u.Host
	path := u.Path

	for _, c := range cj.Cookies {
		if !c.Enabled {
			continue
		}
		if cj.cookieMatches(c, host, path) {
			cookie := c.ToHttpCookie()
			result = append(result, cookie)
		}
	}

	return result
}

func (cj *CookieJar) cookieMatches(c Cookie, host, path string) bool {
	if !c.DomainMatches(host) {
		return false
	}

	if !c.PathMatches(path) {
		return false
	}

	return true
}

func (c *Cookie) DomainMatches(host string) bool {
	if c.Domain == "" {
		return true
	}

	if strings.HasPrefix(c.Domain, ".") {
		return strings.HasSuffix(host, c.Domain) || host == c.Domain[1:]
	}

	return host == c.Domain
}

func (c *Cookie) PathMatches(requestPath string) bool {
	if c.Path == "" || c.Path == "/" {
		return true
	}

	if strings.HasPrefix(requestPath, c.Path) {
		if c.Path == requestPath {
			return true
		}
		if len(requestPath) > len(c.Path) && requestPath[len(c.Path)] == '/' {
			return true
		}
		return false
	}

	return false
}

func (c *Cookie) ToHttpCookie() *http.Cookie {
	cookie := &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
	}

	if c.Domain != "" {
		cookie.Domain = c.Domain
	}

	if c.Expires != "" {
		if expires, err := time.Parse(time.RFC3339, c.Expires); err == nil {
			cookie.Expires = expires
		}
	}

	switch c.SameSite {
	case "none":
		cookie.SameSite = http.SameSiteNoneMode
	case "lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "strict":
		cookie.SameSite = http.SameSiteStrictMode
	}

	return cookie
}

func (cj *CookieJar) ClearCookies(domain string) {
	if cj.Cookies == nil {
		return
	}

	filtered := []Cookie{}
	for _, c := range cj.Cookies {
		if c.Domain != domain && !strings.HasSuffix(c.Domain, "."+domain) {
			filtered = append(filtered, c)
		}
	}

	cj.Cookies = filtered
}

func (cj *CookieJar) ClearAll() {
	cj.Cookies = []Cookie{}
}
