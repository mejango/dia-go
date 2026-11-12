package dia

import coremodel "github.com/mejango/dia-go/models/core"

import "github.com/mejango/blossm-toolbox-go/security/token"

type JWT = token.JWT[coremodel.SessionContext]
type Claims = token.Claims[coremodel.SessionContext]
type ES256Signer = token.ES256Signer
type LocalSigner = token.LocalSigner
type KMSClient = token.KMSClient
type KMSSigner = token.KMSSigner

// Strict decoding is shared with token parsing; domain normalization retains it.
func strictJSON(b []byte, dst any) error { return token.StrictJSON(b, dst) }
