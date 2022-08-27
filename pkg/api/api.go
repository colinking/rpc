package api

import jtd "github.com/jsontypedef/json-typedef-go"

type API struct {
	Endpoints   []Endpoint
	Definitions []Definition
}

type Definition struct {
	Name   []string
	Schema jtd.Schema
}

type Endpoint struct {
	Name     []string
	Verb     string
	Request  jtd.Schema
	Response jtd.Schema
}
