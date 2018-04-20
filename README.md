# msbot
Golang Implementation of Microsoft BotFramework

```
bot := msbot.NewBot(&msbot.BotSettings{
  AppId:       "app-id",
  AppPassword: "app-password",
 })
http.Handle("/api/messages", bot)
http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("."))))
go http.ListenAndServe(":3980", nil)

updates := bot.GetUpdatesChannel()

for activity := range updates {
  if activity.Type != "message" {
    continue
  }
  
  response := msbot.Activity{}
  response.Type = "message"
  response.From = activity.Recipient
  response.Recipient = activity.From
  response.Text = "You said: " + activity.Text
  response.ReplyToId = activity.Id
  response.Conversation = activity.Conversation
  response.ServiceUrl = activity.ServiceUrl
  identity, err := bot.Send(&response)
  fmt.Printf("bot.Send: %v, %v\n", identity, err)
}
```
