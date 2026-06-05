// examples/context — context propagation with tlog.
//
// Shows how to embed a logger in a context.Context so it flows through
// function call chains without explicit parameter threading.
//
// Patterns demonstrated:
//   - NewContext / FromContext round-trip.
//   - WithContext to attach fields to the embedded logger.
//   - Middleware-style pattern: enrich the logger once per request, pass ctx.
//
// Run:
//
//	go run ./examples/context/
package main

import (
	"context"
	"time"

	"github.com/tidjee-dev/tlog"
)

func main() {
	// Root logger — typically created once at program startup.
	root := tlog.New(tlog.WithConsole()).With(
		tlog.String("service", "order-service"),
	)
	defer root.Close()

	// Embed the root logger in a background context.
	ctx := tlog.NewContext(context.Background(), root)

	// Simulate two concurrent requests.
	handleOrder(ctx, "ord-111", "alice")
	handleOrder(ctx, "ord-222", "bob")
}

// handleOrder simulates request middleware: enriches the context logger with
// request-scoped fields, then passes ctx down to sub-functions.
func handleOrder(ctx context.Context, orderID, userID string) {
	// Attach request-scoped fields without mutating the root logger.
	ctx = tlog.WithContext(ctx,
		tlog.String("order_id", orderID),
		tlog.String("user_id", userID),
	)

	tlog.FromContext(ctx).Info("order processing started")

	validateOrder(ctx)
	chargePayment(ctx)
	sendConfirmation(ctx)

	tlog.FromContext(ctx).Info("order processing complete")
}

func validateOrder(ctx context.Context) {
	log := tlog.FromContext(ctx)
	log.Debug("validating order")
	// Simulate work.
	time.Sleep(1 * time.Millisecond)
	log.Debug("order validated")
}

func chargePayment(ctx context.Context) {
	// Enrich the logger further within this function scope.
	ctx = tlog.WithContext(ctx, tlog.String("payment_gateway", "stripe"))
	log := tlog.FromContext(ctx)
	log.Info("charging payment")
	time.Sleep(2 * time.Millisecond)
	log.Info("payment succeeded", tlog.Float64("amount_usd", 49.99))
}

func sendConfirmation(ctx context.Context) {
	log := tlog.FromContext(ctx)
	log.Info("sending confirmation email")
}
