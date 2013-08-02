/*
 * Copyright (c) 2013 Santiago Arias | Remy Jourde | Carlos Bernal
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package sessions

import (
	"appengine"
	"appengine/urlfetch"
	"net/http"
	"fmt"
	
	"code.google.com/p/goauth2/oauth"
	
	"github.com/santiaago/purple-wing/helpers/auth"
	usermdl "github.com/santiaago/purple-wing/models/user"
)

const root string = "/m"
// Set up a configuration.
func config(host string) *oauth.Config{
	return &oauth.Config{
		ClientId:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Scope:        "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  fmt.Sprintf("http://%s%s/oauth2callback", host, root),
	}
}

func SessionAuth(w http.ResponseWriter, r *http.Request){
	if !auth.LoggedIn(r) {
		url := config(r.Host).AuthCodeURL(r.URL.RawQuery)
		http.Redirect(w, r, url, http.StatusFound)
	} else {
		//redirect to home page
		http.Redirect(w, r, root, http.StatusFound)
	}
}

func SessionAuthCallback(w http.ResponseWriter, r *http.Request){
	// Exchange code for an access token at OAuth provider.
	code := r.FormValue("code")
	t := &oauth.Transport{
		Config: config(r.Host),
		Transport: &urlfetch.Transport{
			Context: appengine.NewContext(r),
		},
	}
	
	var userInfo *usermdl.GPlusUserInfo
	
	if _, err := t.Exchange(code); err == nil {
		userInfo, _ = usermdl.FetchUserInfo(r, t.Client())
	}
	if auth.IsAuthorized(userInfo) {
		var user *usermdl.User
		// find user
		if user = usermdl.Find(r, "Email", userInfo.Email); user == nil {
			// create user if it does not exist
			user = usermdl.Create(r, userInfo.Email, userInfo.Name, auth.GenerateAuthKey())
		}
		// set 'auth' cookie
		auth.SetAuthCookie(w, user.Auth)
		// store in memcache auth key in memcaches
		auth.StoreAuthKey(r, user.Id, user.Auth)
	}

	http.Redirect(w, r, root, http.StatusFound)
}

func SessionLogout(w http.ResponseWriter, r *http.Request){
	auth.ClearAuthCookie(w)
	
	http.Redirect(w, r, root, http.StatusFound)
}