package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type State struct {
	Manifest string          `json:"manifest"`
	Options  json.RawMessage `json:"options"`
}

type States []State

func (s States) Json() []byte {
	a, _ := json.Marshal(s)
	return a
}

func Tify(state string) (States, error) {
	u, err := url.Parse(state)
	if err != nil {
		return nil, err
	}
	params := u.Query()
	var states States
	for i := 0; i < len(params)/2; i++ {
		manifestKey := "manifest"
		tifyKey := "tify"
		if i > 0 {
			manifestKey = fmt.Sprintf("manifest%d", i)
			tifyKey = fmt.Sprintf("tify%d", i)
		}
		manifest := params.Get(manifestKey)
		options := params.Get(tifyKey)

		iiifInfo := State{Manifest: manifest, Options: []byte(options)}
		states = append(states, iiifInfo)
	}
	return states, nil
}
