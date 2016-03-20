package redditoauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
  "strings"
  "bytes"
)

type secrets struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"secret"`
}

type credentials struct {
	AccessToken  string `json:"access"`
	RefreshToken string `json:"refresh"`
	Duration     string `json:duration`
}

var creds credentials

func readSecrets(filename string) (secrets, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return secrets{}, err
	}

	var secs secrets
	err = json.Unmarshal(data, &secs)
	if err != nil {
		return secrets{}, err
	}

	return secs, nil
}

func buildurl(secs secrets, scopes []string, perm bool) (string, error) {
	params := url.Values{}
	params.Add("client_id", secs.ClientID)
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

func PerformHandshake(filename string, scopes []string, perm bool) error {
  redir := "http://localhost"
  secs, err := readSecrets(filename)
  if err != nil {
    return err
  }

	link, err := buildurl(secs, scopes, perm)
	if err != nil {
		return err
	}
	fmt.Println("Please visit the following url:", link)
	fmt.Print("Please enter in the code: ")
	var code string
	_, err = fmt.Scanln(&code)
	if err != nil {
		return err
	}
  params := url.Values{}
  params.Add("grant_type", "authorization_code")
  params.Add("code", code)
  params.Add("redirect_uri", redir)

  req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(params.Encode()))
  if err != nil {
    return err
  }

  req.Header.Add("User-Agent", "native: markov-monitor:beta-testing")
  req.SetBasicAuth(secs.ClientID, secs.ClientSecret)
  client := http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return err
  }

  fmt.Println(resp.Status)
  dataBuf := new(bytes.Buffer)
  dataBuf.ReadFrom(resp.Body)
  err = json.Unmarshal(dataBuf.Bytes(), &creds)
  if err != nil {
    return err
  }
  writeCreds()
  return nil
}

func writeCreds() error {
  data, err := json.Marshal(creds)
  if err != nil {
    return err
  }
  ioutil.WriteFile("creds.json", data, 0600)
  return nil
}

func refreshCreds() error {
  
  params := url.Values{}
  params.Add("grant_type", "refresh_token")
  params.Add("refresh_token", creds.RefreshToken)

  req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(params.Encode()))
  if err != nil {
    return err
  }
  req.SetBasicAuth(secs.ClientID, secs.ClientSecret)
  req.Header.Add("User-Agent", "native: markov-monitor:beta-testing")

}
