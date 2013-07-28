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
	"net/http"
	"html/template"
	"bytes"
	"fmt"
	
	"appengine"
	
	usermdl "github.com/santiaago/purple-wing/models/user"
	"github.com/santiaago/purple-wing/helpers"
)

// Data struct holds the data for templates
type data struct{
	User *usermdl.User
	Msg string
}

//temporary main handler: for landing page
func TempHome(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w, "Hello, purple wing!")
}

//main handler: for home page
func Home(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)

	data := data{
		helpers.CurrentUser(r),		
		"Home handler",
	}

	userdata := helpers.UserData{helpers.CurrentUser(r),}
	
	funcs := template.FuncMap{
		"LoggedIn": func() bool { return LoggedIn(r) },
		"Home": func() bool {return true},
	}
	
	t := template.Must(template.New("tmpl_main").
		Funcs(funcs).
		ParseFiles("templates/pages/main.html"))
	
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf,"tmpl_main", data)
	main := buf.Bytes()
	
	if err != nil{
		c.Errorf("pw: error executing template  main: %v", err)
	}
	err = helpers.Render(c, w, main, funcs, userdata, "renderMain")
	
	if err != nil{
		c.Errorf("pw: error when calling Render from helpers in Home Handler: %v", err)
	}

}

//about handler: for about page
func About(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)

	data := data{
		helpers.CurrentUser(r),		
		"About handler",
	}

	userdata := helpers.UserData{helpers.CurrentUser(r),}
	
	funcs := template.FuncMap{
		"LoggedIn": func() bool { return LoggedIn(r) },
		"About": func() bool {return true},
	}
	
	t := template.Must(template.New("tmpl_about").
		Funcs(funcs).
		ParseFiles("templates/pages/about.html"))
	
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf,"tmpl_about", data)
	main := buf.Bytes()
	
	if err != nil{
		c.Errorf("pw: error executing template  about: %v", err)
	}
	err = helpers.Render(c, w, main, funcs, userdata, "renderAbout")
	
	if err != nil{
		c.Errorf("pw: error when calling Render from helpers in About Handler: %v", err)
	}
}


//contact handler: for contact page
func Contact(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)

	data := data{
		helpers.CurrentUser(r),		
		"Contact handler",
	}

	userdata := helpers.UserData{
		helpers.CurrentUser(r),
	}
	
	funcs := template.FuncMap{
		"LoggedIn": func() bool { return LoggedIn(r) },
		"Contact": func() bool {return true},
	}
	
	t := template.Must(template.New("tmpl_contact").
		Funcs(funcs).
		ParseFiles("templates/pages/contact.html"))
	
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf,"tmpl_contact", data)
	main := buf.Bytes()
	
	if err != nil{
		c.Errorf("pw: error executing template  contact: %v", err)
	}
	err = helpers.Render(c, w, main, funcs, userdata, "renderContact")
	
	if err != nil{
		c.Errorf("pw: error when calling Render from helpers in Contact Handler: %v", err)
	}
}