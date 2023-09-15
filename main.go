package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func main() {
	// Read command line arguments
	if len(os.Args) < 2 {
		log.Fatal("USAGE: ./dungeons-and-trolls-go-bot API_KEY")
	}
	apiKey := os.Args[1]

	// Initialize the HTTP client and set the base URL for the API
	cfg := swagger.NewConfiguration()
	// TODO: use prod path
	cfg.BasePath = "https://docker.tivvit.cz"

	// Set the X-API-key header value
	ctx := context.WithValue(context.Background(), swagger.ContextAPIKey, swagger.APIKey{Key: apiKey})

	// Create a new client instance
	client := swagger.NewAPIClient(cfg)

	// TODO: remove respawn
	if len(os.Args) > 2 && os.Args[2] == "respawn" {
		respawn(ctx, client)
		return
	}

	// Use the client to make API requests
	gameResp, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsGame(ctx, nil)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
	// fmt.Println("Response:", resp)
	fmt.Println("Running bot ...")
	command := run(gameResp)
	fmt.Printf("Command: %+v\n", command)

	_, httpResp, err = client.DungeonsAndTrollsApi.DungeonsAndTrollsCommands(ctx, *command)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
}

func respawn(ctx context.Context, client *swagger.APIClient) {
	dummyPayload := ctx
	log.Println("Respawning ...")
	_, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsRespawn(ctx, dummyPayload)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
}

func run(state swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsCommandsBatch {
	log.Println("Score:", state.Score)
	log.Println("Character.Money:", state.Character.Money)
	log.Println("CurrentPosition.Level:", state.CurrentPosition.Level)
	log.Println("CurrentPosition.PositionX:", state.CurrentPosition.PositionX)
	log.Println("CurrentPosition.PositionY:", state.CurrentPosition.PositionY)

	var mainHandItem *swagger.DungeonsandtrollsItem
	for _, item := range state.Character.Equip {
		if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
			mainHandItem = &item
			break
		}
	}

	if mainHandItem == nil {
		log.Println("Looking for items to buy ...")
		item := shop(&state)
		if item != nil {
			return &swagger.DungeonsandtrollsCommandsBatch{
				Buy: &swagger.DungeonsandtrollsIdentifiers{Ids: []string{item.Id}},
			}
		}
		log.Println("ERROR: Found no item to buy!")
	}

	log.Println("I like this weapon:", mainHandItem.Name)

	monster := findMonster(&state)

	if monster != nil {
		log.Println("Let's fight!")
		// TODO: Use Skill if monster in range
		return &swagger.DungeonsandtrollsCommandsBatch{
			Move: monster.Position,
		}
	}
	log.Println("No monsters. Let's find stairs ...")

	stairsCoords := findStairs(&state)

	if stairsCoords == nil {
		log.Println("Can't find stairs")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Where are the stairs? I can't find them!",
			},
		}
	}

	// Add seed
	if state.CurrentPosition.Level > 2 {
		rand.Seed(time.Now().UnixNano())
		randomYell := rand.Intn(2)
		var yells []string = []string{
			"I'm so scared!",
			"Help me!",
		}
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: yells[randomYell],
			},
		}
	}

	log.Println("Moving towards stairs ...")
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: stairsCoords,
	}
}

func shop(state *swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsItem {
	shop := state.ShopItems
	for _, item := range shop {
		if item.Price <= state.Character.Money && *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
			log.Println("Chosen item:", item.Name)
			return &item
		}
	}
	return nil
}

func findMonster(state *swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsMapObjects {
	level := state.CurrentPosition.Level
	currentMap := state.Map_.Levels[level]
	for _, object := range currentMap.Objects {
		if object.Monsters != nil {
			return &object
		}
	}
	return nil
}

func findStairs(state *swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsCoordinates {
	level := state.CurrentPosition.Level
	currentMap := state.Map_.Levels[level]
	for _, object := range currentMap.Objects {
		if object.IsStairs {
			return object.Position
		}
	}
	return nil
}
