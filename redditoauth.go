package redditoauth

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type credentials struct {
	ClientID     string
	ClientSecret string
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Duration     string
	UserAgent    string
}

var creds credentials

//retrieves the saved clientID.
//If no clientID is set, it returns the empty string
func GetClientID() string {
	return creds.ClientID
}

//sets the clientID to the given string
func SetClientID(cid string) {
	creds.ClientID = cid
}

//retrieves the saved client secret.
//If no client secret is set, it returns the empty string
func GetClientSecret() string {
	return creds.ClientSecret
}

//sets the clientID to the given string
func SetClientSecret(sec string) {
	creds.ClientSecret = sec
}

//retrieves the saved access token.
//If no access token is set, it returns the empty string
func GetAccessToken() string {
	return creds.AccessToken
}

//sets the access token to the given string.
//It should automatically be set in the PerformHandshake function
func SetAccessToken(acc string) {
	creds.AccessToken = acc
}

//retrieves the saved refresh token.
//If no refresh token is set, it returns the empty string
func GetRefreshToken() string {
	return creds.RefreshToken
}

//sets the refresh token to the given string.
//It should automatically be set in the PerformHandshake function
func SetRefreshToken(refresh string) {
	creds.RefreshToken = refresh
}

//retrieves the saved user agent string
//if it isn't set, it is set to the empty string
func GetUserAgent() string {
	return creds.UserAgent
}

//sets the user agent string the the target string
func SetUserAgent(agent string) {
	creds.UserAgent = agent
}

func buildurl(scopes []string, perm bool) (string, error) {
	params := url.Values{}
	params.Add("client_id", creds.ClientID)
	params["scope"] = scopes
	params.Add("response_type", "code")
	if perm {
		params.Add("duration", "permanent")
	} else {
		params.Add("duration", "temporary")
	}
	params.Add("state", "redditoauth")
	params.Add("redirect_uri", "http://localhost")

	return "https://https://www.reddit.com/api/v1/authorize?" + params.Encode(), nil
}

//performs an oauth handshake. Returns 3 values.
//The first is an access token. The second a refreshToken. If an error occurs,
//the third value will have an error and the other return fields will be empty strings
//The first argument is the callback uri,
//the second is a slice of strings representing reddit scopes.
//The last argument is a bool that determines if we should ask for a refresh token.
func PerformHandshake(callback string, scopes []string, perm bool) (string, string, error) {
	err := validateCreds()
	if err != nil {
		return "", "", err
	}

	link, err := buildurl(scopes, perm)
	if err != nil {
		return "", "", err
	}
	fmt.Println("Please visit the following url:", link)
	fmt.Print("Please enter in the code: ")
	var code string
	_, err = fmt.Scanln(&code)
	if err != nil {
		return "", "", err
	}
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("code", code)
	params.Add("redirect_uri", callback)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Add("User-Agent", creds.UserAgent)
	req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", errors.New("error in request: " + resp.Status)
	}
	dataBuf := new(bytes.Buffer)
	dataBuf.ReadFrom(resp.Body)
	err = json.Unmarshal(dataBuf.Bytes(), &creds)
	if err != nil {
		return "", "", err
	}
	return creds.AccessToken, creds.RefreshToken, nil
}

//exchanges refresh token for access token
func refreshCreds() error {
	err := validateCreds()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", creds.RefreshToken)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
	req.Header.Add("User-Agent", creds.UserAgent)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error in request: " + resp.Status)
	}

	dataBuf := new(bytes.Buffer)
	dataBuf.ReadFrom(resp.Body)
	err = json.Unmarshal(dataBuf.Bytes(), &creds)
	if err != nil {
		return err
	}

	return nil
}

func validateCreds() error {
	if creds.ClientID == "" {
		return errors.New("Client ID not set")
	}
	if creds.ClientSecret == "" {
		return errors.New("Client Secret not set")
	}
	if creds.UserAgent == "" {
		return errors.New("User agent not set")
	}
	return nil
}

//makes an api request to reddit.
//The first argument is the http method (GET, POST, PATCH, etc...)
//The second argument is the url string with all parameters (e.g. https://oauth.reddit.com/api/v1/me).
//The third argument is an optional body to the request.
//The fourth argument is a pointer to a map from string to empty interface to store the response json if successful.
//It returns an error if any.
//Make sure to have all fields, including access token and/or refresh token
func MakeApiReq(method, urlstr string, body io.Reader, result *map[string]interface{}) error {
	if creds.AccessToken == "" {
		return errors.New("Access token missing")
	}

	req, err := http.NewRequest(method, urlstr, body)
	req.Header.Add("User-Agent", creds.UserAgent)
	req.Header.Add("Authorization", "bearer "+creds.AccessToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error in request: " + urlstr + ": " + resp.Status)
	}

	dataBuf := new(bytes.Buffer)
	dataBuf.ReadFrom(resp.Body)
	err = json.Unmarshal(dataBuf.Bytes(), result)
	if err != nil {
		return err
	}
	return nil
}
