package pubsub

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
)

var (
	PUBLISH = "publish"
	SUBSCRIBE = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type PubSub struct {
	Clients []Client
	Subscriptions []Subscription
}

type Client struct {
	Id string
	Connection *websocket.Conn
}

type Message struct {
	Action string	`json:"action"`
	Topic string	`json:"topic"`
	Message json.RawMessage		`json:"message"`
}

type Subscription struct {
	Topic string
	Client *Client
}

func (ps *PubSub)AddClient(client Client)  (*PubSub) {
	ps.Clients = append(ps.Clients, client)
	//fmt.Println("adding new client to the list", client.Id, len(ps.Clients))
	payload := []byte("Hello Client ID:" +
		client.Id)
	client.Connection.WriteMessage(1, payload)
	return ps
}

func (ps *PubSub)GetSubscriptions(topic string, client *Client) ([]Subscription) {
	var subscriptionList []Subscription
	for _, subscription := range ps.Subscriptions {
		if client != nil{
			if subscription.Client.Id == client.Id && subscription.Topic == topic{
				subscriptionList = append(subscriptionList, subscription)
			}
		}else {
			if subscription.Topic == topic{
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}
	return subscriptionList
}

func (ps *PubSub)Subscribe(client *Client, topic string) (*PubSub) {
	clientSubs := ps.GetSubscriptions(topic, client)
	if len(clientSubs)>0{
		return ps
	}
	newSubscription := Subscription{
		Topic:topic,
		Client:client,
	}
	ps.Subscriptions = append(ps.Subscriptions, newSubscription)
	return ps
}

func (ps *PubSub)Publish(topic string, message []byte, excludeClient *Client)  {
	subscriptions := ps.GetSubscriptions(topic, nil)
	for _, sub := range subscriptions {
		fmt.Printf("Sending to client id %s message is %s", sub.Client.Id, message)
		//sub.Client.Connection.WriteMessage(1, message)
		sub.Client.Send(message)
	}
}

func (client *Client)Send(message []byte) (error) {
	return client.Connection.WriteMessage(1, message)
}

func (ps *PubSub)Unsubscribe(client *Client, topic string) (*PubSub) {
	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {
		if sub.Client.Id == client.Id && sub.Topic == topic{
			// found this subscription from client do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}
	return ps
}

func (ps *PubSub)HandleReceiveMessage(client Client, messageType int, payload []byte) (*PubSub) {
	m := Message{}
	err := json.Unmarshal(payload, &m)
	if err != nil{
		fmt.Println("This is not correct message payload")
		return ps
	}
	fmt.Println("Client correct message payload:", m.Action, m.Message, m.Topic)
	switch m.Action {
	case PUBLISH:
		fmt.Println("This is publish new message")
		ps.Publish(m.Topic, m.Message, nil)
	case SUBSCRIBE:
		fmt.Println("This is subscribe new message")
		ps.Subscribe(&client, m.Topic)
		fmt.Println("new subscriber to topic", m.Topic, len(ps.Subscriptions), client.Id)
	case UNSUBSCRIBE:
		fmt.Println("Client want to unsubscribe the topic", m.Topic, client.Id)
		ps.Unsubscribe(&client, m.Topic)
	default:
		break
	}

	return ps
}