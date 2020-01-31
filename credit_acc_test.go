package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/testhelpers/docker"
	"github.com/isaacasensio/vault-eas/encryption"
	"github.com/isaacasensio/vault-eas/pkg/adding"
	"github.com/isaacasensio/vault-eas/pkg/getting"
	"github.com/isaacasensio/vault-eas/storage"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"path"
	"strings"
	"testing"
)

const (
	username = "user"
	password = "password"
	database = "database"
)

func TestEncryptionDecryptionAccountNumber(t *testing.T) {
	cs, postgres := setupPostgres(t)
	defer postgres.Terminate(context.Background())

	cleanup, vaultAddr, vaultToken, vaultKeyName := prepareTestContainer(t)
	defer cleanup()

	repo, err := storage.NewCustomerRepository(cs)
	require.NoError(t, err)

	client, err := encryption.NewVaultClient(vaultAddr, vaultToken, vaultKeyName)
	require.NoError(t, err)

	addService := adding.NewService(repo, client)
	getService := getting.NewService(repo, client)

	c := adding.Customer{
		Name:          "Peter Parker",
		AccountNumber: adding.AccountNumber("4169936079246876"),
	}

	id, err := addService.AddCustomer(c)
	require.NoError(t, err)

	encryptedAcc := fetchAccountFromDB(t, id, cs)
	assert.True(t, strings.HasPrefix(encryptedAcc, "vault:v1:"), "message not encrypted at rest")

	stored, err := getService.GetCustomer(id)
	require.NoError(t, err)

	assert.Equal(t, string(c.AccountNumber), string(stored.AccountNumber))
	assert.Equal(t, c.Name, stored.Name)
}

func fetchAccountFromDB(t *testing.T, id int64, cs string) string {
	t.Helper()

	db, err := sql.Open("postgres", cs)
	require.NoError(t, err)
	defer db.Close()

	var account string
	row := db.QueryRow(`SELECT CC_NUMBER FROM customer where id=$1`, id)
	err = row.Scan(&account)
	require.NoError(t, err)

	return account
}

func setupPostgres(t *testing.T) (string, testcontainers.Container) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:12.1",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     username,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       database,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ip, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}

	cs := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		ip, port.Port(), username, password, database)

	db, err := sql.Open("postgres", cs)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE customer(
			ID serial PRIMARY KEY, 
			NAME VARCHAR (50) NOT NULL, 
			CC_NUMBER VARCHAR(256) NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}

	return cs, container
}

func prepareTestContainer(t *testing.T) (cleanup func(), retAddress, token, keyName string) {
	t.Helper()

	testToken, err := uuid.GenerateUUID()
	require.NoError(t, err)
	testKeyName, err := uuid.GenerateUUID()
	require.NoError(t, err)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	dockerOptions := &dockertest.RunOptions{
		Repository: "vault",
		Tag:        "latest",
		Cmd: []string{"server", "-log-level=trace", "-dev", fmt.Sprintf("-dev-root-token-id=%s", testToken),
			"-dev-listen-address=0.0.0.0:8200"},
	}
	resource, err := pool.RunWithOptions(dockerOptions)
	require.NoError(t, err)

	cleanup = func() {
		docker.CleanupResource(t, pool, resource)
	}

	retAddress = fmt.Sprintf("http://127.0.0.1:%s", resource.GetPort("8200/tcp"))
	// exponential backoff-retry
	if err = pool.Retry(func() error {
		vaultConfig := api.DefaultConfig()
		vaultConfig.Address = retAddress
		if err := vaultConfig.ConfigureTLS(&api.TLSConfig{
			Insecure: true,
		}); err != nil {
			return err
		}
		vault, err := api.NewClient(vaultConfig)
		if err != nil {
			return err
		}
		vault.SetToken(testToken)

		// Set up transit
		if err := vault.Sys().Mount("transit", &api.MountInput{
			Type: "transit",
		}); err != nil {
			return err
		}

		// Create default aesgcm key
		if _, err := vault.Logical().Write(path.Join("transit", "keys", testKeyName), map[string]interface{}{}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		cleanup()
		require.NoError(t, err)
	}
	return cleanup, retAddress, testToken, testKeyName
}

