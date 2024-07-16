package notification

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
)

type ErrorType string

const (
	Technical ErrorType = "technicalError"
)

var (
	errTypeToMsgTemplate = map[ErrorType]string{
		Technical: "A technical error occurred. Please contact your administrator.",
	}
)

type NotificationBuilder struct {
	tenantID      uint64
	stepID        string
	notifType     string
	prefixMessage string
	message       string
}

func NewNotificationBuilder(stepID string, tenantID uint64) *NotificationBuilder {
	return &NotificationBuilder{
		stepID:   stepID,
		tenantID: tenantID,
	}
}

func (nb *NotificationBuilder) WithMessage(message string) *NotificationBuilder {
	nb.message = message
	return nb
}

func (nb *NotificationBuilder) WithPrefixMessage(prefixMessage string) *NotificationBuilder {
	nb.prefixMessage = prefixMessage
	return nb
}

func (nb *NotificationBuilder) WithNotifType(notifType string) *NotificationBuilder {
	nb.notifType = notifType
	return nb
}

func (nb *NotificationBuilder) WithTechnicalError(msg string) *NotificationBuilder {
	if msg != "" {
		nb.message = fmt.Sprintf("A technical error occurred (%s). Please contact your administrator.", msg)
	} else {
		nb.message = errTypeToMsgTemplate[Technical]
	}
	nb.notifType = string(Technical)
	return nb
}

func (nb *NotificationBuilder) Build() *model.Notification {
	message := nb.message
	if nb.prefixMessage != "" {
		message = fmt.Sprintf("%s : %s", nb.prefixMessage, nb.message)
	}
	return &model.Notification{
		ID:       fmt.Sprintf("%s-%s", nb.stepID, nb.notifType),
		TenantID: nb.tenantID,
		Message:  message,
	}
}

// Cause is the same as errors.Cause but with a maxDeep param allowing to get mid level Cause of an error
func Cause(err error, maxDeep int) error {
	type causer interface {
		Cause() error
	}

	for i := 0; i < maxDeep; i++ {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
