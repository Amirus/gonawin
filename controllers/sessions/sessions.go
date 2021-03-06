/*
 * Copyright (c) 2014 Santiago Arias | Remy Jourde
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

// Package sessions provides the JSON handlers to handle connections to gonawin app.
package sessions

import (
	"errors"
	golog "log"
	"net/http"
	"net/url"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"appengine/user"

	oauth "github.com/garyburd/go-oauth/oauth"

	"github.com/santiaago/gonawin/helpers"
	authhlp "github.com/santiaago/gonawin/helpers/auth"
	"github.com/santiaago/gonawin/helpers/log"
	"github.com/santiaago/gonawin/helpers/memcache"
	templateshlp "github.com/santiaago/gonawin/helpers/templates"

	gwconfig "github.com/santiaago/gonawin/config"
	mdl "github.com/santiaago/gonawin/models"
)

var (
	config                 *gwconfig.GwConfig
	twitterConfig          oauth.Client
	twitterCallbackURL     string
	googleVerifyTokenURL   string
	facebookVerifyTokenURL string
)

func init() {
	// read config file.
	var err error
	if config, err = gwconfig.ReadConfig(""); err != nil {
		golog.Printf("Error: unable to read config file; %v", err)
	}
	// Set up a configuration for twitter.
	twitterConfig = oauth.Client{
		Credentials:                   oauth.Credentials{Token: config.Twitter.Token, Secret: config.Twitter.Secret},
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
	}
	twitterCallbackURL = "/j/auth/twitter/callback"
	googleVerifyTokenURL = "https://www.google.com/accounts/AuthSubTokenInfo?bearer_token"
	facebookVerifyTokenURL = "https://graph.facebook.com/me?access_token"
}

// JSON authentication handler
func Authenticate(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	if r.Method == "GET" {
		userInfo := authhlp.UserInfo{Id: r.FormValue("id"), Email: r.FormValue("email"), Name: r.FormValue("name")}

		var verifyURL string
		if r.FormValue("provider") == "google" {
			verifyURL = googleVerifyTokenURL
		} else if r.FormValue("provider") == "facebook" {
			verifyURL = facebookVerifyTokenURL
		}

		if !authhlp.CheckUserValidity(r, verifyURL, r.FormValue("access_token")) {
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsAccessTokenNotValid)}
		}

		var user *mdl.User
		var err error
		if user, err = mdl.SigninUser(w, r, "Email", userInfo.Email, userInfo.Name, userInfo.Name); err != nil {
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsUnableToSignin)}
		}

		// return user
		userData := struct {
			User *mdl.User
		}{
			user,
		}

		return templateshlp.RenderJson(w, c, userData)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// JSON authentication for Twitter.
func TwitterAuth(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	desc := "Twitter Auth handler:"
	if r.Method == "GET" {
		credentials, err := twitterConfig.RequestTemporaryCredentials(urlfetch.Client(c), "http://"+r.Host+twitterCallbackURL, nil)
		if err != nil {
			c.Errorf("JsonTwitterAuth, error = %v", err)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetTempCredentials)}
		}

		if err = memcache.Set(c, "secret", credentials.Secret); err != nil {
			// store secret in datastore
			secretId, _, err := datastore.AllocateIDs(c, "Secret", nil, 1)
			if err != nil {
				log.Errorf(c, "%s Cannot allocate ID for secret. %v", desc, err)
			}

			key := datastore.NewKey(c, "Secret", "", secretId, nil)

			_, err = datastore.Put(c, key, credentials.Secret)
			if err != nil {
				log.Errorf(c, "%s Cannot put secret in Datastore. %v", desc, err)
				return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotSetSecretValue)}
			}
		}

		// return OAuth token
		oAuthToken := struct {
			OAuthToken string
		}{
			credentials.Token,
		}

		return templateshlp.RenderJson(w, c, oAuthToken)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// Twitter Authentication Callback
func TwitterAuthCallback(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		http.Redirect(w, r, "http://"+r.Host+"/#/auth/twitter/callback?oauth_token="+r.FormValue("oauth_token")+"&oauth_verifier="+r.FormValue("oauth_verifier"), http.StatusFound)
		return nil
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// Twitter Authentication Callback
func TwitterUser(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	desc := "Twitter User handler:"
	if r.Method == "GET" {

		var user *mdl.User

		log.Infof(c, "%s oauth_verifier = %s", desc, r.FormValue("oauth_verifier"))
		log.Infof(c, "%s oauth_token = %s", desc, r.FormValue("oauth_token"))

		// get the request token
		requestToken := r.FormValue("oauth_token")
		// update credentials with request token and 'secret cookie value'
		var cred oauth.Credentials
		cred.Token = requestToken
		if secret, err := memcache.Get(c, "secret"); secret != nil {
			cred.Secret = string(secret.([]byte))
		} else {
			log.Errorf(c, "%s cannot get secret value from memcache: %v", desc, err)
			// try to get secret from datastore
			q := datastore.NewQuery("Secret")
			var secrets []string
			if keys, err := q.GetAll(c, &secrets); err == nil && len(secrets) > 0 {
				secret = secrets[0]

				// delete secret from datastore
				if err = datastore.Delete(c, keys[0]); err != nil {
					log.Errorf(c, "%s Error when trying to delete 'secret' key in Datastore: %v", desc, err)
				}

			} else if err != nil || len(secrets) == 0 {
				log.Errorf(c, "%s cannot get secret value from Datastore: %v", desc, err)
				return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetSecretValue)}
			}
		}

		if err := memcache.Delete(c, "secret"); err != nil {
			log.Errorf(c, "%s Error when trying to delete memcached 'secret' key: %v", desc, err)
		}

		token, values, err := twitterConfig.RequestToken(urlfetch.Client(c), &cred, r.FormValue("oauth_verifier"))
		if err != nil {
			log.Errorf(c, "%s Error when trying to delete memcached 'secret' key: %v", desc, err)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetRequestToken)}
		}

		// get user info
		urlValues := url.Values{}
		urlValues.Set("user_id", values.Get("user_id"))
		resp, err := twitterConfig.Get(urlfetch.Client(c), token, "https://api.twitter.com/1.1/users/show.json", urlValues)
		if err != nil {
			log.Errorf(c, "%s Cannot get user info from twitter. %v", desc, err)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetUserInfo)}
		}

		userInfo, err := authhlp.FetchTwitterUserInfo(resp)
		if err != nil {
			log.Errorf(c, "%s Cannot get user info by fetching twitter response. %v", desc, err)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetUserInfo)}
		}

		if user, err = mdl.SigninUser(w, r, "Username", "", userInfo.Screen_name, userInfo.Name); err != nil {
			log.Errorf(c, "%s Unable to signin user %s. %v", desc, userInfo.Name, err)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsUnableToSignin)}
		}

		// return user
		userData := struct {
			User *mdl.User
		}{
			user,
		}

		return templateshlp.RenderJson(w, c, userData)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// JSON handler to get Google accounts login URL.
func GoogleAccountsLoginURL(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	desc := "Google Accounts Login URL Handler:"
	if r.Method == "GET" {
		var url string
		var err error
		url, err = user.LoginURL(c, "/j/auth/google/callback/")
		if err != nil {
			log.Errorf(c, "%s error when getting Google accounts login URL", desc)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsCannotGetGoogleLoginUrl)}
		}

		// return user
		loginData := struct {
			Url string
		}{
			url,
		}

		return templateshlp.RenderJson(w, c, loginData)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// Google Authentication Callback
func GoogleAuthCallback(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	desc := "Google Accounts Auth Callback Handler:"
	if r.Method == "GET" {
		u := user.Current(c)
		if u == nil {
			log.Errorf(c, "%s user cannot be nil", desc)
			return &helpers.InternalServerError{Err: errors.New("user cannot be nil")}
		}

		http.Redirect(w, r, "http://"+r.Host+"/#/auth/google/callback?auth_token="+u.ID, http.StatusFound)
		return nil
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// JSON handler to get Google accounts user.
func GoogleUser(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	desc := "Google Accounts User Handler:"
	if r.Method == "GET" {
		u := user.Current(c)
		if u == nil {
			log.Errorf(c, "%s user cannot be nil", desc)
			return &helpers.InternalServerError{Err: errors.New("user cannot be nil")}
		}

		if u.ID != r.FormValue("auth_token") {
			log.Errorf(c, "%s Auth token doesn't match user ID", desc)
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsAccessTokenNotValid)}
		}

		userInfo := authhlp.GetUserGoogleInfo(u)

		var user *mdl.User
		var err error
		if user, err = mdl.SigninUser(w, r, "Email", userInfo.Email, userInfo.Name, userInfo.Name); err != nil {
			return &helpers.InternalServerError{Err: errors.New(helpers.ErrorCodeSessionsUnableToSignin)}
		}

		// return user
		userData := struct {
			User *mdl.User
		}{
			user,
		}

		return templateshlp.RenderJson(w, c, userData)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

// JSON handler to delete cookie created by Google account
func GoogleDeleteCookie(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	if r.Method == "GET" {
		cookieName := "ACSID"
		if appengine.IsDevAppServer() {
			cookieName = "dev_appserver_login"
		}
		cookie := http.Cookie{Name: cookieName, Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)

		return templateshlp.RenderJson(w, c, "Google user has been logged out")
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}

func AuthServiceIds(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	if r.Method == "GET" {
		data := struct {
			GooglePlusClientId string
			FacebookAppId      string
		}{
			config.GooglePlus.ClientId,
			config.Facebook.AppId,
		}
		return templateshlp.RenderJson(w, c, data)
	}
	return &helpers.BadRequest{Err: errors.New(helpers.ErrorCodeNotSupported)}
}
