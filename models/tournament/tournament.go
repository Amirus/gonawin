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

package tournament

import (
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/santiaago/purple-wing/helpers"
	tournamentinvidmdl "github.com/santiaago/purple-wing/models/tournamentInvertedIndex"
	tournamentrelmdl "github.com/santiaago/purple-wing/models/tournamentrel"
)

type Tournament struct {
	Id int64
	KeyName string
	Name string
	Description string
	Start time.Time
	End time.Time
	Created time.Time
}

type CounterTournament struct {
	Count int64
}


func Create(r *http.Request, name string, description string, start time.Time, end time.Time ) *Tournament {
	c := appengine.NewContext(r)
	// create new tournament
	tournamentID, _, _ := datastore.AllocateIDs(c, "Tournament", nil, 1)
	key := datastore.NewKey(c, "Tournament", "", tournamentID, nil)

	tournament := &Tournament{ tournamentID, helpers.TrimLower(name), name, description, start, end, time.Now() }

	_, err := datastore.Put(c, key, tournament)
	if err != nil {
		c.Errorf("Create: %v", err)
	}

	tournamentinvidmdl.Add(r, helpers.TrimLower(name),tournamentID)
	return tournament
}

func Find(r *http.Request, filter string, value interface{}) *Tournament {
	q := datastore.NewQuery("Tournament").Filter(filter + " =", value).Limit(1)

	var tournaments []*Tournament

	if _, err := q.GetAll(appengine.NewContext(r), &tournaments); err == nil && len(tournaments) > 0 {
		return tournaments[0]
	}

	return nil
}

func ById(r *http.Request, id int64)(*Tournament, error){
	c := appengine.NewContext(r)

	var t Tournament
	key := datastore.NewKey(c, "Tournament", "", id, nil)

	if err := datastore.Get(c, key, &t); err != nil {
		c.Errorf("pw: tournament not found : %v", err)
		return &t, err
	}
	return &t, nil
}


func KeyById(r *http.Request, id int64)(*datastore.Key){
	c := appengine.NewContext(r)

	key := datastore.NewKey(c, "Tournament", "", id, nil)

	return key
}


func Update(r *http.Request, id int64, t *Tournament) error{
	c := appengine.NewContext(r)
	k := KeyById(r, id)
	oldTournament := new(Tournament)
	if err := datastore.Get(c, k, oldTournament); err == nil{
		if _, err = datastore.Put(c, k, t); err != nil {
			return err
		}
		tournamentinvidmdl.Update(r, oldTournament.Name, t.Name, id)
	}
	return nil
}

func FindAll(r *http.Request) []*Tournament {
	q := datastore.NewQuery("Tournament")

	var tournaments []*Tournament

	q.GetAll(appengine.NewContext(r), &tournaments)

	return tournaments
}

func Joined(r *http.Request, tournamentId int64, userId int64) bool {
	tournamentRel := tournamentrelmdl.FindByTournamentIdAndUserId(r, tournamentId, userId)
	return tournamentRel != nil
}

func Join(r *http.Request, tournamentId int64, userId int64) error {
	if tournamentRel := tournamentrelmdl.Create(r, tournamentId, userId); tournamentRel == nil {
		return errors.New("error during tournament relationship creation")
	}

	return nil
}

func Leave(r *http.Request, tournamentId int64, userId int64) error {
	return tournamentrelmdl.Destroy(r, tournamentId, userId)
}
