package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONRawScanString(t *testing.T) {
	var j JSONRaw
	err := j.Scan(`{"onboarding_completed":true}`)
	require.NoError(t, err)
	require.Equal(t, `{"onboarding_completed":true}`, string(j))
}

func TestJSONRawScanBytes(t *testing.T) {
	var j JSONRaw
	err := j.Scan([]byte(`{"goal":"save"}`))
	require.NoError(t, err)
	require.Equal(t, `{"goal":"save"}`, string(j))
}

func TestJSONRawScanNil(t *testing.T) {
	var j JSONRaw
	err := j.Scan(nil)
	require.NoError(t, err)
	require.Equal(t, `{}`, string(j))
}

func TestJSONRawValue(t *testing.T) {
	j := JSONRaw(`{"a":1}`)
	v, err := j.Value()
	require.NoError(t, err)
	require.Equal(t, `{"a":1}`, v)

	empty := JSONRaw(nil)
	v, err = empty.Value()
	require.NoError(t, err)
	require.Equal(t, `{}`, v)
}
