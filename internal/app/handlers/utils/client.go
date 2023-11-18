package utils

import (
	"context"
	"fmt"
	pb "flash-card-manager/internal/app/grpc"
	"flash-card-manager/pkg/logger"
	"strconv"
)

func HandleCommand(ctx context.Context, deckClient pb.DeckServiceClient, cardClient pb.CardServiceClient, cmd string, args []string) error {
	switch cmd {
	case "createDeck":
		return createDeck(ctx, deckClient, args...)
	case "getDeckById":
		return getDeckById(ctx, deckClient, args[0])
	case "updateDeck":
		return updateDeck(ctx, deckClient, args...)
	case "deleteDeck":
		return deleteDeck(ctx, deckClient, args[0])
	case "createCard":
		return createCard(ctx, cardClient, args...)
	case "getCardById":
		return getCardById(ctx, cardClient, args[0])
	case "updateCard":
		return updateCard(ctx, cardClient, args...)
	case "deleteCard":
		return deleteCard(ctx, cardClient, args[0])
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func createDeck(ctx context.Context, client pb.DeckServiceClient, args ...string) error {
	if len(args) != 3 {
		return fmt.Errorf("createDeck requires 3 arguments: title, description, author")
	}
	title, description, author := args[0], args[1], args[2]

	resp, err := client.CreateDeck(ctx, &pb.CreateDeckRequest{
		Title:       title,
		Description: description,
		Author:      author,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to create deck: %v", err)
		return err
	}

	logger.Infof(ctx, "Deck created: %v", resp)
	return nil
}

func getDeckById(ctx context.Context, client pb.DeckServiceClient, deckIdStr string) error {
	deckId, err := strconv.ParseInt(deckIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid deck ID format: %v", err)
	}

	resp, err := client.GetDeckById(ctx, &pb.GetDeckByIdRequest{Id: deckId})
	if err != nil {
		logger.Errorf(ctx, "Failed to get deck: %v", err)
		return err
	}

	logger.Infof(ctx, "Deck retrieved: %v", resp)
	return nil
}

func updateDeck(ctx context.Context, client pb.DeckServiceClient, args ...string) error {
	if len(args) != 4 {
		return fmt.Errorf("updateDeck requires 4 arguments: id, title, description, author")
	}

	deckId, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID: %v", err)
	}
	title, description, author := args[1], args[2], args[3]

	resp, err := client.UpdateDeck(ctx, &pb.UpdateDeckRequest{
		Id:          deckId,
		Title:       title,
		Description: description,
		Author:      author,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to update deck: %v", err)
		return err
	}

	logger.Infof(ctx, "Deck updated: %v", resp)
	return nil
}

func deleteDeck(ctx context.Context, client pb.DeckServiceClient, deckIdStr string) error {
	deckId, err := strconv.ParseInt(deckIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid deck ID format: %v", err)
	}

	_, err = client.DeleteDeck(ctx, &pb.DeleteDeckRequest{Id: deckId})
	if err != nil {
		logger.Errorf(ctx, "Failed to delete deck: %v", err)
		return err
	}

	logger.Info(ctx, "Deck deleted successfully")
	return nil
}

func createCard(ctx context.Context, client pb.CardServiceClient, args ...string) error {
	if len(args) != 4 {
		return fmt.Errorf("createCard requires 4 arguments: front, back, deckId, author")
	}
	front, back, author := args[0], args[1], args[3]
	deckId, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid deck ID format: %v", err)
	}

	resp, err := client.CreateCard(ctx, &pb.CreateCardRequest{
		Front:  front,
		Back:   back,
		DeckId: deckId,
		Author: author,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to create card: %v", err)
		return err
	}

	logger.Infof(ctx, "Card created: %v", resp)
	return nil
}

func getCardById(ctx context.Context, client pb.CardServiceClient, cardIdStr string) error {
	cardId, err := strconv.ParseInt(cardIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid card ID format: %v", err)
	}

	resp, err := client.GetCardById(ctx, &pb.GetCardByIdRequest{Id: cardId})
	if err != nil {
		logger.Errorf(ctx, "Failed to retrieve card: %v", err)
		return err
	}

	logger.Infof(ctx, "Card retrieved: %v", resp)
	return nil
}

func updateCard(ctx context.Context, client pb.CardServiceClient, args ...string) error {
	if len(args) != 5 {
		return fmt.Errorf("updateCard requires 5 arguments: cardId, front, back, deckId, author")
	}

	cardId, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid card ID: %v", err)
	}
	front, back, author := args[1], args[2], args[4]
	deckId, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid deck ID format: %v", err)
	}

	resp, err := client.UpdateCard(ctx, &pb.UpdateCardRequest{
		Id:     cardId,
		Front:  front,
		Back:   back,
		DeckId: deckId,
		Author: author,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to update card: %v", err)
		return err
	}

	logger.Infof(ctx, "Card updated: %v", resp)
	return nil
}

func deleteCard(ctx context.Context, client pb.CardServiceClient, cardIdStr string) error {
	cardId, err := strconv.ParseInt(cardIdStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid card ID format: %v", err)
	}

	_, err = client.DeleteCard(ctx, &pb.DeleteCardRequest{Id: cardId})
	if err != nil {
		logger.Errorf(ctx, "Failed to delete card: %v", err)
		return err
	}

	logger.Info(ctx, "Card deleted successfully")
	return nil
}
