package soauth

import (
	"errors"
	"flag"
	"fmt"

	"gitlab.sendo.vn/core/golang-sdk"
)

const (
	ERR_INTERNAL_SERVER_ERROR = "internal server error"
	ERR_INVALID_TOKEN         = "invalid token"
	ERR_TRANSPORT_IS_CLOSING  = "transport is closing"
)

type OAuthConfig struct {
	App sdms.Application
}

type appService interface {
	InitFlags()
	Configure() error
	Cleanup()
}

type OAuthService interface {
	appService

	GetUserUsingToken(string) (UserStruct, error)
}

type appOAuthServiceImpl struct {
	log sdms.Logger

	oAuthProtocol string
	oAuthAppID    string
	oAuthURI      string

	httpHandler *httpHandler
	gRPCHandler *gRPCHandler
}

func NewOAuthService(config *OAuthConfig) OAuthService {
	return &appOAuthServiceImpl{
		log: config.App.(sdms.SdkApplication).GetLog("soauth"),
	}
}

func (s *appOAuthServiceImpl) InitFlags() {
	flag.StringVar(&s.oAuthProtocol, "oauth-protocol", "grpc", "Choose one of 'grpc' or 'http'; default is 'grpc'.")
	flag.StringVar(&s.oAuthAppID, "oauth-app-id", "", "Application ID for OAuth Service, provided by Account team.")
	flag.StringVar(&s.oAuthURI, "oauth-uri", "", "Connecting URI for OAuth Service, provided by SysAdmin team.")
}

func (s *appOAuthServiceImpl) Configure() error {

	if s.oAuthProtocol != "grpc" && s.oAuthProtocol != "http" {
		fmt.Println("FATAL appConfService: 'oauth-protocol' has not been configured")
		return errors.New("")
	}

	if s.oAuthAppID == "" {
		fmt.Println("FATAL appConfService: 'oauth-app-id' has not been configured")
		return errors.New("")
	}

	if s.oAuthAppID == "" {
		fmt.Println("FATAL appConfService: 'oauth-uri' has not been configured")
		return errors.New("")
	}

	s.log.Infof("Prepare to connect to OAuth Service via %s", s.oAuthProtocol)

	if s.oAuthProtocol == "grpc" {

		s.gRPCHandler = initGRPCHandler(s.log, s.oAuthAppID, s.oAuthURI)

	} else if s.oAuthProtocol == "http" {

		s.httpHandler = initHTTPHandler(s.log, s.oAuthAppID, s.oAuthURI)

	}

	return nil

}

func (s *appOAuthServiceImpl) Cleanup() {
	// Nothing goes here...
}

func (s *appOAuthServiceImpl) GetUserUsingToken(token string) (UserStruct, error) {

	if s.oAuthProtocol == "grpc" {

		return s.gRPCHandler.GetUserUsingToken(token)

	} else if s.oAuthProtocol == "http" {

		return s.httpHandler.GetUserUsingToken(token)

	}

	return UserStruct{}, nil

}
