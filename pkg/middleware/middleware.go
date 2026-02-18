package middleware

type Mode int

const (
	ModeDirect Mode = iota
	ModeGRPC
)

const (
	HeaderProvider = "X-Captcha-Provider"
	HeaderToken    = "X-Captcha-Token"
	HeaderRemoteIP = "X-Captcha-Remote-IP"
)
