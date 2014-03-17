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

package models

import (
	"errors"
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"

	"github.com/santiaago/purple-wing/helpers/log"
)

// A Predict entity is defined the result of a Match: Result1 and Result2 and a MatchId that references a Match entity in the datastore.
type Predict struct {
	Id      int64
	Result1 int64
	Result2 int64
	MatchId int64
	Created time.Time
}

// Create a Predict given a name, a result and a match id admin id and a private mode.
func CreatePredict(c appengine.Context, result1 int64, result2 int64, matchId int64) (*Predict, error) {

	pId, _, err := datastore.AllocateIDs(c, "Predict", nil, 1)
	if err != nil {
		return nil, err
	}
	key := datastore.NewKey(c, "Predict", "", pId, nil)
	p := &Predict{pId, result1, result2, matchId, time.Now()}
	if _, err = datastore.Put(c, key, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Destroy a Predict entity.
func (p *Predict) Destroy(c appengine.Context) error {

	if _, err := PredictById(c, p.Id); err != nil {
		return errors.New(fmt.Sprintf("Cannot find predict with Id=%d", p.Id))
	} else {
		key := datastore.NewKey(c, "Predict", "", p.Id, nil)

		return datastore.Delete(c, key)
	}
}

// Search for an array of Predict entities given a filter and a value.
func FindPredicts(c appengine.Context, filter string, value interface{}) []*Predict {

	q := datastore.NewQuery("Predict").Filter(filter+" =", value)

	var predicts []*Predict

	if _, err := q.GetAll(c, &predicts); err == nil {
		return predicts
	} else {
		log.Errorf(c, " Predict.Find, error occurred during GetAll: %v", err)
		return nil
	}
}

// Get a Predict given an id.
func PredictById(c appengine.Context, id int64) (*Predict, error) {

	var p Predict
	key := datastore.NewKey(c, "Predict", "", id, nil)

	if err := datastore.Get(c, key, &p); err != nil {
		log.Errorf(c, "predict not found : %v", err)
		return &p, err
	}
	return &p, nil
}

// Get a Predict key given an id.
func PredictKeyById(c appengine.Context, id int64) *datastore.Key {

	key := datastore.NewKey(c, "Predict", "", id, nil)

	return key
}

// Update a Predict entity.
func (p *Predict) Update(c appengine.Context) error {
	k := PredictKeyById(c, p.Id)
	old := new(Predict)
	if err := datastore.Get(c, k, old); err == nil {
		if _, err = datastore.Put(c, k, p); err != nil {
			return err
		}
	}
	return nil
}

// Get all Predicts in datastore.
func FindAllPredicts(c appengine.Context) []*Predict {
	q := datastore.NewQuery("Predict")

	var predicts []*Predict

	if _, err := q.GetAll(c, &predicts); err != nil {
		log.Errorf(c, " Predict.FindAll, error occurred during GetAll call: %v", err)
	}
	return predicts
}

// Get an array of pointers to Predict entities with respect to an array of ids.
func PredictsByIds(c appengine.Context, ids []int64) []*Predict {

	var predicts []*Predict
	for _, id := range ids {
		if p, err := PredictById(c, id); err == nil {
			predicts = append(predicts, p)
		} else {
			log.Errorf(c, " Predict.ByIds, error occurred during ByIds call: %v", err)
		}
	}
	return predicts
}

// Get an array of pointers to Predict entities with respect to an array of ids.
// Trying out with a get batch instead of going through each entity.
func PredictsByIds2(c appengine.Context, ids []int64) ([]*Predict, error) {
	keys := make([]*datastore.Key, len(ids))
	for i, _ := range keys {
		keys[i] = PredictKeyById(c, ids[i])
	}
	var predicts []*Predict
	if err := datastore.GetMulti(c, keys, predicts); err != nil {
		if me, ok := err.(appengine.MultiError); ok {
			for i, merr := range me {
				if merr == datastore.ErrNoSuchEntity {
					log.Errorf(c, "Predict.ByIds2, key[%v] is missing: %v", i)

				}
			}
		} else {
			return nil, err
		}
	}
	return predicts, nil
}

type Predicts []*Predict

func (a Predicts) ContainsMatchId(id int64) (bool, int) {
	for i, e := range a {
		if e.MatchId == id {
			return true, i
		}
	}
	return false, -1
}
