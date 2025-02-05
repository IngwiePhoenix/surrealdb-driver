package config

type AuthMethod string

const (
	AuthMethodRoot      AuthMethod = "root"
	AuthMethodDB        AuthMethod = "db"
	AuthMethodRecord    AuthMethod = "record"
	AuthMethodUnknown   AuthMethod = "unknown"
	AuthMethodToken     AuthMethod = "token"
	AuthMethodAnonymous AuthMethod = "none"
)
