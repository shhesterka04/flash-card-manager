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

type CardServiceServer struct {
	repo        interfaces.CardRepository
	eventSender kafka.EventSender
	grpc.UnimplementedCardServiceServer
}

func NewCardServiceServer(r interfaces.CardRepository, eventSender kafka.EventSender) *CardServiceServer {
	return &CardServiceServer{repo: r, eventSender: eventSender}
}

func (s *CardServiceServer) CreateCard(ctx context.Context, req *grpc.CreateCardRequest) (*grpc.CardResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CreateCard")
	defer span.Finish()

	if req.Front == "" || req.Back == "" {
		return nil, status.Error(codes.InvalidArgument, "Missing required field")
	}

	card := structs.Card{
		Front:  req.Front,
		Back:   req.Back,
		DeckID: req.DeckId,
		Author: req.Author,
	}

	id, err := s.repo.Add(ctx, card)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to add card")
	}

	fullCard, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve card after creation")
	}

	logger.Infof(ctx, "Вызов SendSyncMessage с параметрами: %v", req)
	if err := s.eventSender.SendEvent("CreateCard", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &grpc.CardResponse{
		Id:        fullCard.ID,
		Front:     fullCard.Front,
		Back:      fullCard.Back,
		DeckId:    fullCard.DeckID,
		Author:    fullCard.Author,
		CreatedAt: fullCard.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CardServiceServer) GetCardById(ctx context.Context, req *grpc.GetCardByIdRequest) (*grpc.CardResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetCardById")
	defer span.Finish()

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
	}

	card, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err.Error() == "card not found" {
			return nil, status.Error(codes.NotFound, "Card not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := s.eventSender.SendEvent("GetCardById", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &grpc.CardResponse{
		Id:        card.ID,
		Front:     card.Front,
		Back:      card.Back,
		DeckId:    card.DeckID,
		Author:    card.Author,
		CreatedAt: card.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CardServiceServer) UpdateCard(ctx context.Context, req *grpc.UpdateCardRequest) (*grpc.CardResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UpdateCard")
	defer span.Finish()

	if req.Id == 0 || req.Front == "" || req.Back == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid request payload")
	}

	card := structs.Card{
		ID:     req.Id,
		Front:  req.Front,
		Back:   req.Back,
		DeckID: req.DeckId,
		Author: req.Author,
	}

	updatedRows, err := s.repo.Update(ctx, card)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if updatedRows == 0 {
		return nil, status.Error(codes.NotFound, "Card not found")
	}

	if err := s.eventSender.SendEvent("UpdateCard", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &grpc.CardResponse{
		Id:     card.ID,
		Front:  card.Front,
		Back:   card.Back,
		DeckId: card.DeckID,
		Author: card.Author,
	}, nil
}

func (s *CardServiceServer) DeleteCard(ctx context.Context, req *grpc.DeleteCardRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "DeleteCard")
	defer span.Finish()

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid ID format")
	}

	if err := s.repo.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, "Failed to delete card")
	}

	if err := s.eventSender.SendEvent("DeleteCard", req.String()); err != nil {
		logger.Errorf(ctx, "Failed to send event to Kafka: %s", err)
	}

	return &emptypb.Empty{}, nil
}
