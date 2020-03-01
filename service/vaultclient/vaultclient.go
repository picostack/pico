package vaultclient

import (
	"github.com/hashicorp/vault/api"
)

type Client struct {
	vault *api.Client
}
