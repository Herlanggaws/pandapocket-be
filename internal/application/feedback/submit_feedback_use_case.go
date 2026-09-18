package feedback

import (
	"context"
	"time"

	domainFeedback "panda-pocket/internal/domain/feedback"
)

type SubmitFeedbackRequest struct {
	Category string `json:"category" binding:"required,oneof=bug suggestion other"`
	Message  string `json:"message" binding:"required,min=1,max=2000"`
}

type FeedbackResponse struct {
	ID        int    `json:"id"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type SubmitFeedbackUseCase struct {
	repo domainFeedback.FeedbackRepository
}

func NewSubmitFeedbackUseCase(repo domainFeedback.FeedbackRepository) *SubmitFeedbackUseCase {
	return &SubmitFeedbackUseCase{repo: repo}
}

func (uc *SubmitFeedbackUseCase) Execute(ctx context.Context, userID int, req SubmitFeedbackRequest) (*FeedbackResponse, error) {
	category, err := domainFeedback.ParseCategory(req.Category)
	if err != nil {
		return nil, err
	}

	item, err := domainFeedback.NewFeedback(userID, category, req.Message)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	return &FeedbackResponse{
		ID:        item.ID().Value(),
		Category:  item.Category().String(),
		Message:   item.Message(),
		CreatedAt: item.CreatedAt().Format(time.RFC3339),
	}, nil
}
