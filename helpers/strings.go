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
	"strings"
)

// TrimLower returns a lower case slice of the string s, with all leading and trailing white space removed, as defined by Unicode.
func TrimLower(s string) string{
	return strings.TrimSpace(strings.ToLower(s))
}

func SetOfStrings(s string) []string{
	slice := strings.Split(TrimLower(s), " ")
	set := ""
	for _,w := range slice{
		if !StringContains(set, w){
			if len(set) == 0{
				set = w
			}else{
				set = set + " " + w
			}
		}
	}
	return strings.Split(set, " ")
}

func SliceContains(slice []string, s string)bool{
	for _, w := range slice{
		if w == s{
			return true
		}
	}
	return false
}


func StringContains(strToSplit string, s string)bool{
	slice := strings.Split(strToSplit, " ")
	for _, w := range slice{
		if w == s{
			return true
		}
	}
	return false
}

















