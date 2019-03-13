# Golang Polyglot Bot Framework

Supported platforms:
* [Microsoft Bot Framework](https://dev.botframework.com):
  * [Telegram](https://telegram.org/)
  * [Skype](https://www.skype.com)
  * [Facebook Messenger](https://www.messenger.com/)
  * [Kik](https://www.kik.com/)
  * [Line](https://line.me)
  * .. and others supported by [Microsoft Bot Framework](https://dev.botframework.com)
* [Viber](https://www.viber.com)   

```
router := mux.NewRouter()
vBot, err := bots.NewViberBot(&bots.ViberBotConfig{
		Token:      "viber-bot-token,
		WebHookURL: "https://your-domain.com/messages/viber",
})

if err != nil {
    log.Fatal(err)
}
	
router.Handle("/messages/viber", vBot)	
	
msBot := msbot.NewBot(&msbot.BotSettings{
  AppId:       "app-id",
  AppPassword: "app-password",
  Channels:    []string{bots.ChannelWebChat, bots.ChannelTelegram, bots.ChannelLine, bots.ChannelSkype, bots.ChannelKik, bots.ChannelFacebook},
  ValidateRequests: true,
})

router.Handle("/messages/ms/", msBot)
go http.ListenAndServe(":80", router)

multiBot = bots.NewMultiBot(vBot, msBot)

activities, err := multiBot.GetUpdatesChannel()

if err != nil {
    log.Fatal(err)
}

for activity := range activities {
  if activity.Type != "message" {
    continue
  }
  
  response := activity.Response("You said: " + activity.Text)
  identity, err := bot.Send(response)
  fmt.Printf("bot.Send: %v, %v\n", identity, err)
}
```
