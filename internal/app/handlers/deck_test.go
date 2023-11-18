//go:build unit
// +build unit

package handlers

import (
	"context"
	"errors"
	"fmt"
	"homework-3/internal/app/grpc"
	"homework-3/internal/app/handlers/utils"
	"homework-3/internal/infrastructure/kafka"
	mock_kafka "homework-3/internal/infrastructure/kafka/mocks"
	mock_units "homework-3/pkg/repository/interfaces/mocks"
	"homework-3/pkg/repository/structs"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestCreateDeckGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.CreateDeckRequest
		repoErr            error
		repoReturn         int64
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Deck Creation",
			input:              &grpc.CreateDeckRequest{Title: "TestTitle", Description: "TestDescription", Author: "TestAuthor"},
			repoErr:            nil,
			repoReturn:         1,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed Deck Creation",
			input:              &grpc.CreateDeckRequest{Title: "TestTitle", Description: "TestDescription", Author: "TestAuthor"},
			repoErr:            errors.New("some error"),
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.Internal,
			expectProducerCall: false,
		},
		{
			name:               "Invalid Input",
			input:              &grpc.CreateDeckRequest{},
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockDeckRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewDeckServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.input.Title != "" && tt.input.Description != "" && tt.input.Author != "" {
				mockRepo.EXPECT().Add(gomock.Any(), gomock.Any()).Return(tt.repoReturn, tt.repoErr)
			}

			if tt.expectProducerCall && tt.repoErr == nil {
				expectedQuery := fmt.Sprintf("title:\"%s\" description:\"%s\" author:\"%s\"", tt.input.Title, tt.input.Description, tt.input.Author)
				matcher := &utils.GRPCKafkaEventMatcher{
					ExpectedType:  "CreateDeck",
					ExpectedQuery: expectedQuery,
				}
				mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
			}

			resp, err := server.CreateDeck(context.Background(), tt.input)

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.repoReturn, resp.Id)
				assert.Equal(t, tt.input.Title, resp.Title)
				assert.Equal(t, tt.input.Description, resp.Description)
				assert.Equal(t, tt.input.Author, resp.Author)
			}
		})
	}
}

func TestUpdateDeckGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.UpdateDeckRequest
		repoErr            error
		repoReturn         int64
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Deck Update",
			input:              &grpc.UpdateDeckRequest{Id: 1, Title: "UpdatedTitle", Description: "UpdatedDescription", Author: "UpdatedAuthor"},
			repoErr:            nil,
			repoReturn:         1,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Updating Non-Existent Deck",
			input:              &grpc.UpdateDeckRequest{Id: 9999, Title: "UpdatedTitle", Description: "UpdatedDescription", Author: "UpdatedAuthor"},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
		{
			name:               "Invalid Input",
			input:              &grpc.UpdateDeckRequest{},
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockDeckRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewDeckServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.input.Id != 0 && tt.input.Title != "" && tt.input.Description != "" && tt.input.Author != "" {
				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(tt.repoReturn, tt.repoErr)
			}

			if tt.expectProducerCall && tt.repoErr == nil {
				expectedQuery := fmt.Sprintf("id:%d  title:\"%s\"  description:\"%s\"  author:\"%s\"", tt.input.Id, tt.input.Title, tt.input.Description, tt.input.Author)
				matcher := &utils.GRPCKafkaEventMatcher{
					ExpectedType:  "UpdateDeck",
					ExpectedQuery: expectedQuery,
				}
				mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
			}

			resp, err := server.UpdateDeck(context.Background(), tt.input)

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.repoReturn, resp.Id)
				assert.Equal(t, tt.input.Title, resp.Title)
				assert.Equal(t, tt.input.Description, resp.Description)
				assert.Equal(t, tt.input.Author, resp.Author)
			}
		})
	}
}

func TestDeleteDeckGRPC(t *testing.T) {
	tests := []struct {
		name               string
		inputID            int64
		repoErr            error
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Deck Deletion",
			inputID:            1,
			repoErr:            nil,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed Deck Deletion",
			inputID:            1,
			repoErr:            errors.New("internal error"),
			wantErr:            true,
			wantCode:           codes.Internal,
			expectProducerCall: false,
		},
		{
			name:               "Invalid ID",
			inputID:            -1,
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockDeckRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewDeckServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.inputID > 0 {
				mockRepo.EXPECT().Delete(gomock.Any(), tt.inputID).Return(tt.repoErr)
				if tt.expectProducerCall && tt.repoErr == nil {
					expectedQuery := fmt.Sprintf("id:%d", tt.inputID)
					matcher := &utils.GRPCKafkaEventMatcher{
						ExpectedType:  "DeleteDeck",
						ExpectedQuery: expectedQuery,
					}
					mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
				}
			}

			req := &grpc.DeleteDeckRequest{Id: tt.inputID}
			resp, err := server.DeleteDeck(context.Background(), req)

			if tt.wantErr {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestGetDeckByIdGRPC(t *testing.T) {
	tests := []struct {
		name               string
		inputID            int64
		repoReturn         *structs.DeckWithCards
		repoErr            error
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:    "Successful Deck Retrieval",
			inputID: 1,
			repoReturn: &structs.DeckWithCards{
				Deck:  structs.Deck{ID: 1, Title: "TestDeck", Description: "TestDescription", Author: "TestAuthor", CreatedAt: time.Now()},
				Cards: []structs.Card{},
			},
			repoErr:            nil,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Deck Not Found",
			inputID:            1,
			repoReturn:         nil,
			repoErr:            errors.New("deck not found"),
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
		{
			name:               "Invalid ID",
			inputID:            -1,
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockDeckRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewDeckServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.inputID > 0 {
				mockRepo.EXPECT().GetWithCardsByID(gomock.Any(), tt.inputID).Return(tt.repoReturn, tt.repoErr)
				if tt.expectProducerCall && tt.repoErr == nil {
					expectedQuery := fmt.Sprintf("id:%d", tt.inputID)
					matcher := &utils.GRPCKafkaEventMatcher{
						ExpectedType:  "GetDeckById",
						ExpectedQuery: expectedQuery,
					}
					mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
				}
			}

			req := &grpc.GetDeckByIdRequest{Id: tt.inputID}
			resp, err := server.GetDeckById(context.Background(), req)

			if tt.wantErr {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
