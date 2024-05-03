package person

import (
	"strings"

	"github.com/ectrc/snow/aid"
	"github.com/ectrc/snow/storage"
	"github.com/google/uuid"
)

type Item struct {
	ID string
	ProfileID string
	TemplateID string
	Quantity int
	Favorite bool
	HasSeen bool
	Variants []*VariantChannel
	ProfileType string
}

func NewItem(templateID string, quantity int) *Item {
	return &Item{
		ID: uuid.New().String(),
		TemplateID: templateID,
		Quantity: quantity,
		Favorite: false,
		HasSeen: false,
		Variants: []*VariantChannel{},
	}
}

func NewItemWithType(templateID string, quantity int, profile string) *Item {
	return &Item{
		ID: uuid.New().String(),
		TemplateID: templateID,
		Quantity: quantity,
		Favorite: false,
		HasSeen: false,
		Variants: []*VariantChannel{},
		ProfileType: profile,
	}
}

func FromDatabaseItem(item *storage.DB_Item) *Item {
	variants := []*VariantChannel{}

	for _, variant := range item.Variants {
		
		variants = append(variants, FromDatabaseVariant(&variant))
	}

	return &Item{
		ID: item.ID,
		TemplateID: item.TemplateID,
		Quantity: item.Quantity,
		Favorite: item.Favorite,
		HasSeen: item.HasSeen,
		Variants: variants,
	}
}

func FromDatabaseGiftLoot(item *storage.DB_GiftLoot) *Item {
	return &Item{
		ID:	item.ID,
		TemplateID: item.TemplateID,
		Quantity: item.Quantity,
		Favorite: false,
		HasSeen: false,
		Variants: []*VariantChannel{},
		ProfileType: item.ProfileType,
	}
}

func FromDatabasePurchaseLoot(item *storage.DB_PurchaseLoot) *Item {
	return &Item{
		ID:	item.ID,
		TemplateID: item.TemplateID,
		Quantity: item.Quantity,
		Favorite: false,
		HasSeen: false,
		Variants: []*VariantChannel{},
		ProfileType: item.ProfileType,
	}
}

func FromDatabaseReceiptLoot(item *storage.DB_ReceiptLoot) *Item {
	return &Item{
		ID:	item.ID,
		TemplateID: item.TemplateID,
		Quantity: item.Quantity,
		Favorite: false,
		HasSeen: false,
		Variants: []*VariantChannel{},
		ProfileType: item.ProfileType,
	}
}

func (i *Item) GenerateFortniteItemEntry() aid.JSON {
	attributes := aid.JSON{}

	switch strings.Split(i.TemplateID, ":")[0] {
	case "Currency": 
		attributes["platform"] = "Shared"
	default:
		attributes = aid.JSON{
			"variants": i.GenerateFortniteItemVariantChannels(),
			"favorite": i.Favorite,
			"item_seen": i.HasSeen,
		}
	}

	return aid.JSON{
		"templateId": i.TemplateID,
		"attributes": attributes,
		"quantity": i.Quantity,
	}
}

func (i *Item) GenerateFortniteItemVariantChannels() []aid.JSON {
	variants := []aid.JSON{}

	for _, variant := range i.Variants {
		variants = append(variants, aid.JSON{
			"channel": variant.Channel,
			"owned": variant.Owned,
			"active": variant.Active,
		})
	}

	return variants
}

func (i *Item) GetAttribute(attribute string) interface{} {
	switch attribute {
	case "Favorite":
		return i.Favorite
	case "HasSeen":
		return i.HasSeen
	case "Variants":
		return i.Variants
	}

	return nil
}

func (i *Item) Delete() {
	storage.Repo.DeleteItem(i.ID)	
}

func (i *Item) DeleteLoot() {
	storage.Repo.DeleteLoot(i.ID)
}

func (i *Item) DeleteReceiptLoot() {
	storage.Repo.DeleteReceiptLoot(i.ID)
}

func (i *Item) NewChannel(channel string, owned []string, active string) *VariantChannel {
	return &VariantChannel{
		ID: uuid.New().String(),
		ItemID: i.ID,
		Channel: channel,
		Owned: owned,
		Active: active,
	}
}

func (i *Item) AddChannel(channel *VariantChannel) {
	i.Variants = append(i.Variants, channel)
}

func (i *Item) RemoveChannel(channel *VariantChannel) {
	var vId string
	for index, c := range i.Variants {
		if c.Channel == channel.Channel {
			vId = c.ID
			i.Variants = append(i.Variants[:index], i.Variants[index+1:]...)
		}
	}
	storage.Repo.DeleteVariant(vId)
}

func (i *Item) GetChannel(channel string) *VariantChannel {
	for _, c := range i.Variants {
		if c.Channel == channel {
			return c
		}
	}

	return nil
}

func (i *Item) FillChannels(channels []*VariantChannel) {
	i.Variants = []*VariantChannel{}
	for _, channel := range channels {
		i.AddChannel(channel)
	}
}

func (i *Item) EquipChannel(channel string, variant string) {
	for _, c := range i.Variants {
		if c.Channel == channel {
			c.Active = variant
		}
	}
}

func (i *Item) ToDatabase(profileId string) *storage.DB_Item {
	variants := []storage.DB_VariantChannel{}

	for _, variant := range i.Variants {
		variants = append(variants, *variant.ToDatabase())
	}

	return &storage.DB_Item{
		ID: i.ID,
		ProfileID: profileId,
		TemplateID: i.TemplateID,
		Quantity: i.Quantity,
		Favorite: i.Favorite,
		HasSeen: i.HasSeen,
		Variants: variants,
	}
}

func (i *Item) Save() {
	if i.ProfileID == "" {
		return
	}

	for _, variant := range i.Variants {
		variant.Save()
	}

	storage.Repo.SaveItem(i.ToDatabase(i.ProfileID))
}

func (i *Item) ToGiftLootDatabase(giftId string) *storage.DB_GiftLoot {
	return &storage.DB_GiftLoot{
		GiftID: giftId,
		ProfileType: i.ProfileType,
		ID: i.ID,
		TemplateID: i.TemplateID,
		Quantity: i.Quantity,
	}
}

func (i *Item) ToPurchaseLootDatabase(purchaseId string) *storage.DB_PurchaseLoot {
	return &storage.DB_PurchaseLoot{
		ID: i.ID,
		PurchaseID: purchaseId,
		ProfileType: i.ProfileType,
		TemplateID: i.TemplateID,
		Quantity: i.Quantity,
	}
}

func (i *Item) ToReceiptLootDatabase(receiptId string) *storage.DB_ReceiptLoot {
	return &storage.DB_ReceiptLoot{
		ID: i.ID,
		ReceiptID: receiptId,
		ProfileType: i.ProfileType,
		TemplateID: i.TemplateID,
		Quantity: i.Quantity,
	}
}

func (i *Item) SaveLoot(giftId string) {
}

func (i *Item) Snapshot() ItemSnapshot {
	variants := []VariantChannel{}

	for _, variant := range i.Variants {
		variants = append(variants, *variant)
	}

	return ItemSnapshot{
		ID: i.ID,
		TemplateID: i.TemplateID,
		Quantity: i.Quantity,
		Favorite: i.Favorite,
		HasSeen: i.HasSeen,
		Variants: variants,
		ProfileType: i.ProfileType,
	}
}

type VariantChannel struct {
	ID 		string
	ItemID	string
	Channel string
	Owned	 	[]string
	Active	string
}

func FromDatabaseVariant(variant *storage.DB_VariantChannel) *VariantChannel {
	return &VariantChannel{
		ID: variant.ID,
		ItemID: variant.ItemID,
		Channel: variant.Channel,
		Owned: variant.Owned,
		Active: variant.Active,
	}
}

func (v *VariantChannel) ToDatabase() *storage.DB_VariantChannel {
	return &storage.DB_VariantChannel{
		ID: v.ID,
		ItemID: v.ItemID,
		Channel: v.Channel,
		Owned: v.Owned,
		Active: v.Active,
	}
}

func (v *VariantChannel) Save() {
	storage.Repo.SaveVariant(v.ToDatabase())
}