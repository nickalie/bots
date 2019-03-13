package bots

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nickalie/bots/utils"
	"github.com/parnurzeal/gorequest"
)

const UserAgent = "Microsoft-BotFramework/3.1 (MSBot Golang)"

type MSBotEndpoint struct {
	RefreshEndpoint            string
	RefreshScope               string
	BotConnectorOpenIdMetadata string
	BotConnectorIssuers        []string
	BotConnectorAudience       string
	EmulatorOpenIdMetadata     string
	EmulatorIssuers            []string
	EmulatorAudience           string
	StateEndpoint              string
}

type MSBotSettings struct {
	AppId            string
	AppPassword      string
	GzipData         bool
	Endpoint         *MSBotEndpoint
	StateEndpoint    string
	OpenIdMetadata   string
	ValidateRequests bool
	Channels         []string
}

type MSBot struct {
	settings                   *MSBotSettings
	botConnectorOpenIdMetadata *OpenIdMetadata
	emulatorOpenIdMetadata     *OpenIdMetadata
	accessToken                string
	accessTokenExpires         int64
	updatesChannel             chan *Activity
}

func NewMSBot(settings *MSBotSettings) *MSBot {
	if settings == nil {
		settings = &MSBotSettings{}
	}

	if settings.Endpoint == nil {

		openIdMetadata := settings.OpenIdMetadata

		if openIdMetadata == "" {
			openIdMetadata = "https://login.botframework.com/v1/.well-known/openidconfiguration"
		}

		stateEndpoint := settings.StateEndpoint

		if stateEndpoint == "" {
			stateEndpoint = "https://state.botframework.com"
		}

		settings.Endpoint = &MSBotEndpoint{
			RefreshEndpoint:            "https://login.microsoftonline.com/botframework.com/oauth2/v2.0/token",
			RefreshScope:               "https://api.botframework.com/.default",
			BotConnectorOpenIdMetadata: openIdMetadata,
			BotConnectorIssuers:        []string{"https://api.botframework.com"},
			BotConnectorAudience:       settings.AppId,
			EmulatorOpenIdMetadata:     "https://login.microsoftonline.com/botframework.com/v2.0/.well-known/openid-configuration",
			EmulatorAudience:           settings.AppId,
			EmulatorIssuers: []string{"https://sts.windows.net/d6d49420-f39b-4df7-a1dc-d59a935871db/",
				"https://login.microsoftonline.com/d6d49420-f39b-4df7-a1dc-d59a935871db/v2.0",
				"https://sts.windows.net/f8cdef31-a31e-4b4a-93e4-5f571e91255a/",
				"https://login.microsoftonline.com/f8cdef31-a31e-4b4a-93e4-5f571e91255a/v2.0"},
			StateEndpoint: stateEndpoint,
		}
	}

	return &MSBot{
		settings:                   settings,
		botConnectorOpenIdMetadata: NewOpenIdMetadata(settings.Endpoint.BotConnectorOpenIdMetadata),
		emulatorOpenIdMetadata:     NewOpenIdMetadata(settings.Endpoint.EmulatorOpenIdMetadata),
		updatesChannel:             make(chan *Activity),
	}
}

func (b *MSBot) GetFile(attachment *Attachment, activity *Activity) (*http.Response, error) {
	return b.authenticatedRequest(gorequest.New().Get(attachment.ContentUrl), false)
}

func (b *MSBot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		errorResponse(w, "unsupported method: "+r.Method)
		return
	}

	defer r.Body.Close()
	incoming := Activity{}
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&incoming)

	if b.settings.ValidateRequests {
		isEmulator := incoming.ChannelId == "emulator"
		var token string
		authHeaderValue := r.Header.Get("authorization")

		if authHeaderValue == "" {
			authHeaderValue = r.Header.Get("Authorization")
		}

		if authHeaderValue != "" {
			auth := strings.Split(strings.Trim(authHeaderValue, " "), " ")

			if len(auth) == 2 && strings.ToLower(auth[0]) == "bearer" {
				token = auth[1]
			}
		}

		if token != "" {
			if !b.validateToken(token, &incoming, isEmulator, w) {
				return
			}
		} else if isEmulator && b.settings.AppId != "" && b.settings.AppPassword != "" {
			errorResponse(w, "invalid token")
			return
		}
	}

	b.updatesChannel <- &incoming
	w.WriteHeader(http.StatusOK)
}

func (b *MSBot) validateToken(token string, incoming *Activity, isEmulator bool, w http.ResponseWriter) bool {
	var openIDMetadata *OpenIdMetadata

	if isEmulator {
		openIDMetadata = b.emulatorOpenIdMetadata
	} else {
		openIDMetadata = b.botConnectorOpenIdMetadata
	}

	decoded, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		kid := utils.GetString(token.Header, "kid")
		return openIDMetadata.GetKey(kid)
	})

	if err != nil {
		fmt.Printf("jwt.Parse: %v\n", err)
		errorResponse(w, "invalid token")
		return false
	}

	if decoded == nil {
		fmt.Println("decoded token is null")
		errorResponse(w, "invalid token")
		return false
	}

	claims, ok := decoded.Claims.(jwt.MapClaims)

	if !ok {
		fmt.Println("unable to get claims")
		errorResponse(w, "invalid token")
		return false
	}

	var issuers []string

	if isEmulator {
		issuers = b.settings.Endpoint.EmulatorIssuers
	} else {
		issuers = b.settings.Endpoint.BotConnectorIssuers
	}

	validIssuer := false

	for _, v := range issuers {
		if claims.VerifyIssuer(v, true) {
			validIssuer = true
			break
		}
	}

	if !validIssuer {
		fmt.Println("invalid issuer")
		errorResponse(w, "invalid token")
		return false
	}

	if !claims.VerifyAudience(b.settings.AppId, true) {
		fmt.Println("invalid audience")
		errorResponse(w, "invalid token")
		return false
	}

	if !isEmulator && utils.GetString(claims, "serviceurl") != incoming.ServiceUrl {
		fmt.Println("invalid serviceUrl")
		errorResponse(w, "invalid token")
		return false
	}

	return true
}

func errorResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func (b *MSBot) Send(activity *Activity) (*Identification, error) {
	activity = fixActivity(activity)
	path := "v3/conversations/" + url.QueryEscape(activity.Conversation.Id) + "/activities"

	if activity.ReplyToId != "" {
		path += "/" + url.QueryEscape(activity.ReplyToId)
	}

	request := gorequest.New().Post(activity.ServiceUrl + "/" + path).Send(activity)
	resp, err := b.authenticatedRequest(request, false)

	if err != nil {
		return nil, err
	}

	result := &Identification{}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	err = d.Decode(result)
	return result, err
}

func (b *MSBot) Update(activity *Activity) (*Identification, error) {
	activity = fixActivity(activity)
	path := "/v3/conversations/" + url.QueryEscape(activity.Conversation.Id) + "/activities"

	if activity.Id != "" {
		path += "/" + url.QueryEscape(activity.Id)
	}

	request := gorequest.New().Put(activity.ServiceUrl + "/" + path).Send(activity)
	resp, err := b.authenticatedRequest(request, false)

	if err != nil {
		return nil, err
	}

	result := &Identification{}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	err = d.Decode(result)
	return result, err
}

func (b *MSBot) Delete(activity *Activity) error {
	path := "/v3/conversations/" + url.QueryEscape(activity.Conversation.Id) + "/activities"

	if activity.Id != "" {
		path += "/" + url.QueryEscape(activity.Id)
	}

	request := gorequest.New().Delete(activity.ServiceUrl + "/" + path).Send(activity)
	_, err := b.authenticatedRequest(request, false)
	return err
}

func (b *MSBot) authenticatedRequest(request *gorequest.SuperAgent, refresh bool) (*http.Response, error) {
	if refresh {
		b.accessToken = ""
	}

	b.addUserAgent(request)
	err := b.addAccessToken(request)

	if err != nil {
		return nil, err
	}

	resp, body, errs := request.End()

	if len(errs) > 0 {
		errs = append(errs, errors.New(string(body)))
		return resp, utils.ErrorFromArray(errs)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		if !refresh {
			return b.authenticatedRequest(request, true)
		}
	} else if resp.StatusCode < 400 {
		return resp, nil
	}

	return resp, errors.New(fmt.Sprintf("authenticatedRequest failed: %s %d", body, resp.StatusCode))
}

func (b *MSBot) addAccessToken(request *gorequest.SuperAgent) error {
	token, err := b.getAccessToken()

	if err != nil {
		return err
	}

	request.Set("Authorization", "Bearer "+token)
	return nil
}

func (b *MSBot) addUserAgent(request *gorequest.SuperAgent) {
	request.Set("User-Agent", UserAgent)
}

func (b *MSBot) tokenExpired() bool {
	return time.Now().Unix() >= b.accessTokenExpires
}

func (b *MSBot) tokenHalfWayExpired() bool {
	var secondsToHalfWayExpire int64 = 1800
	var secondsToExpire int64 = 300
	var timeToExpiration = (b.accessTokenExpires - time.Now().Unix()) / 1000
	return timeToExpiration < secondsToHalfWayExpire && timeToExpiration > secondsToExpire
}

func (b *MSBot) getAccessToken() (string, error) {
	if b.accessToken == "" || b.tokenExpired() {
		return b.refreshAccessToken()
	} else if b.tokenHalfWayExpired() {
		oldToken := b.accessToken
		_, err := b.refreshAccessToken()

		if err == nil {
			return b.accessToken, nil
		} else {
			return oldToken, nil
		}
	} else {
		return b.accessToken, nil
	}
}

func (b *MSBot) refreshAccessToken() (string, error) {
	r := gorequest.New().Post(b.settings.Endpoint.RefreshEndpoint)
	r.Type("multipart")
	r.Send(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     b.settings.AppId,
		"client_secret": b.settings.AppPassword,
		"scope":         b.settings.Endpoint.RefreshScope,
	})

	b.addUserAgent(r)
	oauthResponse := OAuthResponse{}
	resp, _, errs := r.EndStruct(&oauthResponse)

	if len(errs) > 0 {
		return "", errs[0]
	}

	if resp.StatusCode >= 300 {
		return "", errors.New(fmt.Sprintf("Refresh access token failed with status code: %d", resp.StatusCode))
	}

	b.accessToken = oauthResponse.AccessToken
	b.accessTokenExpires = time.Now().Unix() - oauthResponse.ExpiresIn + 300
	return b.accessToken, nil
}

func (b *MSBot) GetUpdatesChannel() (<-chan *Activity, error) {
	return b.updatesChannel, nil
}

func (b *MSBot) GetChannels() []string {
	return b.settings.Channels
}

func fixActivity(activity *Activity) *Activity {
	if activity.ChannelId == ChannelLine || activity.ChannelId == "kik" {
		if len(activity.Attachments) > 0 {
			newAttachments := make([]*Attachment, 0)
			for _, a := range activity.Attachments {
				if a.ContentType == TypeHeroCard {
					if card, ok := a.Content.(*HeroCard); ok {
						if len(card.Buttons) > 0 {
							newButtons := make([]*CardAction, 0)
							for _, button := range card.Buttons {
								if button.Type == TypeOpenUrl {
									activity.Text += "<br/><br/>" + button.Title + "\n" + button.Value
								} else {
									newButtons = append(newButtons, button)
								}
							}
							card.Buttons = newButtons
						}
					}
				}

				newAttachments = append(newAttachments, a)
			}

			activity.Attachments = newAttachments
		}
	}

	if activity.ChannelId == "kik" {
		if activity.TextFormat == Markdown {
			activity.Text = strings.Replace(activity.Text, "**", "*", -1)
		}
	}

	return activity
}
