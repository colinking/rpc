package schema

import (
	"testing"

	jtd "github.com/jsontypedef/json-typedef-go"
	"github.com/stretchr/testify/require"
)

func TestDiscover(t *testing.T) {
	require := require.New(t)

	api, err := Discover("./fixtures/simple")
	require.NoError(err)
	require.Equal(API{
		Definitions: []Definition{
			{
				Path: []string{"users", "id"},
				Schema: jtd.Schema{
					Type: jtd.TypeString,
				},
			},
		},
		Endpoints: []Endpoint{
			{
				Path: []string{"users", "get"},
				Verb: "GET",
				Request: jtd.Schema{
					Properties: map[string]jtd.Schema{
						"id": {
							Type: jtd.TypeString,
						},
					},
				},
				Response: jtd.Schema{
					Properties: map[string]jtd.Schema{
						"name": {
							Type: jtd.TypeString,
						},
					},
				},
			},
		},
	}, api)
}
