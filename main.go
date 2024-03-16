package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const JOKES_API = "https://official-joke-api.appspot.com/random_joke"

type Joke struct {
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

// access third-party API to get a random joke
func randomJoke() (Joke, error) {
	joke := Joke{}
	resp, err := http.Get(JOKES_API)
	if err != nil {
		fmt.Println("Error accessing API:", err)
		return joke, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading API response:", err)
		return joke, err
	}

	err = json.Unmarshal(body, &joke)
	if err != nil {
		fmt.Println("Error parsing API response:", err)
		return joke, err
	}

	return joke, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	botToken := os.Getenv("BOT_TOKEN")

	// Create a new Discord session
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate event
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is now running. Press Ctrl+C to exit.")

	// Wait for a termination signal (Ctrl+C) to gracefully shutdown the bot
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		fmt.Println("Bot message ignored!")
		return
	}

	if m.Content == "!joke" {
		// Send a reply message
		joke, err := randomJoke()
		if err != nil {
			fmt.Println(err)
			return
		}

		m, err := s.ChannelMessageSend(m.ChannelID, joke.Setup)
		if err != nil {
			fmt.Println(err)
		}

		// wait for 5 seconds
		time.Sleep(5 * time.Second)
		s.ChannelMessageSendReply(m.ChannelID, joke.Punchline, m.Reference())
	}
}
