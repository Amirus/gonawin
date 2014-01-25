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

package helpers

import (
	"time"
)

type UserJson struct{
	Id int64
	Username string
	Name string
	Email string
	IsAdmin bool
	Auth string 
	Created time.Time
	Teams []TeamJsonZip	
}

type UserJsonZip struct{
	Id int64
	Username string
}

// data structure for json API
type TeamJson struct{
	Id int64
	Name string
	Private bool
	Joined bool
	RequestSent bool
	AdminId int64
	Players []UserJsonZip
}

// data structure for json API
type TeamJsonZip struct{
	Id int64
	Name string
}
