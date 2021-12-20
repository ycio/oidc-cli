package main

import (
	"fmt"
	"log"
	cli "ycio/oidc-cli/client"
)

func main() {
	realm := &cli.Realm{
		OpenIdConfigurationEndpoint: "http://localhost:8090/auth/realms/realm-name/.well-known/openid-configuration",
	}

	client := realm.NewClient()

	if token, err := client.GetIdToken(
		"clientId",
		"http://localhost:8080/auth",
		"nonce123",
	); err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(token)
	}
}
