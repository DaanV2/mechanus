package xrand_test

import (
	"testing"

	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"
	"github.com/stretchr/testify/require"
)

func Test_MustID(t *testing.T) {
	for l := range 64 {
		id := xrand.MustID(l)
		require.Len(t, id, l)
	}
}