package soauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"gitlab.sendo.vn/core/golang-sdk"
)

type httpHandler struct {
	log sdms.Logger

	oAuthAppID string
	oAuthURI   string
}

func initHTTPHandler(log sdms.Logger, oAuthAppID string, oAuthURI string) *httpHandler {

	h := &httpHandler{
		log: log,

		oAuthAppID: oAuthAppID,
		oAuthURI:   oAuthURI,
	}

	return h

}

func (h *httpHandler) GetUserUsingToken(token string) (UserStruct, error) {

	requestUri := h.oAuthURI + "/v2/status"

	requestPayload, _ := json.Marshal(map[string]string{
		"app_id": h.oAuthAppID,
		"token":  token,
	})

	request, _ := http.NewRequest("POST", requestUri, bytes.NewBuffer(requestPayload))
	request.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}

	response, err := httpClient.Do(request)

	if err != nil {
		return UserStruct{}, errors.New(ERR_TRANSPORT_IS_CLOSING)
	}

	defer response.Body.Close()

	responseBody, _ := ioutil.ReadAll(response.Body)

	var userObject UserStruct

	err = json.Unmarshal(responseBody, &userObject)

	if err != nil || userObject.FPTID == 0 {
		return UserStruct{}, errors.New(ERR_INVALID_TOKEN)
	}

	return userObject, nil

}
