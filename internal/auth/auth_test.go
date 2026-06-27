package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGoodToken(t *testing.T) {
	// 1. Arrange: set up inputs
	// 2. Act: call the function
	// 3. Assert: check the result with t.Errorf or t.Fatalf

	id := uuid.New()
	tS := "Blahhhh"

	token, err := MakeJWT(id, tS, (time.Minute * 15))
	if err != nil {
		t.Errorf("Failed to generate token %v", err)
		return
	}

	returnedID, err := ValidateJWT(token, tS)
	if err != nil {
		t.Errorf("Failed to validate token %v", err)
		return
	}

	if returnedID != id {
		t.Errorf("Returned token ID not equal to original token ID %v", err)
		return
	}

	fmt.Printf("Test passed")
}

func TestBearerToken(t *testing.T) {
	expected := "abcd1234"
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+expected)
	bearer, err := GetBearerToken(headers)
	if err != nil {
		t.Errorf("Failed to retrieve bearer token %v", err)
		return
	}
	if bearer != expected {
		t.Errorf("Bearer is not expected value")
		return
	}

	fmt.Printf("Test passed")

}
