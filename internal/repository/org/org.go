package org

import (
	"context"
	"encoding/json"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	"github.com/zitadel/zitadel/internal/repository/project"
	"github.com/zitadel/zitadel/internal/repository/user"
)

const (
	uniqueOrgname           = "org_name"
	OrgAddedEventType       = orgEventTypePrefix + "added"
	OrgChangedEventType     = orgEventTypePrefix + "changed"
	OrgDeactivatedEventType = orgEventTypePrefix + "deactivated"
	OrgReactivatedEventType = orgEventTypePrefix + "reactivated"
	OrgRemovedEventType     = orgEventTypePrefix + "removed"
)

func NewAddOrgNameUniqueConstraint(orgName string) *eventstore.EventUniqueConstraint {
	return eventstore.NewAddEventUniqueConstraint(
		uniqueOrgname,
		orgName,
		"Errors.Org.AlreadyExists")
}

func NewRemoveOrgNameUniqueConstraint(orgName string) *eventstore.EventUniqueConstraint {
	return eventstore.NewRemoveEventUniqueConstraint(
		uniqueOrgname,
		orgName)
}

type OrgAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Name string `json:"name,omitempty"`
}

func (e *OrgAddedEvent) Payload() interface{} {
	return e
}

func (e *OrgAddedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return []*eventstore.EventUniqueConstraint{NewAddOrgNameUniqueConstraint(e.Name)}
}

func NewOrgAddedEvent(ctx context.Context, aggregate *eventstore.Aggregate, name string) *OrgAddedEvent {
	return &OrgAddedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			OrgAddedEventType,
		),
		Name: name,
	}
}

func OrgAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	orgAdded := &OrgAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, orgAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "ORG-Bren2", "unable to unmarshal org added")
	}

	return orgAdded, nil
}

type OrgChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Name    string `json:"name,omitempty"`
	oldName string `json:"-"`
}

func (e *OrgChangedEvent) Payload() interface{} {
	return e
}

func (e *OrgChangedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return []*eventstore.EventUniqueConstraint{
		NewRemoveOrgNameUniqueConstraint(e.oldName),
		NewAddOrgNameUniqueConstraint(e.Name),
	}
}

func NewOrgChangedEvent(ctx context.Context, aggregate *eventstore.Aggregate, oldName, newName string) *OrgChangedEvent {
	return &OrgChangedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			OrgChangedEventType,
		),
		Name:    newName,
		oldName: oldName,
	}
}

func OrgChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	orgChanged := &OrgChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := json.Unmarshal(event.Data, orgChanged)
	if err != nil {
		return nil, errors.ThrowInternal(err, "ORG-Bren2", "unable to unmarshal org added")
	}

	return orgChanged, nil
}

type OrgDeactivatedEvent struct {
	eventstore.BaseEvent `json:"-"`
}

func (e *OrgDeactivatedEvent) Payload() interface{} {
	return e
}

func (e *OrgDeactivatedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewOrgDeactivatedEvent(ctx context.Context, aggregate *eventstore.Aggregate) *OrgDeactivatedEvent {
	return &OrgDeactivatedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			OrgDeactivatedEventType,
		),
	}
}

func OrgDeactivatedEventMapper(event *repository.Event) (eventstore.Event, error) {
	return &OrgDeactivatedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}, nil
}

type OrgReactivatedEvent struct {
	eventstore.BaseEvent `json:"-"`
}

func (e *OrgReactivatedEvent) Payload() interface{} {
	return e
}

func (e *OrgReactivatedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func NewOrgReactivatedEvent(ctx context.Context, aggregate *eventstore.Aggregate) *OrgReactivatedEvent {
	return &OrgReactivatedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			OrgReactivatedEventType,
		),
	}
}

func OrgReactivatedEventMapper(event *repository.Event) (eventstore.Event, error) {
	return &OrgReactivatedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}, nil
}

type OrgRemovedEvent struct {
	eventstore.BaseEvent `json:"-"`
	name                 string
	usernames            []string
	loginMustBeDomain    bool
	domains              []string
	externalIDPs         []*domain.UserIDPLink
	samlEntityIDs        []string
}

func (e *OrgRemovedEvent) Payload() interface{} {
	return nil
}

func (e *OrgRemovedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	constraints := []*eventstore.EventUniqueConstraint{
		NewRemoveOrgNameUniqueConstraint(e.name),
	}
	for _, name := range e.usernames {
		constraints = append(constraints, user.NewRemoveUsernameUniqueConstraint(name, e.Aggregate().ID, e.loginMustBeDomain))
	}
	for _, domain := range e.domains {
		constraints = append(constraints, NewRemoveOrgDomainUniqueConstraint(domain))
	}
	for _, idp := range e.externalIDPs {
		constraints = append(constraints, user.NewRemoveUserIDPLinkUniqueConstraint(idp.IDPConfigID, idp.ExternalUserID))
	}
	for _, entityID := range e.samlEntityIDs {
		constraints = append(constraints, project.NewRemoveSAMLConfigEntityIDUniqueConstraint(entityID))
	}
	return constraints
}

func NewOrgRemovedEvent(ctx context.Context, aggregate *eventstore.Aggregate, name string, usernames []string, loginMustBeDomain bool, domains []string, externalIDPs []*domain.UserIDPLink, samlEntityIDs []string) *OrgRemovedEvent {
	return &OrgRemovedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			OrgRemovedEventType,
		),
		name:              name,
		usernames:         usernames,
		domains:           domains,
		externalIDPs:      externalIDPs,
		samlEntityIDs:     samlEntityIDs,
		loginMustBeDomain: loginMustBeDomain,
	}
}

func OrgRemovedEventMapper(event *repository.Event) (eventstore.Event, error) {
	return &OrgRemovedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}, nil
}
