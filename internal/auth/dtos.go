package auth

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"` // base64(RSA encrypted password)
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type publicKeyResponse struct {
	PublicKey string `json:"public_key"`
}
