package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-http-utils/headers"
	"github.com/golang-jwt/jwt/v4"
	"github.com/controlplane-com/libs-go/pkg/common"
	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"io"
	"net/http"
	"strings"
)

func GetBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	p := strings.Split(authHeader, " ")
	if len(p) < 2 {
		return "", errors.New("no bearer token found")
	}
	if p[0] != "Bearer" {
		return "", errors.New("no bearer token found")
	}
	return p[len(p)-1], nil
}

func DoJSONRequestWithBodyAndResult[TBody any, TResult any](client *http.Client, method string, url string, body TBody, token string) (TResult, error) {
	var request *http.Request
	var z TResult

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return z, err
	}
	request, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return z, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	response, err := DoRequest(client, request)
	if err != nil {
		return z, err
	}
	parsed, err := ParseJsonResponseBody[TResult](response)
	if err != nil {
		return z, err
	}

	return parsed, err
}

func DoJSONRequestWithBody[TBody any](client *http.Client, method string, url string, body TBody, token string) error {
	var request *http.Request

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	_, err = DoRequest(client, request)
	return err
}

func DoJSONRequestWithResult[TResult any](client *http.Client, method string, url string, token string) (TResult, error) {
	var request *http.Request
	var z TResult

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return z, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	response, err := DoRequest(client, request)
	if err != nil {
		return z, err
	}

	parsed, err := ParseJsonResponseBody[TResult](response)
	if err != nil {
		return z, err
	}

	return parsed, err
}

func DoRequest(httpClient *http.Client, request *http.Request) (*http.Response, error) {
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode > 299 {
		body, _ := io.ReadAll(response.Body)
		return response, errors.New(fmt.Sprintf("invalid HTTP response code (%d). Details: %s", response.StatusCode, string(body)))
	}
	return response, nil
}

func ParseJsonRequestBody[T any](request *http.Request) (T, error) {
	var z T
	if request == nil {
		return z, errors.New("nil request given")
	}

	var body T
	b, err := io.ReadAll(request.Body)
	if err != nil {
		return z, err
	}
	err = json.Unmarshal(b, &body)
	if err != nil {
		return z, err
	}
	return body, nil
}

func ParseJsonResponseBody[T any](response *http.Response) (T, error) {
	var z T
	if response == nil {
		return z, errors.New("nil response given")
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return z, err
	}

	var parsed T
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return z, err
	}
	return parsed, nil
}

func ReturnError(w http.ResponseWriter, err error) (int, error) {
	errResponse := cplnErrors.ToErrorCode(err)

	w.Header().Set(headers.ContentType, "application/json")
	responseBytes, _ := json.Marshal(errResponse)
	w.WriteHeader(errResponse.Code())
	return w.Write(responseBytes)
}

func ReturnResponse(w http.ResponseWriter, data any) (int, error) {
	return ReturnResponseWithCode(w, data, http.StatusOK)
}

func ReturnResponseWithCode(w http.ResponseWriter, data any, code int) (int, error) {
	w.Header().Set(headers.ContentType, "application/json")
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			_, _ = ReturnError(w, cplnErrors.NewErrorCode("unable to marshal response body", http.StatusInternalServerError))
		}
		w.WriteHeader(code)
		return w.Write(dataBytes)
	}

	w.WriteHeader(code)
	return 0, nil
}

func ValidateToken(token string) error {
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, nil)
	if vErr, ok := err.(*jwt.ValidationError); ok {
		if vErr.Errors&jwt.ValidationErrorMalformed == jwt.ValidationErrorMalformed {
			return errors.New("token is malformed")
		}
		if vErr.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
			return errors.New("token has expired")
		}
		if vErr.Errors&jwt.ValidationErrorNotValidYet == jwt.ValidationErrorNotValidYet {
			return errors.New("token is not yet valid")
		}
	}
	return nil
}

// AuthorizationHeaders type
type AuthorizationHeaders map[string]string

func GetAuthorizationHeaders(request *http.Request) AuthorizationHeaders {
	hdrs := make(AuthorizationHeaders)
	authzHeader := request.Header[headers.Authorization]
	if len(authzHeader) > 0 && len(authzHeader[0]) > 0 {
		authnTokenParts := strings.Split(authzHeader[0], " ")
		if len(authnTokenParts) != 2 {
			logging.LoggerWithContext(request.Context()).Debugf("Invalid authentication token: %d", len(authnTokenParts))
			return AuthorizationHeaders{}
		}
		switch authnTokenParts[0] {
		case common.AuthnTypeBearer:
			hdrs[headers.Authorization] = authzHeader[0]
		}
	}
	if len(request.Header[common.HeaderIdentityLink]) > 0 {
		hdrs[common.HeaderIdentityLink] = request.Header[common.HeaderIdentityLink][0]
	}
	if len(request.Header[common.HeaderWorkloadLink]) > 0 {
		hdrs[common.HeaderWorkloadLink] = request.Header[common.HeaderWorkloadLink][0]
	}
	return hdrs
}

func Token(request *http.Request) string {
	authHeaders := GetAuthorizationHeaders(request)
	tokenParts := strings.Split(authHeaders[headers.Authorization], " ")
	var token string
	if len(tokenParts) > 1 {
		token = tokenParts[1]
	}
	return token
}
