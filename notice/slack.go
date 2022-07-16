package notice

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
)

var token string
var api *slack.Client

func init() {
	token = os.Getenv("SLACK_TOKEN")
	api = slack.New(token)
}

func SendMessage(textMessage string) (string, string, error) {
	channel, timestamp, err := api.PostMessage(
		"#data2notion",
		slack.MsgOptionText(textMessage, false),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		return "", "", err
	}
	return channel, timestamp, nil
}

func SendMessageWithTS(textMessage string, ts string) (string, error) {
	if len(ts) == 0 {
		return "", nil
	}
	_, newTs, err := api.PostMessage(
		"#data2notion",
		slack.MsgOptionText(textMessage, false),
		slack.MsgOptionTS(ts),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		return "", err
	}
	return newTs, err
}

func UpdateMessage(textMessage string, channel, ts string) (string, error) {
	if len(ts) == 0 {
		return "", nil
	}
	_, newTs, _, err := api.UpdateMessage(
		channel,
		ts,
		slack.MsgOptionText(textMessage, false),
		slack.MsgOptionTS(ts),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		return newTs, err
	}
	return newTs, nil
}

func ReceiveMessage(func(message string)) error {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.DesktopNotificationEvent:
			fmt.Printf("Desktop Notification: %v\n", ev)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			return fmt.Errorf("Invalid credentials")
		default:
		}
	}
	return nil
}

func NewSlackThread(title string, args ...interface{}) *SlackThread {
	thread := &SlackThread{
		title: fmt.Sprintf(title, args...),
	}
	thread.channel, thread.ts, _ = SendMessage(title)
	return thread
}

type SlackThread struct {
	channel  string
	ts       string
	title    string
	hasError bool
}

func (s *SlackThread) Title() string {
	return s.title
}

func (s *SlackThread) Send(message string, args ...interface{}) {
	log.Printf(message, args...)
	_, err := SendMessageWithTS(fmt.Sprintf(message, args...), s.ts)
	if err != nil {
		panic(err)
	}
}

func (s *SlackThread) SendError(err error) {
	log.Printf(err.Error())
	SendMessageWithTS(err.Error(), s.ts)
	s.hasError = true
}

func (s *SlackThread) Update(message string, args ...interface{}) {
	s.ts, _ = UpdateMessage(fmt.Sprintf(message, args...), s.channel, s.ts)
}

func (s *SlackThread) Append(message string, args ...interface{}) {
	s.title = s.title + fmt.Sprintf(message, args...)
	s.Update(s.title)
}

func (s *SlackThread) MakeThread(title string, args ...interface{}) INotifier {
	t := &SlackThread{title: fmt.Sprintf(title, args...), channel: s.channel}
	t.ts, _ = SendMessageWithTS(t.title, s.ts)
	return t
}

func (s *SlackThread) Finish(successTag, failedTag string) {
	if s.hasError {
		s.Append(failedTag)
		return
	}
	s.Append(successTag)
}
