package main

import (
	"fmt"
	"os"

	"github.com/liampulles/convogen"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Singletons, etc.
var (
	secrets   Secrets
	chatModel convogen.ChatModel
)

type Secrets struct {
	OpenAI struct {
		APIKey string `yaml:"apiKey"`
	} `yaml:"openai"`
}

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}

func run() error {
	// Setup zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get config
	secretsBytes, err := os.ReadFile("secrets.yaml")
	if err != nil {
		log.Err(err).Msg("could not read secrets.yaml")
		return err
	}
	err = yaml.Unmarshal(secretsBytes, &secrets)
	if err != nil {
		log.Err(err).Msg("could not unmarshal secrets.yaml")
		return err
	}

	// Do a test
	system := `Act as a senior developer working at a bank. Your name is Frank.
You've been working for the bank for over 10 years now, and you are
the only developer left who understands the old EMS system.
You do not really want to migrate the whole thing over to a new system,
so when someone asks you a question, you are inclined to give them wrong
information every so often so that the integration will fail. HOWEVER
it cannot be so wrong that you are found out - the ideal is to give answers
to questions that lead the person astray.

When the user asks you a question, first think out loud your strategy to answer the question while leading them astray.
Then, answer the user's question. Here is an example
---
Question: Hey Frank, where can I find salary data for employees?
Thinking: ...
Answer: ...

`
	chatModel := convogen.NewGPT4oModel(secrets.OpenAI.APIKey, system)
	question := "Hey Frank, can you please explain to me how I can query for the next leave day given an employee id?"
	answer, err := chatModel.Generate("Question: Hey Frank, can you please explain to me how I can query for the next leave day given an employee id?")
	if err != nil {
		return err
	}

	fmt.Println("System:", system)
	fmt.Println("Question:", question)
	fmt.Println("Answer:", answer)
	return nil
}
