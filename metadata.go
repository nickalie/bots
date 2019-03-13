package bots

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/nickalie/bots/utils"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/square/go-jose.v2"
)

type OpenIdMetadata struct {
	url         string
	lastUpdated int64
	keys        []jose.JSONWebKey
}

func NewOpenIdMetadata(url string) *OpenIdMetadata {
	return &OpenIdMetadata{url: url}
}

func (o *OpenIdMetadata) GetKey(kid string) (*rsa.PublicKey, error) {
	if o.lastUpdated < (time.Now().Unix() - 60*60*24*5) {
		err := o.refreshCache()

		if err != nil {
			fmt.Printf("refreshCache: %v\n", err)
			return nil, err
		}
	}

	for _, key := range o.keys {
		if key.KeyID != kid {
			continue
		}

		publicKey, ok := key.Key.(*rsa.PublicKey)

		if !ok {
			continue
		}

		return publicKey, nil
	}

	return nil, errors.New("key not found: " + kid)
}

func (o *OpenIdMetadata) refreshCache() error {
	var openIdConfig IOpenIdConfig

	resp, _, errs := gorequest.New().Get(o.url).EndStruct(&openIdConfig)

	if len(errs) == 0 && resp.StatusCode >= 400 {
		errs = append(errs, errors.New(fmt.Sprintf("Failed to load openID config: %d", resp.StatusCode)))
	}

	if len(errs) > 0 {
		return utils.ErrorFromArray(errs)
	}

	var jwkResponse jose.JSONWebKeySet
	resp, _, errs = gorequest.New().Get(openIdConfig.JwksUri).EndStruct(&jwkResponse)

	if len(errs) == 0 && resp.StatusCode >= 400 {
		errs = append(errs, errors.New(fmt.Sprintf("Failed to load keys: %d", resp.StatusCode)))
	}

	if len(errs) > 0 {
		return utils.ErrorFromArray(errs)
	}

	o.lastUpdated = time.Now().Unix()
	o.keys = jwkResponse.Keys
	return nil
}

type IOpenIdConfig struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	JwksUri               string   `json:"jwks_uri"`
	SigningAlgValues      []string `json:"id_token_signing_alg_values_supported"`
	AuthMethods           []string `json:"token_endpoint_auth_methods_supported"`
}
