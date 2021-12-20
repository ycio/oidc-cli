package client

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-resty/resty/v2"
)

type OpenIdConfiguration struct {
	TokenEndpoint         string `json:"token_endpoint"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
}

type Realm struct {
	OpenIdConfigurationEndpoint string
	openIdConfiguration         *OpenIdConfiguration
}

func (realm *Realm) NewRealm(openIdConfigurationEndpoint string) *Realm {
	return &Realm{
		OpenIdConfigurationEndpoint: openIdConfigurationEndpoint,
	}
}

func (realm *Realm) LoadConfiguration() error {
	var openIdConfiguration OpenIdConfiguration
	restClient := resty.New()

	if _, err := restClient.R().
		EnableTrace().
		SetResult(&openIdConfiguration).
		Get(realm.OpenIdConfigurationEndpoint); err != nil {
		return err
	}

	realm.openIdConfiguration = &openIdConfiguration

	return nil
}

func (realm *Realm) NewClient() *Client {
	return &Client{
		Realm: realm,
	}
}

type Client struct {
	Realm *Realm
}

func (client *Client) GetAuthorizationUrl(
	clientId string,
	redirectUri string,
	nonce string,
) (string, error) {
	client.Realm.LoadConfiguration()

	if client.Realm.openIdConfiguration == nil {
		return "", errors.New("authorization endpoint is not initialized.")
	}

	authorizationEndpoint := client.Realm.openIdConfiguration.AuthorizationEndpoint
	u, err := url.Parse(authorizationEndpoint)

	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("response_type", "id_token")
	query.Set("client_id", clientId)
	query.Set("redirect_uri", redirectUri)
	query.Set("nonce", nonce)
	u.RawQuery = query.Encode()
	authorizationUrl := u.String()

	return authorizationUrl, nil
}

func (client *Client) GetIdToken(
	clientId string,
	redirectUri string,
	nonce string,
) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	locationChan := make(chan string)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch v := ev.(type) {
		case *network.EventResponseReceivedExtraInfo:
			locationValue := v.Headers["Location"]

			switch location := locationValue.(type) {
			case string:
				if strings.HasPrefix(location, redirectUri) {
					locationChan <- location
				}
			}
		}
	})

	authorizationUrl, err := client.GetAuthorizationUrl(clientId, redirectUri, nonce)

	if err != nil {
		return "", err
	}

	if err := chromedp.Run(ctx,
		chromedp.Navigate(authorizationUrl),
	); err != nil {
		return "", err
	}

	location := <-locationChan
	uri, _ := url.Parse(location)
	q, _ := url.ParseQuery(uri.Fragment)
	idToken := q.Get("id_token")

	return idToken, nil
}
