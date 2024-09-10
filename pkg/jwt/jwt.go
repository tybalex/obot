package jwt

import "github.com/golang-jwt/jwt/v5"

// yeah, duh, this isn't secure, that's not the point right now.
const secret = "this is secret"

type TokenContext struct {
	RunID          string
	ThreadID       string
	AgentID        string
	WorkflowID     string
	WorkflowStepID string
	Scope          string
}

type TokenService struct {
}

func (t *TokenService) DecodeToken(token string) (*TokenContext, error) {
	tk, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tk.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}
	return &TokenContext{
		RunID:          claims["RunID"].(string),
		ThreadID:       claims["ThreadID"].(string),
		AgentID:        claims["AgentID"].(string),
		Scope:          claims["Scope"].(string),
		WorkflowID:     claims["WorkflowID"].(string),
		WorkflowStepID: claims["WorkflowStepID"].(string),
	}, nil
}

func (t *TokenService) NewToken(context TokenContext) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"RunID":          context.RunID,
		"ThreadID":       context.ThreadID,
		"AgentID":        context.AgentID,
		"Scope":          context.Scope,
		"WorkflowID":     context.WorkflowID,
		"WorkflowStepID": context.WorkflowStepID,
	})
	return token.SignedString([]byte(secret))
}
