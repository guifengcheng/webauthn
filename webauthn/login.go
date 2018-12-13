package webauthn

import (
	"bytes"
	"fmt"
	"net/http"

	p "github.com/duo-labs/webauthn/protocol"
)

// BEGIN REGISTRATION
// These objects help us creat the CredentialCreationOptions
// that will be passed to the authenticator via the user client

type LoginOption func(*p.PublicKeyCredentialRequestOptions)

func (webauthn *WebAuthn) BeginLogin(user User, opts ...LoginOption) (*p.PublicKeyCredentialRequestOptions, SessionData, error) {
	challenge, err := p.CreateChallenge()
	if err != nil {
		return nil, SessionData{}, err
	}

	requestOptions := p.PublicKeyCredentialRequestOptions{
		Challenge: challenge,
		Timeout:   webauthn.Config.Timeout,
	}

	for _, setter := range opts {
		setter(&requestOptions)
	}

	sessionData := SessionData{
		Challenge: challenge,
		UserID:    user.WebAuthnID(),
	}

	return &requestOptions, sessionData, nil
}

func (webauthn *WebAuthn) FinishLogin(user User, session SessionData, response *http.Request) (*Credential, error) {
	if !bytes.Equal(user.WebAuthnID(), session.UserID) {
		p.ErrBadRequest.WithDetails("ID mismatch for User and Session")
	}

	parsedResponse, err := p.ParseCredentialCreationResponse(response)
	if err != nil {
		fmt.Println(err)
		return nil, p.ErrBadRequest.WithDetails("fuddck")
	}
	fmt.Printf("got the following:\n %+v\n\n", parsedResponse)
	return nil, nil
}
