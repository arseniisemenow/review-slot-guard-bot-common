package external

import (
	"context"
	"fmt"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/ydb"
)

// PopulateProjectFamilies fetches and stores all project families from School 21 API
// This function can be called by both telegram_handler and periodic_job
func PopulateProjectFamilies(ctx context.Context, reviewerLogin string) error {
	tokens, err := ydb.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	client := NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	graph, err := client.GetProjectGraph(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get project graph: %w", err)
	}

	families, err := ExtractFamilies(graph)
	if err != nil {
		return fmt.Errorf("failed to extract families: %w", err)
	}

	err = ydb.UpsertProjectFamilies(ctx, families)
	if err != nil {
		return fmt.Errorf("failed to store project families: %w", err)
	}

	return nil
}
