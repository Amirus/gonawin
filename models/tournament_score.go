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

package models

import (
	"encoding/json"
	"errors"
	"net/url"

	"appengine"
	"appengine/taskqueue"

	"github.com/santiaago/gonawin/helpers"
	"github.com/santiaago/gonawin/helpers/log"
)

// Update the score of the participants to the tournament.
func (t *Tournament) UpdateUsersScore(c appengine.Context, m *Tmatch) error {
	desc := "Update users score:"
	// send task to update scores of user.
	// ------------------------------------------------------------
	log.Infof(c, "%s Sending to taskqueue: update scores", desc)

	b1, errm := json.Marshal(t)
	if errm != nil {
		log.Errorf(c, "%s Error marshaling", desc, errm)
	}
	b2, errm2 := json.Marshal(m)
	if errm2 != nil {
		log.Errorf(c, "%s Error marshaling", desc, errm2)
	}

	task := taskqueue.NewPOSTTask("/a/update/scores/", url.Values{
		"tournament": []string{string(b1)},
		"match":      []string{string(b2)},
	})

	if _, err := taskqueue.Add(c, task, ""); err != nil {
		log.Errorf(c, "%s unable to add task to taskqueue.", desc)
		return err
	} else {
		log.Infof(c, "%s add task to taskqueue successfully", desc)
	}
	return nil
}

// Update the accuracy of the teams members in a specific tournament.
func (t *Tournament) UpdateTeamsAccuracy(c appengine.Context, m *Tmatch) error {
	desc := "Update Teams score:"
	teams := t.Teams(c)

	teamsToUpdate := make([]*Team, 0)
	for _, team := range teams {
		sumScore := int64(0)
		players := team.Players(c)
		if len(players) == 0 {
			// a team with 0 players? this should never happen, just skip to the next.
			continue
		}
		max := 3 * len(players) // maximum score for team in current match.
		for _, u := range players {
			if score, err := u.ScoreForMatch(c, m); err != nil {
				log.Errorf(c, "%s unable udpate user %v score: %v", desc, u.Id, err)
			} else {
				sumScore += score
			}
		}

		// compute current accuracy, get accuracy entity , add accuracy to entity.
		log.Infof(c, "sum of score is: %v", sumScore)
		log.Infof(c, "max: %v", max)
		newAcc := float64(sumScore) / float64(max)
		log.Infof(c, "new Acc: %v", newAcc)
		computedAcc := float64(0)
		if acc, _ := team.TournamentAcc(c, t); acc == nil {
			log.Infof(c, "%s create accuracy if not exist", desc)
			oldmatches := t.OldMatches(c)
			if oldmatches > 0 {
				oldmatches = oldmatches - 1 // do not take into account the match that triggers the update accuracy.
			}
			if acc1, err := CreateAccuracy(c, team.Id, t.Id, oldmatches); err != nil {
				log.Errorf(c, "%s unable to create accuracy", desc)
				return err
			} else {
				team.AddTournamentAcc(c, acc1.Id, t.Id)
				log.Infof(c, "%s accuracy exists now, lets update it", desc)
				var err error
				if computedAcc, err = acc1.Add(c, newAcc); err != nil {
					log.Errorf(c, "%s unable to add accuracy of team %v, ", desc, team.Id, err)
				}
			}
		} else {
			log.Infof(c, "%s accuracy entity exists, lets update it", desc)
			var err error
			if computedAcc, err = acc.Add(c, newAcc); err != nil {
				log.Errorf(c, "%s unable to add accuracy of team %v, ", desc, team.Id, err)
			}
		}

		// ToDo: update team overall accuracy.
		log.Infof(c, "%s ready to update global accuracy for team: %v", desc, team.Id)
		if err := team.UpdateAccuracy(c, t.Id, computedAcc); err != nil {
			log.Errorf(c, "%s unable to update global accuracy for team: %v, %v", desc, team.Id, err)
		} else {
			log.Infof(c, "%s update successfull: %v", desc, team.Id)
		}
	}
	if err := UpdateTeams(c, teamsToUpdate); err != nil {
		log.Errorf(c, "%s unable udpate teams scores: %v", desc, err)
		return errors.New(helpers.ErrorCodeTeamsCannotUpdate)
	}

	return nil
}

// Computes the score to be given with respect to a match and a predict.
func computeScore(c appengine.Context, m *Tmatch, p *Predict) int64 {
	// exact result
	if (m.Result1 == p.Result1) && (m.Result2 == p.Result2) {
		return int64(3)
	}
	// wining trend
	trendW := (m.Result1 > m.Result2)
	ptrendW := (p.Result1 > p.Result2)
	if (trendW == ptrendW) && (trendW == true) {
		return int64(1)
	}
	// losign trend
	trendL := (m.Result1 < m.Result2)
	ptrendL := (p.Result1 < p.Result2)
	if trendL && ptrendL {
		return int64(1)
	}
	// tied trend
	trendT := (m.Result1 == m.Result2)
	ptrendT := (p.Result1 == p.Result2)
	if trendT && ptrendT {
		return int64(1)
	}
	// bad predict
	return int64(0)
}
