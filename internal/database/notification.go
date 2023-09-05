// Copyright (c) 2021 - 2023, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package database

import (
	"context"
	"database/sql"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/internal/logger"
	"github.com/autobrr/autobrr/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
)

type NotificationRepo struct {
	log zerolog.Logger
	db  *DB
}

func NewNotificationRepo(log logger.Logger, db *DB) domain.NotificationRepo {
	return &NotificationRepo{
		log: log.With().Str("repo", "notification").Logger(),
		db:  db,
	}
}

func (r *NotificationRepo) Find(ctx context.Context, params domain.NotificationQueryParams) ([]domain.Notification, int, error) {

	queryBuilder := r.db.squirrel.
		Select("id", "name", "type", "enabled", "events", "webhook", "token", "api_key", "channel", "priority", "topic", "host", "username", "password", "require_encryption", "smtp_port", "from_address", "recipient_addresses", "created_at", "updated_at", "COUNT(*) OVER() AS total_count").
		From("notification").
		OrderBy("name")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, 0, errors.Wrap(err, "error building query")
	}

	rows, err := r.db.handler.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error executing query")
	}

	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	totalCount := 0
	for rows.Next() {
		var n domain.Notification

		var webhook, token, apiKey, channel, topic, host, username, password, fromAddress sql.NullString

		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Enabled, pq.Array(&n.Events), &webhook, &token, &apiKey, &channel, &n.Priority, &topic, &host, &username, &password, &n.RequireEncryption, &n.SmtpPort, &fromAddress, pq.Array(&n.RecipientAddresses), &n.CreatedAt, &n.UpdatedAt, &totalCount); err != nil {
			return nil, 0, errors.Wrap(err, "error scanning row")
		}

		n.APIKey = apiKey.String
		n.Webhook = webhook.String
		n.Token = token.String
		n.Channel = channel.String
		n.Topic = topic.String
		n.Host = host.String
		n.Username = username.String
		n.Password = password.String
		n.FromAddress = fromAddress.String

		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "error rows find")
	}

	return notifications, totalCount, nil
}

func (r *NotificationRepo) List(ctx context.Context) ([]domain.Notification, error) {

	rows, err := r.db.handler.QueryContext(ctx, "SELECT id, name, type, enabled, events, token, api_key,  webhook, title, icon, host, username, password, channel, targets, devices, priority, topic, require_encryption, smtp_port, from_address, recipient_addresses, created_at, updated_at FROM notification ORDER BY name ASC")
	if err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		//var eventsSlice []string

		var token, apiKey, webhook, title, icon, host, username, password, channel, targets, devices, topic, fromAddress sql.NullString
		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Enabled, pq.Array(&n.Events), &token, &apiKey, &webhook, &title, &icon, &host, &username, &password, &channel, &targets, &devices, &n.Priority, &topic, &n.RequireEncryption, &n.SmtpPort, &fromAddress, pq.Array(&n.RecipientAddresses), &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, errors.Wrap(err, "error scanning row")
		}

		//n.Events = ([]domain.NotificationEvent)(eventsSlice)
		n.Token = token.String
		n.APIKey = apiKey.String
		n.Webhook = webhook.String
		n.Title = title.String
		n.Icon = icon.String
		n.Host = host.String
		n.Username = username.String
		n.Password = password.String
		n.Channel = channel.String
		n.Targets = targets.String
		n.Devices = devices.String
		n.Topic = topic.String
		n.FromAddress = fromAddress.String

		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error rows list")
	}

	return notifications, nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, id int) (*domain.Notification, error) {

	queryBuilder := r.db.squirrel.
		Select(
			"id",
			"name",
			"type",
			"enabled",
			"events",
			"token",
			"api_key",
			"webhook",
			"title",
			"icon",
			"host",
			"username",
			"password",
			"channel",
			"targets",
			"devices",
			"priority",
			"topic",
			"require_encryption",
			"smtp_port",
			"from_address",
			"recipient_addresses",
			"created_at",
			"updated_at",
		).
		From("notification").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	row := r.db.handler.QueryRowContext(ctx, query, args...)
	if err := row.Err(); err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	var n domain.Notification

	var token, apiKey, webhook, title, icon, host, username, password, channel, targets, devices, topic, fromAddress sql.NullString
	if err := row.Scan(&n.ID, &n.Name, &n.Type, &n.Enabled, pq.Array(&n.Events), &token, &apiKey, &webhook, &title, &icon, &host, &username, &password, &channel, &targets, &devices, &n.Priority, &topic, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, errors.Wrap(err, "error scanning row")
	}

	n.Token = token.String
	n.APIKey = apiKey.String
	n.Webhook = webhook.String
	n.Title = title.String
	n.Icon = icon.String
	n.Host = host.String
	n.Username = username.String
	n.Password = password.String
	n.Channel = channel.String
	n.Targets = targets.String
	n.Devices = devices.String
	n.Topic = topic.String
	n.FromAddress = fromAddress.String

	return &n, nil
}

func (r *NotificationRepo) Store(ctx context.Context, notification domain.Notification) (*domain.Notification, error) {
	webhook := toNullString(notification.Webhook)
	token := toNullString(notification.Token)
	apiKey := toNullString(notification.APIKey)
	channel := toNullString(notification.Channel)
	topic := toNullString(notification.Topic)
	host := toNullString(notification.Host)
	username := toNullString(notification.Username)
	password := toNullString(notification.Password)
	fromAddress := toNullString(notification.FromAddress)

	queryBuilder := r.db.squirrel.
		Insert("notification").
		Columns(
			"name",
			"type",
			"enabled",
			"events",
			"webhook",
			"token",
			"api_key",
			"channel",
			"priority",
			"topic",
			"host",
			"username",
			"password",
			"require_encryption",
			"smtp_port",
			"from_address",
			"recipient_addresses",
		).
		Values(
			notification.Name,
			notification.Type,
			notification.Enabled,
			pq.Array(notification.Events),
			webhook,
			token,
			apiKey,
			channel,
			notification.Priority,
			topic,
			host,
			username,
			password,
			notification.RequireEncryption,
			notification.SmtpPort,
			fromAddress,
			pq.Array(notification.RecipientAddresses),
		).
		Suffix("RETURNING id").RunWith(r.db.handler)

	// return values
	var retID int64

	if err := queryBuilder.QueryRowContext(ctx).Scan(&retID); err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	r.log.Debug().Msgf("notification.store: added new %v", retID)
	notification.ID = int(retID)

	return &notification, nil
}

func (r *NotificationRepo) Update(ctx context.Context, notification domain.Notification) (*domain.Notification, error) {
	webhook := toNullString(notification.Webhook)
	token := toNullString(notification.Token)
	apiKey := toNullString(notification.APIKey)
	channel := toNullString(notification.Channel)
	topic := toNullString(notification.Topic)
	host := toNullString(notification.Host)
	username := toNullString(notification.Username)
	password := toNullString(notification.Password)
	fromAddress := toNullString(notification.FromAddress)

	queryBuilder := r.db.squirrel.
		Update("notification").
		Set("name", notification.Name).
		Set("type", notification.Type).
		Set("enabled", notification.Enabled).
		Set("events", pq.Array(notification.Events)).
		Set("webhook", webhook).
		Set("token", token).
		Set("api_key", apiKey).
		Set("channel", channel).
		Set("priority", notification.Priority).
		Set("topic", topic).
		Set("host", host).
		Set("username", username).
		Set("password", password).
		Set("require_encryption", notification.RequireEncryption).
		Set("smtp_port", notification.SmtpPort).
		Set("from_address", fromAddress).
		Set("recipient_addresses", pq.Array(notification.RecipientAddresses)).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": notification.ID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "error building query")
	}

	if _, err = r.db.handler.ExecContext(ctx, query, args...); err != nil {
		return nil, errors.Wrap(err, "error executing query")
	}

	r.log.Debug().Msgf("notification.update: %v", notification.Name)

	return &notification, nil
}

func (r *NotificationRepo) Delete(ctx context.Context, notificationID int) error {
	queryBuilder := r.db.squirrel.
		Delete("notification").
		Where(sq.Eq{"id": notificationID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return errors.Wrap(err, "error building query")
	}

	if _, err = r.db.handler.ExecContext(ctx, query, args...); err != nil {
		return errors.Wrap(err, "error executing query")
	}

	r.log.Info().Msgf("notification.delete: successfully deleted: %v", notificationID)

	return nil
}
