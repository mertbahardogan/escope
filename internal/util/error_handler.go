package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
)

func HandleServiceError(err error, operation string) {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("%s failed: %s\n", operation, constants.MsgTimeoutGeneric)
		} else {
			fmt.Printf("%s failed: %v\n", operation, err)
		}
	}
}

func HandleServiceErrorWithReturn(err error, operation string) bool {
	if err != nil {
		HandleServiceError(err, operation)
		return true
	}
	return false
}
