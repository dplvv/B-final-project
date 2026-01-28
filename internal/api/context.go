package api

import "context"

type contextKey string

const userIDKey contextKey = "userID"

func contextWithUserID(ctx context.Context, userID string) context.Context {
  return context.WithValue(ctx, userIDKey, userID)
}

func userIDFromContext(ctx context.Context) string {
  userID, _ := ctx.Value(userIDKey).(string)
  return userID
}
