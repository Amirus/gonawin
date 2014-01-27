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

package users

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"appengine"

	"github.com/santiaago/purple-wing/helpers"
	"github.com/santiaago/purple-wing/helpers/handlers"
	"github.com/santiaago/purple-wing/helpers/log"
	teamrelshlp "github.com/santiaago/purple-wing/helpers/teamrels"
	templateshlp "github.com/santiaago/purple-wing/helpers/templates"
	tournamentrelshlp "github.com/santiaago/purple-wing/helpers/tournamentrels"

	teammdl "github.com/santiaago/purple-wing/models/team"
	teamrequestmdl "github.com/santiaago/purple-wing/models/teamrequest"
	tournamentmdl "github.com/santiaago/purple-wing/models/tournament"
	usermdl "github.com/santiaago/purple-wing/models/user"
)

type Form struct {
	Username      string
	Name          string
	Email         string
	ErrorUsername string
	ErrorName     string
	ErrorEmail    string
}

type UserData struct {
	Username string
	Name     string
	Email    string
}

//used by json api to send only needed info
type userJsonZip struct {
	Id       int64
	Username string
	Name     string
	Email    string
	Created  time.Time
}

type userJson struct {
	Id           int64
	Username     string
	Name         string
	Email        string
	IsAdmin      bool
	Auth         string
	Created      time.Time
	Teams        []teamJsonZip
	Tournaments  []tournamentJsonZip
	TeamRequests []teamRequestJsonZip
}

type teamJsonZip struct {
	Id   int64
	Name string
}

type tournamentJsonZip struct {
	Id   int64
	Name string
}

type teamRequestJsonZip struct {
	Id     int64
	TeamId int64
	UserId int64
}

// Show handler
func Show(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	userId, err := handlers.PermalinkID(r, c, 3)
	if err != nil {
		http.Redirect(w, r, "/m/users/", http.StatusFound)
		return
	}

	funcs := template.FuncMap{
		"Profile": func() bool { return true },
	}

	t := template.Must(template.New("tmpl_user_show").
		Funcs(funcs).
		ParseFiles("templates/user/show.html",
		"templates/user/info.html",
		"templates/user/teams.html",
		"templates/user/tournaments.html",
		"templates/user/requests.html"))

	var user *usermdl.User
	user, err = usermdl.ById(c, userId)
	if err != nil {
		helpers.Error404(w)
		return
	}

	teams := usermdl.Teams(c, userId)
	tournaments := tournamentrelshlp.Tournaments(c, userId)
	teamRequests := teamrelshlp.TeamsRequests(c, teams)

	userData := struct {
		User         *usermdl.User
		Teams        []*teammdl.Team
		Tournaments  []*tournamentmdl.Tournament
		TeamRequests []*teamrequestmdl.TeamRequest
	}{
		user,
		teams,
		tournaments,
		teamRequests,
	}

	templateshlp.RenderWithData(w, r, c, t, userData, funcs, "renderUserShow")
}

// json index user handler
func IndexJson(w http.ResponseWriter, r *http.Request, u *usermdl.User) error {
	c := appengine.NewContext(r)

	if r.Method == "GET" {
		users := usermdl.FindAll(c)
		usersJson := make([]userJson, len(users))
		counterUsers := 0
		for _, user := range users {
			usersJson[counterUsers].Id = user.Id
			usersJson[counterUsers].Username = user.Username
			usersJson[counterUsers].Name = user.Name
			usersJson[counterUsers].Email = user.Email
			usersJson[counterUsers].Created = user.Created
			counterUsers++
		}
		return templateshlp.RenderJson(w, c, usersJson)

	} else {
		return helpers.BadRequest{errors.New("not supported.")}
	}
}

// Json show user handler
func ShowJson(w http.ResponseWriter, r *http.Request, u *usermdl.User) error {
	c := appengine.NewContext(r)
	log.Infof(c, "User Show Json Handler")

	if r.Method == "GET" {

		userId, err := handlers.PermalinkID(r, c, 4)
		if err != nil {
			return helpers.BadRequest{err}
		}

		// get user
		var user *usermdl.User
		user, err = usermdl.ById(c, userId)
		if err != nil {
			return helpers.BadRequest{err}
		}

		// set teams info in user json
		teams := usermdl.Teams(c, userId)
		teamsJson := make([]teamJsonZip, len(teams))
		counterTeams := 0
		for _, team := range teams {
			teamsJson[counterTeams].Id = team.Id
			teamsJson[counterTeams].Name = team.Name
			counterTeams++
		}

		// set tournament info in user json
		tournaments := tournamentrelshlp.Tournaments(c, userId)
		tournamentsJson := make([]tournamentJsonZip, len(tournaments))
		counterTournaments := 0
		for _, tournament := range tournaments {
			tournamentsJson[counterTournaments].Id = tournament.Id
			tournamentsJson[counterTournaments].Name = tournament.Name
			counterTournaments++
		}

		// set request info in user json
		teamRequests := teamrelshlp.TeamsRequests(c, teams)
		teamRequestsJson := make([]teamRequestJsonZip, len(teamRequests))
		counterRequests := 0
		for _, request := range teamRequests {
			teamRequestsJson[counterRequests].Id = request.Id
			teamRequestsJson[counterRequests].UserId = request.UserId
			teamRequestsJson[counterRequests].TeamId = request.TeamId
			counterRequests++
		}

		// copy to json data
		var userJson userJson
		userJson.Id = user.Id
		userJson.Username = user.Username
		userJson.Name = user.Name
		userJson.Email = user.Email
		userJson.IsAdmin = user.IsAdmin
		userJson.Auth = user.Auth
		userJson.Created = user.Created
		userJson.Teams = teamsJson
		userJson.TeamRequests = teamRequestsJson
		userJson.Tournaments = tournamentsJson

		return templateshlp.RenderJson(w, c, userJson)
	} else {
		return helpers.BadRequest{errors.New("not supported.")}
	}
}

// json update user handler
func UpdateJson(w http.ResponseWriter, r *http.Request, u *usermdl.User) error {
	c := appengine.NewContext(r)

	if r.Method == "POST" {
		userId, err := handlers.PermalinkID(r, c, 4)

		if err != nil {
			return helpers.BadRequest{err}
		}
		if userId != u.Id {
			return helpers.BadRequest{errors.New("User cannot be updated")}
		}

		// only work on name other values should not be editable
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return helpers.InternalServerError{errors.New("Error when reading request body")}
		}

		var updatedData UserData
		err = json.Unmarshal(body, &updatedData)
		if err != nil {
			return helpers.InternalServerError{errors.New("Error when decoding request body")}
		}
		if helpers.IsEmailValid(updatedData.Email) && updatedData.Email != u.Email {
			u.Email = updatedData.Email
			usermdl.Update(c, u)
		}

		// return updated user
		return templateshlp.RenderJson(w, c, u)
	} else {
		return helpers.BadRequest{errors.New("not supported.")}
	}
}
