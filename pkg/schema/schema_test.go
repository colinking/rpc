package schema

func TestDiscover(t *testing.T) {
	require := require.New(t)

	api, err := Discover("./fixtures/simple")
	require.NoError(err)
	require.Equal(API{}, api)
}