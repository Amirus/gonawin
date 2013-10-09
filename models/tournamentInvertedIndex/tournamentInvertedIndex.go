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

package tournamentinvid

import (
	"net/http"
	"strings"
	"strconv"

	"appengine"
	"appengine/datastore"

	"github.com/santiaago/purple-wing/helpers"
)

type TournamentInvertedIndex struct {
	Id int64
	KeyName string
	TournamentIds []byte
}

type WordCountTournament struct{
	Count int64
}

func Create(r *http.Request, name string, tournamentIds string) *TournamentInvertedIndex {
	c := appengine.NewContext(r)
	
	id, _, _ := datastore.AllocateIDs(c, "TournamentInvertedIndex", nil, 1)
	key := datastore.NewKey(c, "TournamentInvertedIndex", "", id, nil)
	
	byteIds := []byte(tournamentIds)
	t := &TournamentInvertedIndex{ id, helpers.TrimLower(name), byteIds }

	_, err := datastore.Put(c, key, t)
	if err != nil {
		c.Errorf("Create: %v", err)
	}

	return t
}

// AddToTournamentInvertedIndex
// Split name by words.
// For each word check if it exist in the Tournament Inverted Index. 
// if not create a line with the word as key and team id as value.
func Add(r *http.Request, name string, id int64){

	c := appengine.NewContext(r)
	c.Infof("pw: AddToTournamentInvertedIndex start")
	words := strings.Split(name, " ")
	for _, w:= range words{
		c.Infof("pw: AddToTournamentInvertedIndex: Word: %v", w)

		if inv_id := Find(r, "KeyName", w); inv_id == nil{
			c.Infof("pw: create inv id as word does not exist in table")
			Create(r, w, strconv.FormatInt(id, 10))
		} else{
			// update row with new info
			c.Infof("pw: update row with new info")
			c.Infof("pw: current info: keyname: %v", inv_id.KeyName)
			c.Infof("pw: current info: teamIDs: %v", string(inv_id.TournamentIds))
			k := KeyById(r, inv_id.Id)

			if newIds := helpers.MergeIds(inv_id.TournamentIds, id);len(newIds) > 0{
				c.Infof("pw: current info: new team ids: %v", newIds)
				inv_id.TournamentIds  = []byte(newIds)
				if _, err := datastore.Put(c, k, inv_id);err != nil{
					c.Errorf("pw: AddToTournamentInvertedIndex error on update")
				}
			}
		}
	}
	c.Infof("pw: AddToTournamentInvertedIndex end")	
}

// if the removal of the id makes the entity useless (no more ids in it)
// we will remove the entity as well
func RemoveWord(r *http.Request, w string, id int64){

	c := appengine.NewContext(r)
	c.Infof("pw: RemoveWordFromTournamentInvertedIndex start")

	if inv_id := Find(r, "KeyName", w); inv_id == nil{
		c.Infof("pw: word %v does not exist in Tournament InvertedIndex so nothing to remove",w)
	} else{
		// update row with new info
		k := KeyById(r, inv_id.Id)

		if newIds, err := helpers.RemovefromIds(inv_id.TournamentIds, id); err == nil{
			c.Infof("pw: new tournament ids after removal: %v", newIds)
			if len(newIds) == 0{
				// this entity does not have ids so remove it from the datastore.
				c.Infof("pw: removing key %v from datastore as it is no longer used", k)
				datastore.Delete(c, k)
			}else{
				inv_id.TournamentIds  = []byte(newIds)
				if _, err := datastore.Put(c, k, inv_id);err != nil{
					c.Errorf("pw: RemoveWordFromTournamentInvertedIndex error on update")
				}
			}
		}else{
			c.Errorf("pw: unable to remove id from ids: %v",err)
		}
	}
	c.Infof("pw: RemoveWordFromTournamentInvertedIndex end")
	
}
// add word to tournament inverted index is handled the same way as AddToTournamentInvertedIndex.
// we add this function for clarity
func AddWord(r *http.Request, word string, id int64){
	Add(r, word, id)
}

func Update(r *http.Request, oldname string, newname string, id int64){
	c := appengine.NewContext(r)
	c.Infof("pw: Update start")

	// if word in old and new do nothing
	old_w := strings.Split(oldname," ")
	new_w := strings.Split(newname, " ")

	// remove id  from words in old name that are not present in new name
	for _, wo := range old_w{
		innew := false
		for _, wn := range new_w{
			if wo == wn{
				innew = true
			}
		}
		if !innew{
			c.Infof("pw: remove: %v",wo)
			RemoveWord(r, wo, id)
			//remove it
		}
	}

	// add all id words in new name
	for _, wn := range new_w{
		inold := false
		for _, wo := range old_w{
			if wo == wn{
				inold = true
			}
		}
		if !inold{
			// add it
			c.Infof("pw: add: %v", wn)
			AddWord(r, wn, id)
		}
	}

	c.Infof("pw: AddToTournamentInvertedIndex end")	
}

func Find(r *http.Request, filter string, value interface{}) *TournamentInvertedIndex {
	q := datastore.NewQuery("TournamentInvertedIndex").Filter(filter + " =", value).Limit(1)
	
	var t[]*TournamentInvertedIndex
	
	if _, err := q.GetAll(appengine.NewContext(r), &t); err == nil && len(t) > 0 {
		return t[0]
	}
	
	return nil
}

func KeyById(r *http.Request, id int64) (*datastore.Key) {
	c := appengine.NewContext(r)

	key := datastore.NewKey(c, "TournamentInvertedIndex", "", id, nil)

	return key
}