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

func scalingoAPIToken() string {
	token, ok := os.LookupEnv(SCALINGO_API_TOKEN_VARIABLE)
	if !ok {
		return ""
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

func (s *Scalingo) PrepareRequest(method, apiURL, path, token string, requestBody *bytes.Buffer) *http.Request {
	if apiURL == "" {
		apiURL = s.ApiURL
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := s.ApiURL + path
	return s.PrepareRequestForURL(method, url, token, requestBody)
}

func (s *Scalingo) PrepareRequestForURL(method, url, token string, requestBody *bytes.Buffer) *http.Request {
	var req *http.Request
	var err error
	if requestBody != nil {
		req, err = http.NewRequest(method, url, requestBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	maybePanic(err)
	if token == "" {
		log.Printf("Preparing %s request of %s now.\n", method, url)
	} else {
		req.SetBasicAuth("", token)
		log.Printf("Preparing %s request of %s with token now.\n", method, url)
	}
	return req
}

func (s *Scalingo) PerformRequest(method, apiURL, path string, requestBody *bytes.Buffer) ([]byte, error) {
	var body []byte
	if s.BearerToken == "" {
		s.BearerToken = s.fetchBearerToken()
	}
	body, err := s.PerformRequestWithToken(method, apiURL, path, s.BearerToken, requestBody)
	if err == tokenInvalidError {
		s.BearerToken = s.fetchBearerToken()
		body, err = s.PerformRequestWithToken(method, apiURL, path, s.BearerToken, requestBody)
	}
	return body, err
}

func (s *Scalingo) PerformRequestWithToken(method, apiURL, path, token string, requestBody *bytes.Buffer) ([]byte, error) {
	var err error
	req := s.PrepareRequest(method, apiURL, path, token, requestBody)
	var body []byte
	s.FetchResponse(req, func(res *http.Response) {
		if res.StatusCode == 401 {
			err = tokenInvalidError
			return
		}
		if res.StatusCode < 400 {
			body, err = ioutil.ReadAll(res.Body)
			maybePanic(err)
			log.Printf("Performed %s request to %s successfully.\n", req.Method, req.URL)
		} else {
			err = fmt.Errorf(
				"Requesting %s %s has failed with status=%d\n",
				req.Method, req.URL, res.StatusCode,
			)
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
