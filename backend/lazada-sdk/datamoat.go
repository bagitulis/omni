package lazadasdk

import (
	"context"
	"strconv"
	"time"
)

type ComputeRiskParams struct {
	AppName string
	UserID  string
	UserIP  string
	ATI     string
}

func (c *Client) ComputeRisk(ctx context.Context, p ComputeRiskParams) (*RawResponse, error) {
	if p.AppName == "" {
		return nil, ErrBadRequest("appName is required")
	}
	if p.UserID == "" {
		return nil, ErrBadRequest("userId is required")
	}
	if p.UserIP == "" {
		return nil, ErrBadRequest("userIp is required")
	}
	if p.ATI == "" {
		return nil, ErrBadRequest("ati is required")
	}

	params := map[string]string{
		"time":    strconv.FormatInt(time.Now().UnixMilli(), 10),
		"appName": p.AppName,
		"userId":  p.UserID,
		"userIp":  p.UserIP,
		"ati":     p.ATI,
	}

	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/datamoat/compute_risk", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

type LoginParams struct {
	AppName      string
	UserID       string
	TID          string
	UserIP       string
	ATI          string
	LoginResult  string
	LoginMessage string
}

func (c *Client) DataMoatLogin(ctx context.Context, p LoginParams) (*RawResponse, error) {
	if p.AppName == "" {
		return nil, ErrBadRequest("appName is required")
	}
	if p.UserID == "" {
		return nil, ErrBadRequest("userId is required")
	}
	if p.TID == "" {
		return nil, ErrBadRequest("tid is required")
	}
	if p.UserIP == "" {
		return nil, ErrBadRequest("userIp is required")
	}
	if p.ATI == "" {
		return nil, ErrBadRequest("ati is required")
	}
	if p.LoginResult == "" {
		return nil, ErrBadRequest("loginResult is required")
	}
	if p.LoginMessage == "" {
		return nil, ErrBadRequest("loginMessage is required")
	}

	params := map[string]string{
		"time":         strconv.FormatInt(time.Now().UnixMilli(), 10),
		"appName":      p.AppName,
		"userId":       p.UserID,
		"tid":          p.TID,
		"userIp":       p.UserIP,
		"ati":          p.ATI,
		"loginResult":  p.LoginResult,
		"loginMessage": p.LoginMessage,
	}

	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/datamoat/login", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
