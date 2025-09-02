// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package controller

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"veloera/common"
	"veloera/model"

	"github.com/gin-gonic/gin"
)

type NodelocUser struct {
	Sub                string   `json:"sub"`
	PreferredUsername  string   `json:"preferred_username"`
	Name               string   `json:"name"`
	Email              string   `json:"email"`
	Picture            string   `json:"picture"`
	EmailVerified      bool     `json:"email_verified"`
	Groups             []string `json:"groups"`
}

func NodelocBind(c *gin.Context) {
	code := c.Query("code")
	nodelocUser, err := getNodelocUserInfoByCode(code, c)
	if err != nil {
		respondWithError(c, http.StatusOK, err.Error())
		return
	}

	oauthUser := &OAuthUser{
		ID:          nodelocUser.Sub,
		Username:    nodelocUser.PreferredUsername,
		DisplayName: nodelocUser.Name,
		Email:       nodelocUser.Email,
		Provider:    "nodeloc",
	}

	config := &OAuthConfig{
		Enabled: common.NodelocOAuthEnabled,
	}

	handleOAuthBind(c, oauthUser, config,
		model.IsNodelocIdAlreadyTaken,
		func(user *model.User) error {
			user.NodelocId = oauthUser.ID
			return user.FillUserByNodelocId()
		},
	)
}

func getNodelocUserInfoByCode(code string, c *gin.Context) (*NodelocUser, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}

	// Get access token using Basic auth
	tokenEndpoint := "https://conn.nodeloc.cc/oauth2/token"
	credentials := common.NodelocClientId + ":" + common.NodelocClientSecret
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))

	// Get redirect URI from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/api/oauth/nodeloc", scheme, c.Request.Host)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to connect to Nodeloc server")
	}
	defer res.Body.Close()

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IdToken      string `json:"id_token"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenRes); err != nil {
		return nil, err
	}

	if tokenRes.AccessToken == "" {
		errorMsg := tokenRes.Error
		if tokenRes.ErrorDesc != "" {
			errorMsg = tokenRes.ErrorDesc
		}
		return nil, fmt.Errorf("failed to get access token: %s", errorMsg)
	}

	// Get user info
	userEndpoint := "https://conn.nodeloc.cc/oauth2/userinfo"
	req, err = http.NewRequest("GET", userEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
	req.Header.Set("Accept", "application/json")

	res2, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to get user info from Nodeloc")
	}
	defer res2.Body.Close()

	var nodelocUser NodelocUser
	if err := json.NewDecoder(res2.Body).Decode(&nodelocUser); err != nil {
		return nil, err
	}

	if nodelocUser.Sub == "" {
		return nil, errors.New("invalid user info returned")
	}

	return &nodelocUser, nil
}

func NodelocOAuth(c *gin.Context) {
	// Handle error from OAuth provider
	errorCode := c.Query("error")
	if errorCode != "" {
		errorDescription := c.Query("error_description")
		respondWithError(c, http.StatusOK, errorDescription)
		return
	}

	code := c.Query("code")
	nodelocUser, err := getNodelocUserInfoByCode(code, c)
	if err != nil {
		respondWithError(c, http.StatusOK, err.Error())
		return
	}

	oauthUser := &OAuthUser{
		ID:          nodelocUser.Sub,
		Username:    nodelocUser.PreferredUsername,
		DisplayName: nodelocUser.Name,
		Email:       nodelocUser.Email,
		Provider:    "nodeloc",
	}

	config := &OAuthConfig{
		Enabled: common.NodelocOAuthEnabled,
	}

	handleOAuthLogin(c, oauthUser, config,
		model.IsNodelocIdAlreadyTaken,
		func(user *model.User) error {
			user.NodelocId = oauthUser.ID
			return user.FillUserByNodelocId()
		},
		createNodelocUser,
	)
}

func createNodelocUser(oauthUser *OAuthUser) (*model.User, error) {
	user := &model.User{
		Username:    oauthUser.Username,
		DisplayName: oauthUser.DisplayName,
		Email:       oauthUser.Email,
		NodelocId:   oauthUser.ID,
	}

	if err := user.Insert(0); err != nil {
		return nil, err
	}

	return user, nil
}