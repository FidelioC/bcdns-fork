// verifier_node.go
// A basic verifier node with pubsub message exchange

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	host "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const TopicName = "verifier-messages"

type VerifierNode struct {
	host  host.Host
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

func NewVerifierNode(ctx context.Context, bootstrap string) *VerifierNode {
	h, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	if bootstrap != "" {
		maddr, err := ma.NewMultiaddr(bootstrap)
		if err != nil {
			log.Fatal(err)
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatal(err)
		}
		h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		h.Connect(ctx, *info)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		log.Fatal(err)
	}
	topic, err := ps.Join(TopicName)
	if err != nil {
		log.Fatal(err)
	}
	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	return &VerifierNode{
		host:  h,
		ps:    ps,
		topic: topic,
		sub:   sub,
	}
}

func (vn *VerifierNode) PrintHostInfo() {
	fmt.Println("Node ID:", vn.host.ID())
	for _, addr := range vn.host.Addrs() {
		fmt.Println("Listening on:", addr.Encapsulate(ma.StringCast("/p2p/"+vn.host.ID().String())))
	}
}

func (vn *VerifierNode) ListenForMessages(ctx context.Context) {
	go func() {
		for {
			msg, err := vn.sub.Next(ctx)
			if msg.ReceivedFrom != vn.host.ID(){
				if err != nil {
					log.Println("Error reading from subscription:", err)
					continue
				}
				fmt.Printf("Received: %s\n", string(msg.Data))
			}
			
		}
	}()
}

func (vn *VerifierNode) SendMessage(ctx context.Context, message string) {
	err := vn.topic.Publish(ctx, []byte(message))
	if err != nil {
		log.Println("Error publishing message:", err)
	}
}

func main() {
	bootstrap := ""
	if len(os.Args) >= 2 {
		bootstrap = strings.TrimSpace(os.Args[1])
	}

	ctx := context.Background()
	vn := NewVerifierNode(ctx, bootstrap)
	vn.PrintHostInfo()
	vn.ListenForMessages(ctx)

	// Simple CLI loop to send messages
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages to send to other nodes:")
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		vn.SendMessage(ctx, fmt.Sprintf("%s: %s", vn.host.ID().ShortString(), text))
		time.Sleep(100 * time.Millisecond)
	}
}
