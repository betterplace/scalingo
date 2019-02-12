package scalingo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antonholmquist/jason" // move to type struct parsing
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	SCALINGO_API_TOKEN_VARIABLE = "SCALINGO_API_TOKEN"
	SCALINGO_API_URL            = "https://api.scalingo.com"
	SCALINGO_AUTH_API_URL       = "https://auth.scalingo.com/v1/tokens/exchange"
)

type Scalingo struct {
	ApiURL      string
	ApiToken    string
	BearerToken string
}

var tokenInvalidError = errors.New("token invalid")

type HTTPError struct {
	msg        string
	StatusCode int
}

func (e *HTTPError) Error() string { return e.msg }

func scalingoAPIToken() string {
	token, ok := os.LookupEnv(SCALINGO_API_TOKEN_VARIABLE)
	if !ok {
		log.Panicf(
			"Couldn't get Scalingo API Token via env var %s",
			SCALINGO_API_TOKEN_VARIABLE,
		)
	}
	return token
}

func NewScalingo(apiToken string) *Scalingo {
	if apiToken == "" {
		apiToken = scalingoAPIToken()
	}
	return &Scalingo{ApiToken: apiToken, ApiURL: SCALINGO_API_URL}
}

func maybePanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func (s *Scalingo) PrepareRequest(method, apiURL, path, token string, jsonBody *bytes.Buffer) *http.Request {
	if apiURL == "" {
		apiURL = s.ApiURL
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := apiURL + path
	return s.PrepareRequestForURL(method, url, token, jsonBody)
}

func (s *Scalingo) PrepareRequestForURL(method, url, token string, jsonBody *bytes.Buffer) *http.Request {
	var req *http.Request
	var err error
	if jsonBody != nil {
		req, err = http.NewRequest(method, url, jsonBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	maybePanic(err)
	if token == "" {
		log.Printf("Preparing %s request of %s now.\n", method, url)
	} else {
		if jsonBody != nil {
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
		}
		req.SetBasicAuth("", token)
		log.Printf("Preparing %s request of %s with token now.\n", method, url)
	}
	return req
}

func (s *Scalingo) PerformRequest(method, apiURL, path string, jsonBody *bytes.Buffer) ([]byte, error) {
	var body []byte
	if s.BearerToken == "" {
		s.BearerToken = s.fetchBearerToken()
	}
	body, err := s.PerformRequestWithToken(method, apiURL, path, s.BearerToken, jsonBody)
	if err == tokenInvalidError {
		s.BearerToken = s.fetchBearerToken()
		body, err = s.PerformRequestWithToken(method, apiURL, path, s.BearerToken, jsonBody)
	}
	return body, err
}

func (s *Scalingo) PerformRequestWithToken(method, apiURL, path, token string, jsonBody *bytes.Buffer) ([]byte, error) {
	var err error
	req := s.PrepareRequest(method, apiURL, path, token, jsonBody)
	var body []byte
	s.FetchResponse(req, func(res *http.Response) {
		if res.StatusCode == 401 {
			err = tokenInvalidError
			return
		}
		body, err = ioutil.ReadAll(res.Body)
		maybePanic(err)
		if res.StatusCode < 400 {
			log.Printf("Performed %s request to %s successfully.\n", req.Method, req.URL)
		} else {
			msg := fmt.Sprintf(
				"Requesting %s %s has failed with status=%d:\n%s",
				req.Method, req.URL, res.StatusCode, string(body),
			)
			err = &HTTPError{StatusCode: res.StatusCode, msg: msg}
		}
	})
	return body, err
}

type BearerTokenResponse struct {
	Token string `json:"token"`
}

func (s *Scalingo) fetchBearerToken() string {
	req := s.PrepareRequestForURL("POST", SCALINGO_AUTH_API_URL, s.ApiToken, nil)
	var token string
	s.FetchResponse(req, func(res *http.Response) {
		if res.StatusCode >= 400 {
			log.Printf("Requesting %s %s has failed with status=%d:\n",
				req.Method, req.URL, res.StatusCode)
		} else {
			body, err := ioutil.ReadAll(res.Body)
			maybePanic(err)
			log.Printf("Performed %s request to %s successfully.\n", req.Method, req.URL)
			var btr BearerTokenResponse
			err = json.Unmarshal(body, &btr)
			maybePanic(err)
			token = btr.Token
		}
	})
	return token
}

func (s *Scalingo) FetchResponse(req *http.Request, block func(res *http.Response)) {
	client := &http.Client{}
	res, err := client.Do(req)
	defer res.Body.Close()
	maybePanic(err)
	block(res)
}

func (s *Scalingo) GetAppNames() []string {
	path := "/v1/apps"
	jsonText, err := s.PerformRequest("GET", "", path, nil)
	maybePanic(err)
	data, err := jason.NewObjectFromBytes(jsonText)
	maybePanic(err)
	appsData, err := data.GetObjectArray("apps")
	maybePanic(err)
	var appNames []string
	for _, appData := range appsData {
		name, err := appData.GetString("name")
		maybePanic(err)
		appNames = append(appNames, name)
	}
	return appNames
}

func (s *Scalingo) GetAppVariables(app string) map[string]string {
	envVars := make(map[string]string)
	path := "/v1/apps/" + app + "/variables"
	jsonText, err := s.PerformRequest("GET", "", path, nil)
	maybePanic(err)
	data, err := jason.NewObjectFromBytes(jsonText)
	maybePanic(err)
	variables, err := data.GetObjectArray("variables")
	for _, varData := range variables {
		name, err := varData.GetString("name")
		maybePanic(err)
		value, err := varData.GetString("value")
		maybePanic(err)
		envVars[name] = value
	}
	return envVars
}
