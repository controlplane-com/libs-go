package secret

import (
	"encoding/json"
)

type DockerSecretAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

type DockerSecret struct {
	Auths map[string]DockerSecretAuth `json:"auths"`
}

type GcrSecret struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

func ParseSecretData(s *Secret) error {
	switch s.Type {
	case "docker":
		var d DockerSecret
		err := json.Unmarshal([]byte(s.Data.(string)), &d)
		if err != nil {
			return err
		}
		s.Data = d
		break
	case "gcp":
		var g GcrSecret
		err := json.Unmarshal([]byte(s.Data.(string)), &g)
		if err != nil {
			return err
		}
		s.Data = g
		break
	case "ecr":
		rawData, err := json.Marshal(s.Data)
		if err != nil {
			return err
		}
		var e EcrPull
		err = json.Unmarshal(rawData, &e)
		if err != nil {
			return err
		}
		s.Data = e
		break
	}
	return nil
}

// TODO: Extend the go-converter to generate this method for JOI schemas that use the "alternatives" option.
func UnmarshalSecretJSON(b []byte, s *Secret) error {
	err := json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	var rawData *json.RawMessage
	err = json.Unmarshal(*objMap["data"], &rawData)
	if err != nil {
		return err
	}

	err = ParseSecretData(s)
	if err != nil {
		return err
	}
	return nil
}
