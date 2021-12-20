package client

import (
	"context"
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
}

func (realm *Realm) NewClient() *Client {
	return &Client{
		Realm: realm,
	}
}

type Client struct {
	Realm *Realm
}

func (client *Client) GetAuthorizationEndpoint(
	clientId string,
	redirectUri string,
	nonce string,
) string {
	var openIdConfiguration OpenIdConfiguration
	restClient := resty.New()

	if _, err := restClient.R().EnableTrace().SetResult(&openIdConfiguration).Get(client.Realm.OpenIdConfigurationEndpoint); err != nil {
		log.Fatalln(err)
	}

	u, _ := url.Parse(openIdConfiguration.AuthorizationEndpoint)
	query := u.Query()
	query.Set("response_type", "id_token")
	query.Set("client_id", clientId)
	query.Set("redirect_uri", redirectUri)
	query.Set("nonce", nonce)
	u.RawQuery = query.Encode()
	authorizationEndpoint := u.String()

	return authorizationEndpoint
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

	authorizationEndpoint := client.GetAuthorizationEndpoint(clientId, redirectUri, nonce)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(authorizationEndpoint),
	); err != nil {
		return "", err
	}

	location := <-locationChan
	uri, _ := url.Parse(location)
	q, _ := url.ParseQuery(uri.Fragment)
	idToken := q.Get("id_token")

	return idToken, nil
}
