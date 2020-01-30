package soauth

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"

	"gitlab.sendo.vn/core/golang-sdk"
	"gitlab.sendo.vn/protobuf/internal-apis-go/oauth"
)

type gRPCHandler struct {
	log sdms.Logger

	oAuthAppID string
	oAuthURI   string

	streamIncrement uint64
	streamClient    oauth.OAuthService_GetUserUsingTokenV2Client

	mapWaiter map[uint64]chan oauth.GetUserUsingTokenResponseV2
	mu        *sync.Mutex

	reconnectCounter uint64
}

func initGRPCHandler(log sdms.Logger, oAuthAppID string, oAuthURI string) *gRPCHandler {

	h := &gRPCHandler{
		log: log,

		oAuthAppID: oAuthAppID,
		oAuthURI:   oAuthURI,

		streamIncrement: 0,
	}

	h.mapWaiter = make(map[uint64]chan oauth.GetUserUsingTokenResponseV2)
	h.mu = &sync.Mutex{}

	go h.init()

	for {

		if h.streamClient != nil {

			h.log.Info("Connected to OAuth Service successfully.")

			break

		}

	}

	return h

}

func (h *gRPCHandler) init() {

	for {

		h.mu.Lock()

		err := h.connect()

		h.mu.Unlock()

		if err != nil {

			h.cleanWaitClient()

			h.reconnectCounter++

			if h.reconnectCounter > 5 {

				log.Fatalf("Failed to connect to OAuth Service, panic...")

				panic(1)

			}

			h.streamClient = nil

			time.Sleep(time.Second)

			continue

		}

		h.reconnectCounter = 0

		for {

			in, err := h.streamClient.Recv()

			if err != nil {

				break

			}

			h.mu.Lock()

			if waitClient, found := h.mapWaiter[in.MessageId]; found {

				waitClient <- *in

				close(waitClient)

				delete(h.mapWaiter, in.MessageId)

			}

			h.mu.Unlock()

		}

	}

}

func (h *gRPCHandler) connect() error {

	conn, err := grpc.Dial(
		h.oAuthURI,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithTimeout(5*time.Second),
	)

	if err != nil {

		h.log.Errorf("Can't connect to OAuth Service via gRPC, retrying %d/%d... (%v)", h.reconnectCounter, 5, err.Error())

		return err

	}

	client := oauth.NewOAuthServiceClient(conn)

	h.streamClient, err = client.GetUserUsingTokenV2(context.Background())

	if err != nil {

		h.log.Errorf("Can't initialize gRPC OAuth Client, retrying %d/%d... (%v)", h.reconnectCounter, 5, err.Error())

		return err

	}

	return nil

}

func (h *gRPCHandler) cleanWaitClient() {

	for i, waitClient := range h.mapWaiter {

		close(waitClient)

		delete(h.mapWaiter, i)

	}

}

func (h *gRPCHandler) GetUserUsingToken(token string) (UserStruct, error) {

	if h.streamClient == nil {
		return UserStruct{}, errors.New(ERR_TRANSPORT_IS_CLOSING)
	}

	waitClient := make(chan oauth.GetUserUsingTokenResponseV2)

	// Sending Request

	h.mu.Lock()

	h.streamIncrement++

	increment := h.streamIncrement

	h.mapWaiter[increment] = waitClient

	h.streamClient.Send(&oauth.GetUserUsingTokenRequestV2{
		MessageId: increment,
		AppId:     h.oAuthAppID,
		Token:     token,
	})

	h.mu.Unlock()

	// Receive Request

	res := <-waitClient

	if res.ErrorId == 403 {
		return UserStruct{}, errors.New(ERR_INVALID_TOKEN)
	}

	if res.ErrorId == 500 {
		return UserStruct{}, errors.New(ERR_INTERNAL_SERVER_ERROR)
	}

	return UserStruct{
		CustomerID: res.CustomerId,
		FPTID:      res.FptId,

		Email: res.Email,
		Phone: res.Phone,

		CheckoutVerified: res.CheckoutVerified,
		EmailVerified:    res.EmailVerified,
		PhoneVerified:    res.PhoneVerified,

		FirstName: res.FirstName,
		LastName:  res.LastName,

		Avatar:   res.Avatar,
		Birthday: res.Birthday,
		Gender:   res.Gender,

		DefaultShipping: res.DefaultShipping,

		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil

}
