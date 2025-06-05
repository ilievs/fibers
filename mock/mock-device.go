package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func main() {
	// App will run until cancelled by user (e.g. ctrl-c)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// We will connect to the Eclipse test server (note that you may see messages that other users publish)
	u, err := url.Parse("mqtt://localhost:1883")
	if err != nil {
		panic(err)
	}

	deviceName := "psu1"
	commandTopic := "devices/" + deviceName + "/command"
	stateTopic := "devices/" + deviceName + "/state"

	cliCfg := autopaho.ClientConfig{
		ConnectUsername: "psu1",
		ConnectPassword: []byte("password1"),
		ServerUrls: []*url.URL{u},
		KeepAlive:  20, // Keepalive message should be sent every 20 seconds
		// CleanStartOnInitialConnection defaults to false. Setting this to true will clear the session on the first connection.
		CleanStartOnInitialConnection: false,
		// SessionExpiryInterval - Seconds that a session will survive after disconnection.
		// It is important to set this because otherwise, any queued messages will be lost if the connection drops and
		// the server will not queue messages while it is down. The specific setting will depend upon your needs
		// (60 = 1 minute, 3600 = 1 hour, 86400 = one day, 0xFFFFFFFE = 136 years, 0xFFFFFFFF = don't expire)
		SessionExpiryInterval: 60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
			// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
			// the connection drops)
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: commandTopic, QoS: 1},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			}
			fmt.Println("mqtt subscription made")
		},
		OnConnectError: func(err error) {
			 fmt.Printf("error whilst attempting connection: %s\n", err)
		},
		// eclipse/paho.golang/paho provides base mqtt functionality, the below config will be passed in for each connection
		ClientConfig: paho.ClientConfig{
			// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
			ClientID: "psu1",
			// OnPublishReceived is a slice of functions that will be called when a message is received.
			// You can write the function(s) yourself or use the supplied Router
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					fmt.Printf("received message on topic %s; body: %s (retain: %t)\n", pr.Packet.Topic, pr.Packet.Payload, pr.Packet.Retain)
					return true, nil
				}},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	for {
		c, err := autopaho.NewConnection(ctx, cliCfg) // starts process; will reconnect until context cancelled
		if err != nil {
			panic(err)
		}
		// Wait for the connection to come up
		if err = c.AwaitConnection(ctx); err != nil {
			panic(err)
		}

		state := make(map[string]string)
		ticker := time.NewTicker(time.Second)
		msgCount := 0
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				msgCount++

				state["voltage"] = strconv.Itoa(rand.Intn(200) + 50)
				state["current"] = strconv.Itoa(rand.Intn(5) + 1)

				payload, err := json.Marshal(state)
				if err != nil {
					log.Println("Failed to convert state", state, "to string")
					continue
				}

				// Publish a test message (use PublishViaQueue if you don't want to wait for a response)
				_, err = c.Publish(ctx, &paho.Publish{
					QoS:     1,
					Topic:   stateTopic,
					Payload: payload,
				});
				
				if err != nil {
					if ctx.Err() == nil {
						// Publish will exit the loop when context cancelled or if something went wrong
						// and we will reconnect
						continue
					}
				}
				
				log.Println("published state:", state)
				
				continue
			case <-ctx.Done():
			}
			break
		}
	}
}
