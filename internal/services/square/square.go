package square

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/square/square-go-sdk"
	"github.com/square/square-go-sdk/checkout"
	"github.com/square/square-go-sdk/client"
	"github.com/square/square-go-sdk/option"
	"net/http"
	"strconv"
)

func GetSquareClient() (*client.Client, *errLib.CommonError) {
	squareAccessToken := config.Envs.SquareAccessToken

	if squareAccessToken == "" {
		return nil, errLib.New("square access token required", http.StatusInternalServerError)
	}

	return client.NewClient(
		option.WithBaseURL(
			square.Environments.Sandbox,
		),
		option.WithToken(
			squareAccessToken,
		),
	), nil
}

func GetPaymentLink(squareClient *client.Client, ctx context.Context, userID uuid.UUID, itemName string, quantity int, price decimal.Decimal) (string, *errLib.CommonError) {

	userIDStr := userID.String()

	response, err := squareClient.Checkout.PaymentLinks.Create(
		ctx,
		&checkout.CreatePaymentLinkRequest{
			Order: &square.Order{
				LineItems: []*square.OrderLineItem{
					{
						Quantity: strconv.Itoa(quantity),
						Name: square.String(
							itemName,
						),
						BasePriceMoney: &square.Money{
							Amount: square.Int64(
								price.Mul(decimal.NewFromInt(100)).IntPart(),
							),
							Currency: square.CurrencyCad.Ptr(),
						},
					},
				},
				LocationID: "LZD1EZMQ138SK",
				Metadata: map[string]*string{
					"user_id": &userIDStr, // Add userID to the order metadata
				},
			},
		},
	)

	if err != nil {
		return "", errLib.New(err.Error(), http.StatusInternalServerError)
	}

	urlPtr := response.PaymentLink.URL

	if urlPtr == nil {
		return "", errLib.New("Failed creating payment link", http.StatusInternalServerError)
	}

	return *urlPtr, nil
}
