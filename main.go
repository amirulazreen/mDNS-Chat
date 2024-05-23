package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n   Run './MDNS -port [number]. eg. ./MDNS -port 6668 and ./MDNS -port 6666'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := rand.Reader

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))

	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(privKey),
	)
	if err != nil {
		panic(err)
	}

	host.SetStreamHandler(protocol.ID(cfg.ProtocolID), func(stream network.Stream) {
		handleStream(stream, host)
	})

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, host.ID())

	peerChan := initMDNS(host, cfg.RendezvousString)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		<-sigChan
		fmt.Println("Exiting the chat...")
		cancel()
		os.Exit(0)
	}()

	for {
		select {
		case <-ctx.Done():

			return
		case peer := <-peerChan:
			if peer.ID > host.ID() {
				fmt.Println("Found peer:", peer, " id is greater than us, wait for it to connect to us")
				continue
			}
			fmt.Println("Found peer:", peer, ", connecting")

			if err := host.Connect(ctx, peer); err != nil {
				fmt.Println("Connection failed:", err)
				continue
			}

			stream, err := host.NewStream(ctx, peer.ID, protocol.ID(cfg.ProtocolID))
			if err != nil {
				fmt.Println("Stream open failed", err)
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

				go writeData(rw)
				go readData(rw, host)
				fmt.Println("Connected to:", peer)
			}
		}
	}
}