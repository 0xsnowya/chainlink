package keystore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink/core/internal/cltest"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/configtest"
	"github.com/smartcontractkit/chainlink/core/internal/testutils/pgtest"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/dkgencryptkey"
)

func Test_DKGEncryptKeyStore_E2E(t *testing.T) {
	db := pgtest.NewSqlxDB(t)
	cfg := configtest.NewTestGeneralConfig(t)
	keyStore := keystore.ExposedNewMaster(t, db, cfg)
	require.NoError(t, keyStore.Unlock(cltest.Password))
	ks := keyStore.DKGEncrypt()

	assert.NotNil(t, ks)

	reset := func() {
		_, err := db.Exec("DELETE FROM encrypted_key_rings")
		require.NoError(t, err)
		keyStore.ResetXXXTestOnly()
		require.NoError(t, keyStore.Unlock(cltest.Password))
	}

	t.Run("initializes with an empty state", func(t *testing.T) {
		defer reset()
		keys, err := ks.GetAll()
		require.NoError(t, err)
		require.Equal(t, 0, len(keys))
	})

	t.Run("errors when getting non-existent ID", func(t *testing.T) {
		defer reset()
		_, err := ks.Get("non-existent-id")
		require.Error(t, err)
	})

	t.Run("creates a key", func(t *testing.T) {
		defer reset()
		key, err := ks.Create()
		require.NoError(t, err)
		retrievedKey, err := ks.Get(key.ID())
		require.NoError(t, err)
		require.Equal(t, key, retrievedKey)
	})

	t.Run("imports and exports a key", func(t *testing.T) {
		defer reset()
		key, err := ks.Create()
		require.NoError(t, err)
		exportJSON, err := ks.Export(key.ID(), cltest.Password)
		require.NoError(t, err)
		_, err = ks.Delete(key.ID())
		require.NoError(t, err)
		_, err = ks.Get(key.ID())
		require.Error(t, err)
		importedKey, err := ks.Import(exportJSON, cltest.Password)
		require.NoError(t, err)
		require.Equal(t, key.ID(), importedKey.ID())
		retrievedKey, err := ks.Get(key.ID())
		require.NoError(t, err)
		require.Equal(t, importedKey, retrievedKey)
	})

	t.Run("adds an externally created key / deletes a key", func(t *testing.T) {
		defer reset()
		newKey, err := dkgencryptkey.New()
		require.NoError(t, err)
		err = ks.Add(newKey)
		require.NoError(t, err)
		keys, err := ks.GetAll()
		require.NoError(t, err)
		require.Equal(t, 1, len(keys))
		_, err = ks.Delete(newKey.ID())
		require.NoError(t, err)
		keys, err = ks.GetAll()
		require.NoError(t, err)
		require.Equal(t, 0, len(keys))
		_, err = ks.Get(newKey.ID())
		require.Error(t, err)
	})

	t.Run("adds an externally created key/ensures it already exists", func(t *testing.T) {
		defer reset()

		newKey, err := dkgencryptkey.New()
		assert.NoError(t, err)
		err = ks.Add(newKey)
		assert.NoError(t, err)

		err = keyStore.DKGEncrypt().EnsureKey()
		assert.NoError(t, err)
		keys, err2 := ks.GetAll()
		assert.NoError(t, err2)

		require.Equal(t, 1, len(keys))
		require.Equal(t, newKey.ID(), keys[0].ID())
		require.Equal(t, newKey.PublicKey, keys[0].PublicKey)
	})

	t.Run("auto creates a key if it doesn't exists when trying to ensure it already exists", func(t *testing.T) {
		defer reset()

		keys, err := ks.GetAll()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(keys))

		err = keyStore.DKGEncrypt().EnsureKey()
		assert.NoError(t, err)

		keys, err = ks.GetAll()
		assert.NoError(t, err)

		require.NoError(t, err)
		require.Equal(t, 1, len(keys))
	})
}
