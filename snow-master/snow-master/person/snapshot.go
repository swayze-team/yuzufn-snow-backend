package person

import "github.com/ectrc/snow/storage"

type PersonSnapshot struct {
	ID                  string
	DisplayName         string
	RefundTickets			  int
	Permissions         int64
	AthenaProfile       ProfileSnapshot
	CommonCoreProfile   ProfileSnapshot
	CommonPublicProfile ProfileSnapshot
	Profile0Profile     ProfileSnapshot
	CollectionsProfile  ProfileSnapshot
	CreativeProfile     ProfileSnapshot
	CurrentSeasonStats	SeasonStats
	AllSeasonsStats			[]SeasonStats
	Receipts            []storage.DB_Receipt
	BanHistory          []storage.DB_BanStatus
	Discord             storage.DB_DiscordPerson
	Relationships				map[string]*Relationship
	Parties						 	map[string]*Party
	Invites							map[string]*PartyInvite
	Intentions					map[string]*PartyIntention
}

type ProfileSnapshot struct {
	ID         string
	Items      map[string]ItemSnapshot
	Gifts      map[string]GiftSnapshot
	Variants   map[string]VariantChannel
	Quests     map[string]Quest
	Attributes map[string]Attribute
	Loadouts   map[string]Loadout
	Revision   int
	Type       string
}

type ItemSnapshot struct {
	ID          string
	TemplateID  string
	Quantity    int
	Favorite    bool
	HasSeen     bool
	Variants    []VariantChannel
	ProfileType string
}

type GiftSnapshot struct {
	ID         string
	TemplateID string
	Quantity   int
	FromID     string
	GiftedAt   int64
	Message    string
	Loot       []Item
}