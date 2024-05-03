package socket

import (
	"time"

	"github.com/ectrc/snow/aid"
	"github.com/ectrc/snow/person"
)

func EmitGiftReceived(person *person.Person) {
	s, ok := JabberSockets.Get(person.ID)
	if !ok {
		return
	}

	s.JabberSendMessageToPerson(aid.JSON{
		"payload": aid.JSON{
			"gifts": []aid.JSON{{
				"Wahgsdhjgasjkd": "Wahgsdhjgasjkd",
			}},
		},
		"type": "com.epicgames.gift.received",
		"timestamp": time.Now().Format("2006-01-02T15:04:05.999Z"),
	})
}

func EmitPartyMemberJoined(party *person.Party, joiningMember *person.PartyMember) {
	for _, partyMember := range party.Members {
		s, ok := JabberSockets.Get(partyMember.Person.ID)
		if !ok {
			continue
		}

		joiningMember.IncreaseRevision()
		s.JabberSendMessageToPerson(aid.JSON{
			"account_id": joiningMember.Person.ID,
			"account_dn": joiningMember.Person.DisplayName,
			"connection": joiningMember.Connection,
			"member_state_updated": joiningMember.Meta,
			"updated_at": joiningMember.UpdatedAt.Format(time.RFC3339),
			"joined_at": joiningMember.JoinedAt.Format(time.RFC3339),
			"ns": "Fortnite",
			"party_id": party.ID,
			"sent": time.Now().Format(time.RFC3339),
			"revision": joiningMember.Revision,
			"type": "com.epicgames.social.party.notification.v0.MEMBER_JOINED",
		})

		s.JabberSendMessageToPerson(aid.JSON{
			"interactions": []aid.JSON{{
				"app": "Snow",
				"namespace": "Fortnite",
				"fromAccountId": joiningMember.Person.ID,
				"toAccountId": partyMember.Person.ID,
				"interactionScoreIncremental": aid.JSON{
					"count": 1,
					"total": 1,
				},
				"interactionType": "PartyJoined",
				"isFriend": true,
				"happenedAt": time.Now().Unix(),
			}},
			"type": "com.epicgames.social.interactions.notification.v2",
		})
	}
}

func EmitPartyMemberLeft(party *person.Party, leavingMember *person.PartyMember) {
	for _, m := range party.Members {
		s, ok := JabberSockets.Get(m.Person.ID)
		if !ok {
			continue
		}

		leavingMember.IncreaseRevision()
		s.JabberSendMessageToPerson(aid.JSON{
			"account_id": leavingMember.Person.ID,
			"member_state_updated": aid.JSON{},
			"ns": "Fortnite",
			"party_id": party.ID,
			"sent": time.Now().Format(time.RFC3339),
			"revision": leavingMember.Revision,
			"type": "com.epicgames.social.party.notification.v0.MEMBER_LEFT",
		})
	}
}

func EmitPartyMemberMetaUpdated(party *person.Party, member *person.PartyMember, update map[string]interface{}, deleted []string) {
	for _, m := range party.Members {
		s, ok := JabberSockets.Get(m.Person.ID)
		if !ok {
			continue
		}

		member.IncreaseRevision()
		s.JabberSendMessageToPerson(aid.JSON{
			"account_id": member.Person.ID,
			"account_dn": member.Person.DisplayName,
			"member_state_updated": update,
			"member_state_removed": deleted,
			"member_state_overriden": aid.JSON{},
			"updated_at": member.UpdatedAt.Format(time.RFC3339),
			"joined_at": member.JoinedAt.Format(time.RFC3339),
			"ns": "Fortnite",
			"party_id": party.ID,
			"sent": time.Now().Format(time.RFC3339),
			"revision": member.Revision,
			"type": "com.epicgames.social.party.notification.v0.MEMBER_STATE_UPDATED",
		})
	}
}

func EmitPartyMetaUpdated(party *person.Party, override map[string]interface{}, deleted []string, update map[string]interface{}) {
	for _, m := range party.Members {
		s, ok := JabberSockets.Get(m.Person.ID)
		if !ok {
			continue
		}

		party.IncreaseRevision()
		s.JabberSendMessageToPerson(aid.JSON{
			"captain_id": party.Captain.Person.ID,
			"party_state_updated": update,
			"party_state_removed": deleted,
			"party_state_overriden": override,
			"party_privacy_type": party.Config["joinability"],
			"party_type": party.Config["type"],
			"party_sub_type": party.Config["sub_type"],
			"max_number_of_members": party.Config["max_size"],
			"invite_ttl_seconds": party.Config["invite_ttl"],
			"intention_ttl_seconds": party.Config["intention_ttl"],
			"updated_at": time.Now().Format(time.RFC3339),
			"created_at": party.CreatedAt.Format(time.RFC3339),
			"ns": "Fortnite",
			"party_id": party.ID,
			"sent": time.Now().Format(time.RFC3339),
			"revision": party.Revision,
			"type": "com.epicgames.social.party.notification.v0.PARTY_UPDATED",
		})
	}
}

func EmitPartyNewCaptain(party *person.Party) {
	for _, m := range party.Members {
		s, ok := JabberSockets.Get(m.Person.ID)
		if !ok {
			continue
		}

		party.IncreaseRevision()
		s.JabberSendMessageToPerson(aid.JSON{
			"account_id": party.Captain.Person.ID,
			"account_dn": party.Captain.Person.DisplayName,
			"ns": "Fortnite",
			"party_id": party.ID,
			"sent": time.Now().Format(time.RFC3339),
			"revision": party.Revision,
			"type": "com.epicgames.social.party.notification.v0.MEMBER_NEW_CAPTAIN",
		})
	}
}

func EmitPartyInvite(invite *person.PartyInvite) {
	s, ok := JabberSockets.Get(invite.Towards.ID)
	if !ok {
		return
	}

	s.JabberSendMessageToPerson(aid.JSON{
		"inviter_id": invite.Inviter.ID,
		"inviter_dn": invite.Inviter.DisplayName,
		"invitee_id": invite.Towards.ID,
		"meta": invite.Meta,
		"sent_at": invite.CreatedAt.Format(time.RFC3339),
		"updated_at": invite.UpdatedAt.Format(time.RFC3339),
		"friends_ids": []string{},
		"members_count": len(invite.Party.Members),
		"party_id": invite.Party.ID,
		"ns": "Fortnite",
		"sent": time.Now().Format(time.RFC3339),
		"type": "com.epicgames.social.party.notification.v0.INITIAL_INVITE",
	})
}

func EmitPartyIntention(invite *person.PartyIntention) {
	s, ok := JabberSockets.Get(invite.Towards.ID)
	if !ok {
		return
	}

	s.JabberSendMessageToPerson(aid.JSON{
		"requester_id": invite.Requester.ID,
		"requester_dn": invite.Requester.DisplayName,
		"requester_pl": "win",
		"requester_pl_dn": invite.Requester.DisplayName,
		"requestee_id": invite.Towards.ID,
		"meta": invite.Meta,
		"sent_at": invite.CreatedAt.Format(time.RFC3339),
		"friends_ids": []string{},
		"members_count": len(invite.Party.Members),
		"party_id": invite.Party.ID,
		"ns": "Fortnite",
		"sent": time.Now().Format(time.RFC3339),
		"type": "com.epicgames.social.party.notification.v0.INITIAL_INTENTION",
	})
}

func EmitPartyPingFromInvite(i *person.PartyInvite) {
	s, ok := JabberSockets.Get(i.Towards.ID)
	if !ok {
		return
	}

	meta := i.Meta
	meta["urn:epic:member:dn_s"] = i.Inviter.DisplayName
	meta["urn:epic:invite:platformdata_s"] = "RequestToJoin"
	meta["urn:epic:conn:platform_s"] = "WIN"
	meta["urn:epic:conn:platform:dn_s"] = i.Inviter.DisplayName

	s.JabberSendMessageToPerson(aid.JSON{
		"expires_at": i.CreatedAt.Add(time.Minute * 60).Format(time.RFC3339),
		"pinger_id": i.Inviter.ID,
		"pinger_dn": i.Inviter.DisplayName,
		"pinger_pl": "win",
		"pinger_pl_dn": i.Inviter.DisplayName,
		"meta": meta,
		"ns": "Fortnite",
		"sent": time.Now().Format(time.RFC3339),
		"type": "com.epicgames.social.party.notification.v0.PING",
	})
}