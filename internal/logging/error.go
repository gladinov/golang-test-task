package logging

import (
	"context"
	"log/slog"
)

func LoggerError(ctx context.Context, logger *slog.Logger, msg string, op string, err error) {
	logg := logger.With(slog.String("op", op),
		slog.Any("error", err))
	logg.ErrorContext(ctx, msg)
}
