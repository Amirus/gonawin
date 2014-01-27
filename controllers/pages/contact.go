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

package pages

import (
	"html/template"
	"net/http"

	"appengine"

	"github.com/santiaago/purple-wing/helpers/auth"
	templateshlp "github.com/santiaago/purple-wing/helpers/templates"
)

// contact handler: for contact page
func Contact(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	data := data{
		auth.CurrentUser(r, c),
		"Contact handler",
	}

	funcs := template.FuncMap{}

	t := template.Must(template.New("tmpl_contact").
		Funcs(funcs).
		ParseFiles("templates/pages/contact.html"))

	templateshlp.RenderWithData(w, r, c, t, data, funcs, "renderContact")
}

// json contact handler: for contact page
func ContactJson(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	data := data{
		auth.CurrentUser(r, c),
		"Contact handler",
	}
	return templateshlp.RenderJson(w, c, data)
}
