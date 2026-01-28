package api

import (
  "crypto/rand"
  "encoding/base64"
)

func randomToken(size int) (string, error) {
  buffer := make([]byte, size)
  if _, err := rand.Read(buffer); err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(buffer), nil
}
