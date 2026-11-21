package secret

// validate.go attaches validation to the canonical secret types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_secret.go +
// the secret wiring in schema_reflect_iam.go. secret.ts reflects Base.keys + type +
// data; the `data` value is a type-selected union (secret.ts:155
// alternatives().conditional('type', …)) applied here as an object-level Refine on
// Secret, dispatching to the per-type sub-rule ValidationRule methods.

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/nats-io/nkeys"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	reAwsAccessKey = regexp.MustCompile(`^AKIA.*`)
	reArn          = regexp.MustCompile(`^arn:.*`)
	reNatsAccount  = regexp.MustCompile(`^A[A-Z0-9]{55}$`)
	reNatsSeed     = regexp.MustCompile(`^SA[A-Z0-9]{56}$`)
	reAccountID    = regexp.MustCompile(`[0-9]{12}`)         // ECR account segment (secret.ts:69)
	reDictKey      = regexp.MustCompile(`^[-._a-zA-Z0-9]+$`) // Dictionary key pattern (secret.ts:124)
)

// knownRegions is the ECR region allowlist (secret.ts:30-48). A repo whose region
// segment is outside this set is rejected.
var knownRegions = map[string]bool{
	"ap-northeast-1": true, "ap-northeast-2": true, "ap-northeast-3": true,
	"ap-south-1": true, "ap-southeast-1": true, "ap-southeast-2": true,
	"ca-central-1": true, "eu-central-1": true, "eu-north-1": true,
	"eu-west-1": true, "eu-west-2": true, "eu-west-3": true, "sa-east-1": true,
	"us-east-1": true, "us-east-2": true, "us-west-1": true, "us-west-2": true,
}

// ValidationRule reflects Base.keys + type + data (secret.ts:170) and layers the
// data-by-type union as an object-level Refine (secret.ts:155). This is the schema
// the "secret" kind validates against.
func (Secret) ValidationRule() *validate.Rule {
	return validate.Reflect(Secret{}).CustomTyped(secretDataRefine)
}

// secretDataRefine ports `data: SecretData = Joi.alternatives().conditional('type',
// [...])` (secret.ts:155): the `type` scalar selects the sub-rule the `data` value is
// validated against; its normalized output (e.g. defaulted encoding, sorted/deduped
// ecr repos, base64 check) is written back. An unmatched type leaves data
// unconstrained (Joi's conditional with no matching branch and no otherwise).
func secretDataRefine(sec any, h *validate.Helper) (any, error) {
	var rule *validate.Rule
	switch validate.Str(sec, "type") {
	case "opaque":
		rule = Opaque{}.ValidationRule()
	case "tls":
		rule = Tls{}.ValidationRule()
	case "gcp", "docker", "azure-sdk":
		rule = validate.String()
	case "aws":
		rule = AwsKey{}.ValidationRule()
	case "ecr":
		rule = EcrPull{}.ValidationRule()
	case "userpass":
		rule = UsernamePassword{}.ValidationRule()
	case "keypair":
		rule = KeyPair{}.ValidationRule()
	case "dictionary":
		rule = Dictionary(nil).ValidationRule()
	case "azure-connector":
		rule = AzureConnector{}.ValidationRule()
	case "nats-account":
		rule = NatsAccount{}.ValidationRule()
	default:
		return sec, nil
	}
	data := validate.Get(sec, "data")
	if data == nil {
		if !dataPresent(sec, h) {
			return sec, nil
		}
		path := append(append([]string(nil), h.Path...), "data")
		if t := validate.Str(sec, "type"); t == "gcp" || t == "docker" || t == "azure-sdk" {
			err := &validate.Error{Type: "string.base", Message: `"data" must be a string`, Path: path}
			return nil, h.WithCtx(err, "value", nil)
		}
		err := &validate.Error{Type: "object.base", Message: `"data" must be of type object`, Path: path}
		return nil, h.WithCtx(err, "type", "object", "value", nil)
	}
	// Reported UNDER `data`: the sub-rule is validated standalone, so without the base
	// path every message read `"value" …` and named nothing the caller could find. TS
	// gets the prefix for free — the alternatives live at the `data` key of the secret
	// schema, so Joi is already walking that path when the branch runs.
	out, verr := rule.ValidateAt([]string{"data"}, data, h.Ctx, validate.Prefs{})
	if verr != nil {
		return nil, verr
	}
	validate.Set(sec, "data", out)
	return sec, nil
}

// base64StringRule ports `Joi.string().base64()` — a padded standard base64 string.
//
// It is the engine's rule, not a decode attempt: Go's base64.StdEncoding IGNORES \r and
// \n by design, so a payload with a trailing newline decoded cleanly and was stored,
// where Joi's regex anchors the whole string and rejects it. The engine already carries
// Joi's pattern, and it labels the error with the FIELD's name — the same rule is
// reachable from more than one field, and a message naming a field the caller did not
// write is worse than no message.
func base64StringRule() *validate.Rule { return validate.String().Base64() }

// ValidationRule ports `Opaque` (secret.ts:6).
func (Opaque) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"payload", "encoding"}, map[string]*validate.Rule{
		"payload": validate.Alternatives().Conditional("encoding", validate.AltCase{
			// `.Required()` because secret.ts:8 spells a LITERAL `is: 'base64'`, which
			// Joi compiles to `valid('base64').required()`. Without it an ABSENT encoding
			// matched, and the payload was held to base64 before `encoding` defaulted to
			// "plain". Alphabetical key ordering used to hide this by defaulting encoding
			// first; declaring Joi's real order exposed it.
			Is:        validate.String().Valid("base64").Required(),
			Then:      base64StringRule().Required(),
			Otherwise: validate.String().Required(),
		}),
		"encoding": validate.String().Valid("plain", "base64").Default("plain").Optional(),
	})
}

// ValidationRule ports `AwsKey` (secret.ts:20).
func (AwsKey) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"accessKey", "secretKey", "roleArn", "externalId"}, map[string]*validate.Rule{
		"accessKey":  validate.String().Required().Regex(reAwsAccessKey),
		"secretKey":  validate.String().Required(),
		"roleArn":    validate.String().Regex(reArn),
		"externalId": validate.String(),
	})
}

// ecrRepoCustom ports the EcrPull.repos item custom (secret.ts:59): each entry
// must be an ECR hostname (ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com[/REPO]); the
// normalized value is the bare hostname or hostname+path.
func ecrRepoCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	u, err := url.Parse("https://" + s)
	invalid := func() (any, error) {
		return nil, h.Errorf("any.custom", "%s is not a valid ECR repo. A valid repo looks like ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/REPO", s)
	}
	if err != nil {
		return invalid()
	}
	// WHATWG URL (what TS parses with) lowercases the host; Go's url.Hostname keeps the
	// case it was given, so an uppercase `…DKR.ecr.US-EAST-2…` failed the literal
	// comparisons below and was refused where TS accepts it and stores it lowercased.
	host := strings.ToLower(u.Hostname())
	parts := strings.Split(host, ".")
	valid := len(parts) == 6 &&
		reAccountID.MatchString(parts[0]) &&
		parts[1] == "dkr" && parts[2] == "ecr" &&
		knownRegions[parts[3]] && parts[4] == "amazonaws" && parts[5] == "com"
	if !valid {
		return invalid()
	}
	if u.Path == "" || u.Path == "/" {
		return host, nil
	}
	return host + u.Path, nil
}

// ValidationRule ports `EcrPull` (secret.ts:49).
func (EcrPull) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"accessKey", "secretKey", "roleArn", "externalId", "repos"}, map[string]*validate.Rule{
		"accessKey":  validate.String().Required().Regex(reAwsAccessKey),
		"secretKey":  validate.String().Required(),
		"roleArn":    validate.String().Regex(reArn),
		"externalId": validate.String(),
		"repos": validate.Array(validate.String().Custom(ecrRepoCustom)).
			Sort().Dedupe().Max(20).Min(1),
	})
}

// ValidationRule ports `UsernamePassword` (secret.ts:92).
func (UsernamePassword) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"username", "password", "encoding"}, map[string]*validate.Rule{
		"username": validate.String().Required(),
		"password": validate.String().Required(),
		"encoding": validate.String().Valid("plain", "base64").Default("plain"),
	})
}

// ValidationRule ports `AzureConnector` (secret.ts:99).
func (AzureConnector) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"url", "code"}, map[string]*validate.Rule{
		// `.uri({scheme:'https'})`. The engine carries Joi's grammar and its
		// `string.uriCustomScheme` envelope; this was a url.Parse custom, which accepted
		// `https://` with no host and other shapes Joi rejects.
		"url":  validate.String().URIScheme("https").Required(),
		"code": validate.String().Required(),
	})
}

// ValidationRule ports `Tls` (secret.ts:110).
func (Tls) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"key", "cert", "chain"}, map[string]*validate.Rule{
		"key":   validate.String(),
		"cert":  validate.String().Required(),
		"chain": validate.String().Optional(),
	})
}

// ValidationRule ports `KeyPair` (secret.ts:117).
func (KeyPair) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"secretKey", "publicKey", "passphrase"}, map[string]*validate.Rule{
		"secretKey":  validate.String().Required(),
		"publicKey":  validate.String(),
		"passphrase": validate.String(),
	})
}

// ValidationRule ports `Dictionary` (secret.ts:124).
func (Dictionary) ValidationRule() *validate.Rule {
	return validate.Object(nil).Pattern(validate.String().Regex(reDictKey), validate.String().Allow(""))
}

// ValidationRule ports `NatsAccount` (secret.ts:127): the regex + nkeys structural
// checks on the account public key and the account seed.
func (NatsAccount) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"accountId", "privateKey"}, map[string]*validate.Rule{
		"accountId": validate.String().Regex(reNatsAccount).Required().Custom(func(value any, h *validate.Helper) (any, error) {
			s, _ := value.(string)
			if _, err := nkeys.FromPublicKey(s); err != nil {
				// TS interpolates the thrown JS error, whose toString() carries the
				// class name: "NKeysError: nkeys: invalid checksum".
				return nil, h.Errorf("any.custom", "NKeysError: %s", err.Error())
			}
			return s, nil
		}),
		"privateKey": validate.String().Regex(reNatsSeed).Required().Custom(func(value any, h *validate.Helper) (any, error) {
			s, _ := value.(string)
			if _, err := nkeys.FromSeed([]byte(s)); err != nil {
				return nil, h.Errorf("any.custom", "NKeysError: %s", err.Error())
			}
			return s, nil
		}),
	})
}

// dataPresent tells a `data: null` body from one without data.
func dataPresent(sec any, h *validate.Helper) bool {
	if m, ok := sec.(map[string]any); ok {
		_, present := m["data"]
		return present
	}
	return h.Present("data")
}
