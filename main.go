package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/coreos/go-oidc/v3/oidc"
)

var (
	clientID     = "81078899437-35v9t1anof742cji9ubadga6ps7s39om.apps.googleusercontent.com"
	clientSecret = "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"
	redirectURL  = "http://localhost:8080/callback"

	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier

	oauth2Config *oauth2.Config
)

func main() {
	ctx := context.Background()

	var err error
	provider, err = oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}

	// Configure OAuth2 with Google
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)

	fmt.Println("🚀 Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<a href="/login">🔐 Login with Google</a>`)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	state := "random-state" // Use real random in prod
	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	ctx := context.Background()

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "✅ Logged in as %s (%s)", claims.Name, claims.Email)
}
