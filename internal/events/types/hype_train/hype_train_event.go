// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package hype_train

import (
	"encoding/json"
	"time"

	"github.com/twitchdev/twitch-cli/internal/events"
	"github.com/twitchdev/twitch-cli/internal/models"
	"github.com/twitchdev/twitch-cli/internal/util"
)

var transportsSupported = map[string]bool{
	models.TransportWebSub:   true,
	models.TransportEventSub: true,
}
var triggerSupported = []string{"hype-train-begin", "hype-train-progress", "hype-train-end"}
var triggerMapping = map[string]map[string]string{
	models.TransportWebSub: {
		"hype-train-progress": "hypetrain.progression",
		"hype-train-begin":    "hypetrain.progression",
		"hype-train-end":      "hypetrain.progression",
	},
	models.TransportEventSub: {
		"hype-train-progress": "channel.hype_train.progress",
		"hype-train-begin":    "channel.hype_train.begin",
		"hype-train-end":      "channel.hype_train.end",
	},
}

type Event struct{}

func (e Event) GenerateEvent(params events.MockEventParameters) (events.MockEventResponse, error) {
	var event []byte
	var err error
	lastUser := util.RandomUserID()
	lastTotal := util.RandomViewerCount()
	lastType := util.RandomType()
	//Local variables which will be used for the trigger params below
	localTotal := util.RandomViewerCount()
	localGoal := util.RandomViewerCount()
	localProgress := (localTotal / localGoal)

	switch params.Transport {
	case models.TransportEventSub:
		body := *&models.HypeTrainEventSubResponse{
			Subscription: models.EventsubSubscription{
				ID:      params.ID,
				Status:  "enabled",
				Type:    triggerMapping[params.Transport][params.Trigger],
				Version: "1.0",
				Condition: models.EventsubCondition{
					BroadcasterUserID: params.ToUserID,
				},
				Transport: models.EventsubTransport{
					Method:   "webhook",
					Callback: "null",
				},
				Cost:      0,
				CreatedAt: util.GetTimestamp().Format(time.RFC3339Nano),
			},
			Event: models.HypeTrainEventSubEvent{
				BroadcasterUserID:    params.ToUserID,
				BroadcasterUserLogin: params.ToUserName,
				BroadcasterUserName:  params.ToUserName,
				Total:                util.RandomViewerCount(),
				Progress:             localProgress,
				Goal:                 localGoal,
				TopContributions: []models.ContributionData{
					{
						TotalContribution:            util.RandomViewerCount(),
						TypeOfContribution:           util.RandomType(),
						UserWhoMadeContribution:      util.RandomUserID(),
						UserNameWhoMadeContribution:  "cli_user1",
						UserLoginWhoMadeContribution: "cli_user1",
					},
					{
						TotalContribution:            lastTotal,
						TypeOfContribution:           lastType,
						UserWhoMadeContribution:      lastUser,
						UserNameWhoMadeContribution:  "cli_user2",
						UserLoginWhoMadeContribution: "cli_user2",
					},
				},
				LastContribution: models.ContributionData{
					TotalContribution:            lastTotal,
					TypeOfContribution:           lastType,
					UserWhoMadeContribution:      lastUser,
					UserNameWhoMadeContribution:  "cli_user2",
					UserLoginWhoMadeContribution: "cli_user2",
				},
				StartedAtTimestamp: util.GetTimestamp().Format(time.RFC3339Nano),
				ExpiresAtTimestamp: util.GetTimestamp().Format(time.RFC3339Nano),
			},
		}
		if triggerMapping[params.Transport][params.Trigger] == "hype-train-end " {
			body.Event.CooldownEndsAtTimestamp = util.GetTimestamp().Format(time.RFC3339Nano)
		}
		event, err = json.Marshal(body)
		if err != nil {
			return events.MockEventResponse{}, err
		}
	case models.TransportWebSub:
		body := *&models.HypeTrainWebSubResponse{
			Data: []models.HypeTrainWebSubEvent{
				{
					ID:             params.ID,
					EventType:      triggerMapping[params.Transport][params.Trigger],
					EventTimestamp: util.GetTimestamp().Format(time.RFC3339),
					Version:        "1.0",
					EventData: models.HypeTrainWebsubEventData{
						BroadcasterID:        params.ToUserID,
						CooldownEndTimestamp: util.GetTimestamp().Format(time.RFC3339),
						ExpiresAtTimestamp:   util.GetTimestamp().Format(time.RFC3339),
						Goal:                 localGoal,
						Id:                   util.RandomGUID(),
						LastContribution: models.ContributionData{
							TotalContribution:  lastTotal,
							TypeOfContribution: lastType,
							WebSubUser:         lastUser,
						},
						Level:              util.RandomViewerCount() % 4,
						StartedAtTimestamp: util.GetTimestamp().Format(time.RFC3339),
						TopContributions: []models.ContributionData{
							{
								TotalContribution:  lastTotal,
								TypeOfContribution: lastType,
								WebSubUser:         lastUser,
							},
							{
								TotalContribution:  util.RandomViewerCount(),
								TypeOfContribution: util.RandomType(),
								WebSubUser:         util.RandomUserID(),
							},
						},
						Total: localTotal,
					},
				},
			},
		}
		event, err = json.Marshal(body)
		if err != nil {
			return events.MockEventResponse{}, err
		}
	default:
		return events.MockEventResponse{}, nil
	}
	return events.MockEventResponse{
		ID:       params.ID,
		JSON:     event,
		FromUser: params.FromUserID,
		ToUser:   params.ToUserID,
	}, nil
}
func (e Event) ValidTransport(t string) bool {
	return transportsSupported[t]
}
func (e Event) ValidTrigger(t string) bool {
	for _, ts := range triggerSupported {
		if ts == t {
			return true
		}
	}
	return false
}
func (e Event) GetTopic(transport string, trigger string) string {
	return triggerMapping[transport][trigger]
}
