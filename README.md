# Go Chat Machines

[![codecov](https://codecov.io/gh/gin-gonic/gin/branch/master/graph/badge.svg)](https://codecov.io/gh/gin-gonic/gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/tomaszbk/chat-machines-go)](https://goreportcard.com/report/github.com/tomaszbk/chat-machines-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/tomaszbk/chat-machines-go?status.svg)](https://pkg.go.dev/github.com/tomaszbk/chat-machines-go?tab=doc)
[![Release](https://img.shields.io/github/release/chat-machines-go/gin.svg?style=flat-square)](https://github.com/tomaszbk/chat-machines-go/releases)

chatmachine is a web framework written in [Go](https://go.dev/).
Small, simple library for building chatbots using state machines, that can be integrated with REST APIs, discord, whatsapp, telegram, etc.

## Installation

With [Go's module support](https://go.dev/wiki/Modules#how-to-use-modules), `go [build|run|test]` automatically fetches the necessary dependencies when you add the import in your code:

```sh
import "github.com/tomaszbk/chat-machines-go"
```

Alternatively, use `go get`:

```sh
go get -u github.com/tomaszbk/chat-machines-go
```

## Usage

```go
package example

import (
	"chatmachine"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// HelloState demonstrates a simple state that transitions to WorldState.
type HelloState struct{}

func (h *HelloState) OnEnter(s *chatmachine.SessionState) {
	s.AddOutput("Entered Hello State")
}

func (h *HelloState) OnUpdate(s *chatmachine.SessionState) {
	if s.Input == "next" {
		s.ChangeState(&WorldState{})
	} else {
		s.AddOutput("Waiting for 'next' input...")
	}
}

func (h *HelloState) OnExit(s *chatmachine.SessionState) {
	s.AddOutput("Exiting Hello State")
}

// WorldState ends the session after greeting.
type WorldState struct{}

func (w *WorldState) OnEnter(s *chatmachine.SessionState) {
	s.AddOutput("Hello, World!")
	s.End()
}

func (w *WorldState) OnUpdate(s *chatmachine.SessionState) {}

func (w *WorldState) OnExit(s *chatmachine.SessionState) {}

var cm *chatmachine.ChatMachine

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cm = chatmachine.NewChatMachine(&HelloState{})
	// Initialize Discord session
	discordToken := os.Getenv("DISCORD_TOKEN")
	discord, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("error creating discord bot")
	}
	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	output := cm.Run(m.Content, s.State.SessionID)
	s.ChannelMessageSend(m.ChannelID, output)
}
```

## Explanation

The first State defined will automatically be selected as the initial state. The `on_enter` method is called when the state is entered, and the `on_exit` method is called when the state is exited. The `on_update` method is called on every input received in the state after it has been entered once.

The `SessionState` object is passed to the state methods, which contains the input and output attributes. The `input` attribute contains the user input. The `add_output` method is used to append a message to the output. A custom data pydantic object can be set to `session.data` to store any data you want to keep through states.

The `change_state` method is used to change the state of the session. It will automatically trigger the new state's `on_enter` method. The `end` method is used to end the session.

## Planned
- Database integrations for storing sessions.
- More examples for different platforms.
- More personalization options.
- Increase documentation.

## License
This project is licensed under the terms of the MIT license. See the [LICENSE](LICENSE) file for details.