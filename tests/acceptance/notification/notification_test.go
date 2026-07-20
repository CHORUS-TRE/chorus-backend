//go:build acceptance

package notification_test

import (
	"context"
	"fmt"
	"net/http"

	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	notification "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/notification/client/notification_service"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/notification/models"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	auth := httptransport.BearerToken(t)
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = auth
	}
}

func validUserAuth() func(*runtime.ClientOperation) {
	return getAuthAsClientOpts(helpers.CreateJWTToken(88888, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "88888"}))
}

var _ = Describe("notification service", func() {

	AfterEach(func() {
		cleanTables()
	})

	Describe("count unread notifications", func() {

		Given("an invalid jwt-token", func() {

			When("the route GET '/api/rest/v1/notifications/count' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := notification.NewNotificationServiceCountUnreadNotificationsParams()

					c := helpers.NotificationServiceHTTPClient()
					_, err := c.NotificationService.NotificationServiceCountUnreadNotifications(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route GET '/api/rest/v1/notifications/count' is called", func() {

				Then("notifications should be returned", func() {
					setupTables()
					req := notification.NewNotificationServiceCountUnreadNotificationsParams()

					c := helpers.NotificationServiceHTTPClient()
					resp, err := c.NotificationService.NotificationServiceCountUnreadNotifications(req, validUserAuth())

					ExpectAPIErr(err).Should(BeNil())
					Expect(resp.Payload.Result).Should(Equal(int64(2)))
				})
			})
		})
	})

	Describe("mark notification as read", func() {

		Given("an invalid jwt-token", func() {

			When("the route POST '/api/rest/v1/notifications/read' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := notification.NewNotificationServiceMarkNotificationsAsReadParams()

					c := helpers.NotificationServiceHTTPClient()
					_, err := c.NotificationService.NotificationServiceMarkNotificationsAsRead(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route POST '/api/rest/v1/notifications/read' is called", func() {

				Then("notifications should be marked as read", func() {
					setupTables()
					req := notification.NewNotificationServiceMarkNotificationsAsReadParams().WithBody(
						&models.ChorusMarkNotificationsAsReadRequest{
							NotificationIds: []string{"88888-notEnoughFunds"},
						})

					c := helpers.NotificationServiceHTTPClient()
					_, err := c.NotificationService.NotificationServiceMarkNotificationsAsRead(req, validUserAuth())

					ExpectAPIErr(err).Should(BeNil())
				})
			})

			When("then route POST '/api/rest/v1/notifications/read' is called with mark all checked", func() {

				Then("notifications should be marked as read", func() {
					setupTables()
					req := notification.NewNotificationServiceMarkNotificationsAsReadParams().WithBody(
						&models.ChorusMarkNotificationsAsReadRequest{
							MarkAll: true,
						})

					c := helpers.NotificationServiceHTTPClient()
					_, err := c.NotificationService.NotificationServiceMarkNotificationsAsRead(req, validUserAuth())

					ExpectAPIErr(err).Should(BeNil())
				})
			})
		})
	})

	Describe("get notifications", func() {

		Given("an invalid jwt-token", func() {

			When("the route GET '/api/rest/v1/notifications/{id}' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := notification.NewNotificationServiceGetNotificationsParams()

					c := helpers.NotificationServiceHTTPClient()
					_, err := c.NotificationService.NotificationServiceGetNotifications(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("a valid request", func() {

				When("the route GET '/api/rest/v1/notifications' is called with default params", func() {

					Then("all notifications should be returned in createdat sort order desc", func() {
						setupTables()
						offset, limit, order, sortType := int64(0), int64(10), "desc", "CREATEDAT"
						req := notification.NewNotificationServiceGetNotificationsParams().
							WithIsRead(nil).WithPaginationOffset(&offset).
							WithPaginationLimit(&limit).WithPaginationSortOrder(&order).WithPaginationSortType(&sortType)

						c := helpers.NotificationServiceHTTPClient()
						res, err := c.NotificationService.NotificationServiceGetNotifications(req, validUserAuth())

						ExpectAPIErr(err).Should(BeNil())
						Expect(len(res.Payload.Result)).Should(Equal(4))
						Expect(res.Payload.TotalItems).Should(Equal(int64(4)))
						Expect(res.Payload.Result[0].ID).Should(Equal("88888-notEnoughFunds"))
						Expect(res.Payload.Result[1].ID).Should(Equal("88889-notEnoughFunds"))
					})
				})

				When("the route GET '/api/rest/v1/notifications' is called with default params and limit 2", func() {

					Then("all notifications should be returned in createdat sort order desc", func() {
						setupTables()
						query, offset, limit, order, sortType := []string{""}, int64(0), int64(2), "desc", "CREATEDAT"
						req := notification.NewNotificationServiceGetNotificationsParams().
							WithPaginationQuery(query).WithIsRead(nil).WithPaginationOffset(&offset).
							WithPaginationLimit(&limit).WithPaginationSortOrder(&order).WithPaginationSortType(&sortType)

						c := helpers.NotificationServiceHTTPClient()
						res, err := c.NotificationService.NotificationServiceGetNotifications(req, validUserAuth())

						ExpectAPIErr(err).Should(BeNil())
						Expect(len(res.Payload.Result)).Should(Equal(2))
						Expect(res.Payload.TotalItems).Should(Equal(int64(4)))
						Expect(res.Payload.Result[0].ID).Should(Equal("88888-notEnoughFunds"))
						Expect(res.Payload.Result[1].ID).Should(Equal("88889-notEnoughFunds"))
					})
				})

				// TODO fix the isRead filter, then reintroduce the spec:
				// GET '/api/rest/v1/notifications' with isRead=false should
				// only return unread notifications (88888-notEnoughFunds,
				// 88889-err2 in the fixtures).

				// TODO fix the sort type, then reintroduce the spec:
				// GET '/api/rest/v1/notifications' with query 'notEnoughFunds'
				// should only return notifications whose id contains it
				// (88888-notEnoughFunds, 88889-notEnoughFunds in the fixtures).
			})
		})
	})
})

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants(id, name) VALUES (88888, 'tenant test');
	INSERT INTO users (id, tenantid, username) VALUES (88888, 88888, 'manager01'), (88889, 88888, 'manager02');
	INSERT INTO notifications (id, tenantid, userid, message, createdat) VALUES
		('88888-notEnoughFunds', 88888, 88888, 'Fail: err1', now()),
		('88889-notEnoughFunds', 88888, 88888, 'Fail: err1', '2020-03-08T15:51:50'),
		('88889-err2', 88888, 88888, 'Fail: err2', '2020-03-07T15:51:50'),
		('88890-err3', 88888, 88888, 'Fail: The Transaction for cmta deploy has been rejected', '2020-03-06T15:51:50');

	INSERT INTO notifications_read_by (tenantid, notificationid, userid, readat) VALUES (88888, '88888-notEnoughFunds', 88888, null),
		(88888, '88889-notEnoughFunds', 88888, now()), (88888, '88889-err2', 88888, null), (88888, '88890-err3', 88888, now());

	INSERT INTO notifications_read_by (tenantid, notificationid, userid, readat) VALUES (88888, '88888-notEnoughFunds', 88889, now()),
		(88888, '88889-notEnoughFunds', 88889, now()), (88888, '88889-err2', 88889, now()), (88888, '88890-err3', 88889, now());
	`
	helpers.Populate(q)
}

func cleanTables() {
	q := `
DELETE FROM notifications_read_by WHERE tenantid = 88888;
DELETE FROM notifications WHERE tenantid = 88888;
DELETE FROM users WHERE tenantid = 88888;
DELETE FROM tenants WHERE id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
