package main

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
