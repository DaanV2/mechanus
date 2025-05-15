package xcrypto_test

import (
	"testing"

	xcrypto "github.com/DaanV2/mechanus/server/pkg/extensions/crypto"
	"github.com/stretchr/testify/require"
)

func Test_Generate_RSA(t *testing.T) {
	key, err := xcrypto.GenerateRSAKeys()
	require.NoError(t, err)

	require.Greater(t, len(key.ID()), 0)
}
