/*
Package jwt provides JSON Web Token (JWT) encoding and decoding utilities.

A JWT is a compact, URL-safe string used to transmit claims between parties.
The common representation is the JWS compact format, which contains three
base64url-encoded segments separated by dots.

  - Header: Describes the token, including "alg" (signing algorithm), "typ"
    (token type, often "JWT"), and "kid" (key identifier).
  - Payload: Contains the claims. Registered claim names include "iss"
    (issuer), "sub" (subject), "aud" (audience), "exp" (expiration time),
    "nbf" (not before), "iat" (issued at), and "jti" (token ID).
  - Signature: The cryptographic signature over the ASCII bytes of
    "<base64url(header)>.<base64url(payload)>". The signature provides
    integrity and authenticity, but it does not provide confidentiality.

Base64url uses a URL-safe alphabet and omits padding characters. Because the
payload is not encrypted, it must not contain secrets unless the token is
separately encrypted by another mechanism. Transport security, such as TLS,
is still required.

This package focuses on signed JWTs and exposes the decoded header and claims
to callers. Callers are responsible for validating claim semantics, such as
audience checks and time-based claim evaluation, according to their needs.

This package currently supports the EdDSA (Ed25519) signing algorithm and
intentionally avoids symmetric algorithms to reduce algorithm-confusion risk.

The following illustrates the compact format structure:

	header.payload.signature

Each segment is base64url-encoded JSON. For example, a header might include
{"alg":"EdDSA","typ":"JWT","kid":"example"} and a payload might include
{"sub":"user-123","exp":1700000000}.
*/
package jwt
