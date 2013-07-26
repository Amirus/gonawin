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

package pw

import (
	"net/http"

	"github.com/santiaago/purple-wing/controllers"
	"github.com/santiaago/purple-wing/helpers"
)

func init(){
	h := new(helpers.RegexpHandler)
	// usual pages
	h.HandleFunc("/", controllers.TempHome)
	h.HandleFunc("/m", controllers.Home)
	//h.HandleFunc("/", controllers.About)
	//h.HandleFunc("/", controllers.Contact)
	// session
	h.HandleFunc("/m/auth/?", controllers.SessionAuth)
	h.HandleFunc("/m/oauth2callback/?", controllers.SessionAuthCallback)
	h.HandleFunc("/m/logout/?", controllers.SessionLogout)	
	// user
	h.HandleFunc("/m/users/[0-9]+/?", controllers.UserShow)
	h.HandleFunc("/m/users/[0-9]+/edit/?", controllers.UserEdit)
	// admin
	h.HandleFunc("/m/a/?", controllers.AdminShow)
	h.HandleFunc("/m/a/users/?", controllers.AdminUsers)
	// team
	h.HandleFunc("/m/teams/?", controllers.TeamIndex)
	h.HandleFunc("/m/teams/[0-9]+/?", controllers.TeamShow)
	h.HandleFunc("/m/teams/[0-9]+/edit/?", controllers.TeamEdit)

	http.Handle("/", h)
}
