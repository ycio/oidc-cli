package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"ycio/oidc-cli/client"

	"github.com/jessevdk/go-flags"
)

func main() {
	osArgs := os.Args

	if len(osArgs) > 1 && osArgs[1] == "implicit-flow" {
		var opts struct {
			Endpoint    string `short:"e" long:"endpoint" description:"Endpoint of well-know openid configuration" required:"true"`
			ClientId    string `short:"c" long:"client-id" description:"Client id" required:"true"`
			RedirectURI string `short:"r" long:"redirect-uri" description:"Redirect URI" required:"true"`
			Nonce       string `short:"n" long:"nonce" description:"Nonce (optional)" required:"false"`
		}

		if _, err := flags.ParseArgs(&opts, osArgs[2:]); err != nil {
		} else {
			realm := &client.Realm{
				OpenIdConfigurationEndpoint: opts.Endpoint,
			}

			client := realm.NewClient()

			var nonce string

			if strings.TrimSpace(opts.Nonce) == "" {
				nonce = fmt.Sprint((time.Now().UnixNano()))
			} else {
				nonce = opts.Nonce
			}

			if token, err := client.GetIdToken(
				opts.ClientId,
				opts.RedirectURI,
				nonce,
			); err != nil {
				log.Fatalln(err)
			} else {
				fmt.Println(token)
			}
		}
	} else {
		log.Fatalln(errors.New("incorrect or lack of subcommand"))
	}
}
