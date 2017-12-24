package main

import (
	"context"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/tevjef/uct-core/hermes/token"
	"google.golang.org/grpc"
)

type tokenProvider struct {
	addr string
	port string
}

func (provider *tokenProvider) Token() (string, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", provider.addr, provider.port), grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Fatal("error connecting to token service")
	}

	defer conn.Close()

	client := token.NewTokenServiceClient(conn)

	tokenResponse, err := client.GetToken(context.Background(), &token.TokenRequest{})
	if err != nil {
		return "", nil
	}

	return tokenResponse.Token, err
}
