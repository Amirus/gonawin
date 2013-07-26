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

package controllers

import (
	"bytes"
	"html/template"
	"net/http"
	"time"

	"appengine"	

	"github.com/santiaago/purple-wing/helpers"
	usermdl "github.com/santiaago/purple-wing/models/user"
)

func UserShow(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	
	funcs := template.FuncMap{
		"LoggedIn": func() bool { return LoggedIn(r) },
	}
	
	t := template.Must(template.New("tmpl_user_show").
		Funcs(funcs).
		ParseFiles("templates/user/show.html", "templates/user/info.html"))
	
	user := usermdl.User{ 1, "test@example.com", "John Doe", nil, time.Now() }
	
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf,"tmpl_user_show", user)
	show := buf.Bytes()
	
	if err != nil{
		c.Errorf("pw: error in parse template user_show: %v", err)
	}

	err = helpers.Render(c, w, show, nil, "renderUserShow")
	if err != nil{
		c.Errorf("pw: error when calling Render from helpers: %v", err)
	}
}

func UserEdit(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)

	funcs := template.FuncMap{
		"LoggedIn": func() bool { return LoggedIn(r) },
	}
	
	t := template.Must(template.New("tmpl_user_show").
		Funcs(funcs).
		ParseFiles("templates/user/show.html", "templates/user/edit.html"))

	user := usermdl.User{ 1, "test@example.com", "John Doe", nil, time.Now() }

	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf,"tmpl_user_edit", user)
	edit := buf.Bytes()

	if err != nil{
		c.Errorf("pw: error in parse template user_edit: %v", err)
	}

	err = helpers.Render(c, w, edit, nil, "renderUserEdit")
	if err != nil{
		c.Errorf("pw: error when calling Render from helpers: %v", err)
	}
}
