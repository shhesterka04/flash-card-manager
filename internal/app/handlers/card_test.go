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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateCardGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.CreateCardRequest
		repoErr            error
		repoReturn         int64
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Creation",
			input:              &grpc.CreateCardRequest{Front: "TestFront", Back: "TestBack", DeckId: 1, Author: "Author"},
			repoErr:            nil,
			repoReturn:         1,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed Creation",
			input:              &grpc.CreateCardRequest{Front: "TestFront", Back: "TestBack", DeckId: 1, Author: "Author"},
			repoErr:            errors.New("some error"),
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.Internal,
			expectProducerCall: false,
		},
		{
			name:               "Empty Input",
			input:              &grpc.CreateCardRequest{},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
		{
			name:               "Missing Required Field",
			input:              &grpc.CreateCardRequest{Front: "TestFront"},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockCardRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewCardServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.input.Front != "" && tt.input.Back != "" {
				mockRepo.EXPECT().Add(gomock.Any(), gomock.Any()).Return(tt.repoReturn, tt.repoErr)
			}

			if tt.expectProducerCall && tt.repoErr == nil {
				expectedQuery := fmt.Sprintf("front:\"%s\"  back:\"%s\"  deck_id:%d  author:\"%s\"", tt.input.Front, tt.input.Back, tt.input.DeckId, tt.input.Author)

				matcher := &utils.GRPCKafkaEventMatcher{
					ExpectedType:  "CreateCard",
					ExpectedQuery: string(expectedQuery),
				}
				mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
			}

			resp, err := server.CreateCard(context.Background(), tt.input)

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
				assert.Equal(t, tt.input.Front, resp.Front)
				assert.Equal(t, tt.input.Author, resp.Author)
				assert.Equal(t, tt.input.Back, resp.Back)
				assert.Equal(t, tt.input.DeckId, resp.DeckId)
			}
		})
	}
}

func TestUpdateCardGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.UpdateCardRequest
		repoErr            error
		repoReturn         int64
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Update",
			input:              &grpc.UpdateCardRequest{Id: 1, Front: "UpdatedFront", Back: "UpdatedBack"},
			repoErr:            nil,
			repoReturn:         1,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed Update - Card Not Found",
			input:              &grpc.UpdateCardRequest{Id: 1, Front: "UpdatedFront", Back: "UpdatedBack"},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
		{
			name:               "Update Non-Existent Card",
			input:              &grpc.UpdateCardRequest{Id: 999, Front: "UpdatedFront", Back: "UpdatedBack"},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
		{
			name:               "Invalid ID",
			input:              &grpc.UpdateCardRequest{Id: -1, Front: "UpdatedFront", Back: "UpdatedBack"},
			repoErr:            nil,
			repoReturn:         0,
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockCardRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewCardServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(tt.repoReturn, tt.repoErr)

			if tt.expectProducerCall && tt.repoErr == nil {
				expectedQuery := fmt.Sprintf("id:%d  front:\"%s\"  back:\"%s\"", tt.input.Id, tt.input.Front, tt.input.Back)

				matcher := &utils.GRPCKafkaEventMatcher{
					ExpectedType:  "UpdateCard",
					ExpectedQuery: expectedQuery,
				}
				mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
			}

			resp, err := server.UpdateCard(context.Background(), tt.input)

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
				assert.Equal(t, tt.input.Front, resp.Front)
				assert.Equal(t, tt.input.Author, resp.Author)
				assert.Equal(t, tt.input.Back, resp.Back)
				assert.Equal(t, tt.input.DeckId, resp.DeckId)
			}
		})
	}
}

func TestDeleteCardGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.DeleteCardRequest
		repoErr            error
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful Deletion",
			input:              &grpc.DeleteCardRequest{Id: 1},
			repoErr:            nil,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed Deletion - Card Not Found",
			input:              &grpc.DeleteCardRequest{Id: 1},
			repoErr:            errors.New("card not found"),
			wantErr:            true,
			wantCode:           codes.Internal,
			expectProducerCall: false,
		},
		{
			name:               "Invalid ID",
			input:              &grpc.DeleteCardRequest{Id: -1},
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockCardRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewCardServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.input.Id > 0 {
				mockRepo.EXPECT().Delete(gomock.Any(), tt.input.Id).Return(tt.repoErr)
			}

			if tt.expectProducerCall && tt.repoErr == nil {
				expectedQuery := fmt.Sprintf("id:%d", tt.input.Id)
				matcher := &utils.GRPCKafkaEventMatcher{
					ExpectedType:  "DeleteCard",
					ExpectedQuery: expectedQuery,
				}
				mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
			}

			resp, err := server.DeleteCard(context.Background(), tt.input)

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
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

func TestGetCardByIdGRPC(t *testing.T) {
	tests := []struct {
		name               string
		input              *grpc.GetCardByIdRequest
		repoReturn         *structs.Card
		repoErr            error
		wantErr            bool
		wantCode           codes.Code
		expectProducerCall bool
	}{
		{
			name:               "Successful GetByID",
			input:              &grpc.GetCardByIdRequest{Id: 1},
			repoReturn:         &structs.Card{ID: 1, Front: "TestFront", Back: "TestBack"},
			repoErr:            nil,
			wantErr:            false,
			wantCode:           codes.OK,
			expectProducerCall: true,
		},
		{
			name:               "Failed GetByID - Card Not Found",
			input:              &grpc.GetCardByIdRequest{Id: 1},
			repoReturn:         nil,
			repoErr:            errors.New("card not found"),
			wantErr:            true,
			wantCode:           codes.NotFound,
			expectProducerCall: false,
		},
		{
			name:               "Invalid ID",
			input:              &grpc.GetCardByIdRequest{Id: -1},
			wantErr:            true,
			wantCode:           codes.InvalidArgument,
			expectProducerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockRepo := mock_units.NewMockCardRepository(mockCtrl)
			mockProducer := mock_kafka.NewMockProducerInterface(mockCtrl)

			server := NewCardServiceServer(mockRepo, kafka.NewKafkaEventSender(mockProducer))

			if tt.input.Id > 0 {
				mockRepo.EXPECT().GetByID(gomock.Any(), tt.input.Id).Return(tt.repoReturn, tt.repoErr)
				if tt.expectProducerCall && tt.repoErr == nil {
					expectedQuery := fmt.Sprintf("id:%d", tt.input.Id)
					matcher := &utils.GRPCKafkaEventMatcher{
						ExpectedType:  "GetCardById",
						ExpectedQuery: expectedQuery,
					}
					mockProducer.EXPECT().SendSyncMessage(matcher).Return(int32(1), int64(1), nil)
				}
			}

			resp, err := server.GetCardById(context.Background(), tt.input)

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.repoReturn.ID, resp.Id)
				assert.Equal(t, tt.repoReturn.Front, resp.Front)
			}
		})
	}
}
