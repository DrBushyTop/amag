package auth

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"time"
)

type Client struct {
	cred azcore.TokenCredential
}

func NewAuthClient(opts ...AuthClientOpts) (*Client, error) {
	client := Client{}

	for _, opt := range opts {
		err := opt(&client)
		if err != nil {
			return nil, fmt.Errorf("NewAuthClient: failed to apply option: %w", err)
		}
	}

	if client.cred == nil {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("NewAuthClient: failed to create default azure credential: %w", err)
		}
		client.cred = cred
	}

	return &client, nil
}

func (a *Client) GetAccessToken(scopes []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := a.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: scopes,
	})

	if err != nil {
		return "", fmt.Errorf("GetAccessToken: failed to get token: %w", err)
	}

	return token.Token, nil
}

type AuthClientOpts func(client *Client) error

func WithCredential(cred azcore.TokenCredential) AuthClientOpts {
	return func(client *Client) error {
		client.cred = cred
		return nil
	}
}

func WithCredentialOptions(options azidentity.DefaultAzureCredentialOptions) AuthClientOpts {
	return func(client *Client) error {
		cred, err := azidentity.NewDefaultAzureCredential(&options)
		if err != nil {
			return err
		}
		client.cred = cred
		return nil
	}
}