package handlers

import (
	"context"
	"flash-card-manager/internal/app/grpc"
	"flash-card-manager/internal/infrastructure/kafka"
	"flash-card-manager/pkg/logger"
	"flash-card-manager/pkg/repository/interfaces"
	"flash-card-manager/pkg/repository/structs"
	"time"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DeckServiceServer struct {
	repo        interfaces.DeckRepository
	eventSender kafka.EventSender
	grpc.UnimplementedDeckServiceServer
}

func NewDeckServiceServer(r interfaces.DeckRepository, eventSender kafka.EventSender) *DeckServiceServer {
	return &DeckServiceServer{repo: r, eventSender: eventSender}
}

func (s *DeckServiceServer) CreateDeck(ctx context.Context, req *grpc.CreateDeckRequest) (*grpc.DeckResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateDeck")
	defer span.Finish()

	if req.Title == "" || req.Description == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title, description, and author are required")
	}

	deck := structs.Deck{
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
	}

	id, err := s.repo.Add(ctx, deck)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.eventSender.SendEvent("CreateDeck", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &grpc.DeckResponse{
		Id:          id,
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
	}, nil
}

func (s *DeckServiceServer) GetDeckById(ctx context.Context, req *grpc.GetDeckByIdRequest) (*grpc.DeckWithCardsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetDeckById")
	defer span.Finish()

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID parameter")
	}

	deckWithCards, err := s.repo.GetWithCardsByID(ctx, req.Id)
	if err != nil {
		if err.Error() == "deck not found" {
			return nil, status.Error(codes.NotFound, "Deck not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if deckWithCards == nil {
		return nil, status.Error(codes.NotFound, "Deck not found")
	}

	if err := s.eventSender.SendEvent("GetDeckById", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	deckResponse := &grpc.DeckResponse{
		Id:          deckWithCards.Deck.ID,
		Title:       deckWithCards.Deck.Title,
		Description: deckWithCards.Deck.Description,
		Author:      deckWithCards.Deck.Author,
		CreatedAt:   deckWithCards.Deck.CreatedAt.Format(time.RFC3339),
	}

	var cardResponses []*grpc.CardResponse
	for _, card := range deckWithCards.Cards {
		cardResponses = append(cardResponses, &grpc.CardResponse{
			Id:        card.ID,
			Front:     card.Front,
			Back:      card.Back,
			DeckId:    card.DeckID,
			Author:    card.Author,
			CreatedAt: card.CreatedAt.Format(time.RFC3339),
		})
	}

	return &grpc.DeckWithCardsResponse{
		Deck:  deckResponse,
		Cards: cardResponses,
	}, nil
}

func (s *DeckServiceServer) UpdateDeck(ctx context.Context, req *grpc.UpdateDeckRequest) (*grpc.DeckResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UpdateDeck")
	defer span.Finish()

	if req.Id == 0 || req.Title == "" || req.Description == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid request payload")
	}

	deck := structs.Deck{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
	}

	updatedRows, err := s.repo.Update(ctx, deck)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if updatedRows == 0 {
		return nil, status.Error(codes.NotFound, "Deck not found")
	}

	if err := s.eventSender.SendEvent("UpdateDeck", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &grpc.DeckResponse{
		Id:          deck.ID,
		Title:       deck.Title,
		Description: deck.Description,
		Author:      deck.Author,
	}, nil
}

func (s *DeckServiceServer) DeleteDeck(ctx context.Context, req *grpc.DeleteDeckRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "DeleteDeck")
	defer span.Finish()

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
	}

	err := s.repo.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.eventSender.SendEvent("DeleteDeck", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &emptypb.Empty{}, nil
}
