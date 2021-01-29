package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

func (ps *PubSub) AddClient(client Client) *PubSub {

	ps.Clients = append(ps.Clients, client)
	//fmt.Println("Adding new client to list", client.Id)
	payload := []byte("Hello Client ID" + client.Id)
	client.Connection.WriteMessage(1, payload)
	return ps
}

func (ps *PubSub) RemoveClient(client Client) *PubSub {

	//firsr remove all subsc by this client
	for index, sub := range ps.Subscriptions {
		if client.Id == sub.Client.Id {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}
	// remove client from pubsub
	for index, c := range ps.Clients {
		if c.Id == client.Id {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}
	return ps
}

func (ps *PubSub) GetSubscriptions(topic string, client *Client) []Subscription {

	var subscriptionList []Subscription

	for _, subscription := range ps.Subscriptions {
		if client != nil {
			if subscription.Client.Id == client.Id && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		} else {
			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}

	return subscriptionList
}

func (ps *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {
		// client is subscribed to this topic b4
		return ps
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)

	return ps
}

func (ps *PubSub) Publish(topic string, message []byte, excludeClient *Client) {

	subscriptions := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {
		fmt.Println("Sending to client id %s message is %s \n", sub.Client.Id, message)
		//sub.Client.Connection.WriteMessage(1, message)
		sub.Client.Send(message)
	}
}

func (client *Client) Send(message []byte) error {
	return client.Connection.WriteMessage(1, message)
}

func (ps *PubSub) Unsubscribe(client *Client, topic string) *PubSub {
	//clientSubscriptions := ps.GetSubscriptions(topic, client)
	for index, sub := range ps.Subscriptions {
		if sub.Client.Id == client.Id && sub.Topic == topic {
			ps.Subscriptions = append(ps.Subscriptions[:index], ps.Subscriptions[index+1:]...)
		}
	}
	return ps
}

func (ps *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte) *PubSub {

	m := Message{}

	err := json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("This is not correct message payload")
		return ps
	}

	switch m.Action {
	case PUBLISH:
		fmt.Println("This is publish new message")
		ps.Publish(m.Topic, m.Message, nil)
		break
	case SUBSCRIBE:
		ps.Subscribe(&client, m.Topic)
		fmt.Println("New subscriber to topic", m.Topic)
		break
	case UNSUBSCRIBE:
		fmt.Println("Client wants to unsubscribe to topic", m.Topic, client.Id)
		ps.Unsubscribe(&client, m.Topic)
		break
	default:
		break
	}

	return ps
}
