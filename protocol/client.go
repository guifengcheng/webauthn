package protocol

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
)

// CollectedClientData represents the contextual bindings of both the WebAuthn Relying Party
// and the client. It is a key-value mapping whose keys are strings. Values can be any type
// that has a valid encoding in JSON. Its structure is defined by the following Web IDL.
// https://www.w3.org/TR/webauthn/#sec-client-data
type CollectedClientData struct {
	// Type the string "webauthn.create" when creating new credentials,
	// and "webauthn.get" when getting an assertion from an existing credential. The
	// purpose of this member is to prevent certain types of signature confusion attacks
	//(where an attacker substitutes one legitimate signature for another).
	Type         CeremonyType  `json:"type"`
	Challenge    string        `json:"challenge"`
	Origin       string        `json:"origin"`
	TokenBinding *TokenBinding `json:"tokenBinding,omitempty"`
}

type CeremonyType string

const (
	CreateCeremony CeremonyType = "webauthn.create"
	AssertCeremony CeremonyType = "webauthn.get"
)

type TokenBinding struct {
	Status TokenBindingStatus `json:"status"`
	ID     string             `json:"id,omitempty"`
}

type TokenBindingStatus string

const (
	// Present - Indicates token binding was used when communicating with the
	// Relying Party. In this case, the id member MUST be present.
	Present TokenBindingStatus = "present"
	// Supported -  Indicates token binding was used when communicating with the
	// negotiated when communicating with the Relying Party.
	Supported TokenBindingStatus = "supported"
)

// Verify handles steps 3 through 6 of verfying the registering client data of a
// new credential and steps 7 through 10 of verifying an authentication assertion
// See https://www.w3.org/TR/webauthn/#registering-a-new-credential
// and https://www.w3.org/TR/webauthn/#verifying-assertion
func (c *CollectedClientData) Verify(storedChallenge []byte, ceremony CeremonyType, relyingPartyOrigin string) error {

	// Registration Step 3. Verify that the value of C.type is webauthn.create.

	// Assertion Step 7. Verify that the value of C.type is the string webauthn.get.
	if c.Type != ceremony {
		err := ErrVerification.WithDetails("Error validating ceremony type")
		err.WithInfo(fmt.Sprintf("Expected Value: %s\n Received: %s\n", ceremony, c.Type))
		return err
	}

	// Registration Step 4. Verify that the value of C.challenge matches the challenge
	// that was sent to the authenticator in the create() call.

	// Assertion Step 8. Verify that the value of C.challenge matches the challenge
	// that was sent to the authenticator in the PublicKeyCredentialRequestOptions
	// passed to the get() call.

	clientChallengeBytes, err := base64.RawStdEncoding.DecodeString(c.Challenge)
	encodedStoredChallenge := make([]byte, len(clientChallengeBytes))
	base64.StdEncoding.Encode(encodedStoredChallenge, storedChallenge)
	if err != nil {
		return ErrParsingData.WithDetails("Error parsing the authenticator challenge")
	}

	if !bytes.Equal(encodedStoredChallenge, clientChallengeBytes) {
		err := ErrVerification.WithDetails("Error validating challenge")
		fmt.Printf("Expected b Value: %s\nReceived b: %s\n", encodedStoredChallenge, clientChallengeBytes)
		return err.WithInfo(fmt.Sprintf("Expected b Value: %#v\nReceived b: %#v\n", encodedStoredChallenge, clientChallengeBytes))
	}

	// Registration Step 5 & Assertion Step 9. Verify that the value of C.origin matches
	// the Relying Party's origin.
	clientDataOrigin, err := url.Parse(c.Origin)
	if err != nil {
		return ErrParsingData.WithDetails("Error decoding clientData origin as URL")
	}

	if clientDataOrigin.Hostname() != relyingPartyOrigin {
		fmt.Printf("Expected Value: %s\n Received: %s\n", relyingPartyOrigin, c.Origin)
		err := ErrVerification.WithDetails("Error validating origin")
		return err.WithInfo(fmt.Sprintf("Expected Value: %s\n Received: %s\n", relyingPartyOrigin, c.Origin))
	}

	// Registration Step 6 and Assertion Step 10. Verify that the value of C.tokenBinding.status
	// matches the state of Token Binding for the TLS connection over which the assertion was
	// obtained. If Token Binding was used on that TLS connection, also verify that C.tokenBinding.id
	// matches the base64url encoding of the Token Binding ID for the connection.

	// Not yet fully implemented by the spec, browsers, and me.

	return nil
}
