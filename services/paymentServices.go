package services

import (
	"context"
	"fmt"
	"handworks-api/types"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PaymentService) withTx(
	ctx context.Context,
	fn func(pgx.Tx) error,
) (err error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.Logger.Error("rollback failed: %v", rbErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()
	return fn(tx)
}

func (s *PaymentService) GetQuotePrices(ctx context.Context, quoteId string) (*types.CleaningPrices, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var prices *types.CleaningPrices
	if err := s.withTx(dbCtx, func(tx pgx.Tx) error {
		cleaningPrices, err := s.Tasks.VerifyQuoteAndFetchPrices(dbCtx, tx, quoteId)
		if err != nil {
			return err
		}
		prices = cleaningPrices
		return nil
	}); err != nil {
		s.Logger.Error("Failed to Get Quote Prices: %v", err)
		return nil, err
	}
	return prices, nil
}

func (s *PaymentService) MakePublicQuotation(ctx context.Context, req types.QuoteRequest) (*types.QuoteResponse, error) {
	s.Logger.Info("Generating Quote Preview")
	quotePrev, err := s.Tasks.CalculateQuotePreview(ctx, &req)
	if err != nil {
		s.Logger.Error("Failed to genearte Quote Preview: %v", err)
		return nil, fmt.Errorf("failed to genearte Quote Preview: %v", err)
	}
	addonsBreakdown := s.Tasks.MapAddonstoAddonBreakdown(&quotePrev.Addons)
	return &types.QuoteResponse{
		QuoteId:          quotePrev.ID,
		MainServiceName:  quotePrev.MainService,
		MainServiceTotal: quotePrev.Subtotal,
		TotalPrice:       quotePrev.TotalPrice,
		AddonTotal:       quotePrev.AddonTotal,
		Addons:           addonsBreakdown,
	}, nil
}

func (s *PaymentService) MakeQuotation(ctx context.Context, req types.QuoteRequest) (*types.QuoteResponse, error) {
	var quoteResponse types.QuoteResponse
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		quote, err := s.Tasks.CreateQuote(ctx, tx, &req)
		if err != nil {
			return fmt.Errorf("failed to create Quote: %v", err)
		}
		quoteResponse.QuoteId = quote.ID
		quoteResponse.MainServiceName = quote.MainService
		quoteResponse.MainServiceTotal = quote.TotalPrice
		quoteResponse.AddonTotal = quote.AddonTotal
		quoteResponse.TotalPrice = quote.TotalPrice
		quoteResponse.Addons = s.Tasks.MapAddonstoAddonBreakdown(&quote.Addons)
		return nil
	}); err != nil {
		s.Logger.Error("Failed to create Quote: %v", err)
		return nil, err
	}
	return &quoteResponse, nil
}

func (s *PaymentService) GetAllQuotesFromCustomer(
	ctx context.Context,
	customerId, startDate, endDate string,
	page, limit int,
) (*types.FetchAllQuotesResponse, error) {

	var result *types.FetchAllQuotesResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllQuotesByCustomer(
			ctx, tx, customerId, startDate, endDate, page, limit, s.Logger,
		)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch Quotes: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *PaymentService) GetAllQuotes(
    ctx context.Context,
    startDate, endDate string,
    page, limit int,
) (*types.FetchAllQuotesResponse, error) {

    var result *types.FetchAllQuotesResponse

    if err := s.withTx(ctx, func(tx pgx.Tx) error {
        var err error
        result, err = s.Tasks.FetchAllQuotes(ctx, tx, startDate, endDate, page, limit, s.Logger)
        return err
    }); err != nil {
        s.Logger.Error("Failed to fetch Quotes: %v", err)
        return nil, err
    }

    return result, nil
}

func (s *PaymentService) GetQuoteByIDForCustomer(ctx context.Context, quoteId, customerId string) (*types.QuoteResponse, error) {
	var quote *types.QuoteResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		quote, err = s.Tasks.FetchQuoteByIDbyCustomer(ctx, tx, quoteId, customerId)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch quote by ID for customer %s: %v", customerId, err)
		return nil, err
	}

	if quote == nil {
		return nil, fmt.Errorf("quote not found for this customer")
	}

	return quote, nil
}
