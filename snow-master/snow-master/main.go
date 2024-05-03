package main

import (
	_ "embed"
	"fmt"

	"github.com/ectrc/snow/aid"
	"github.com/ectrc/snow/discord"
	"github.com/ectrc/snow/fortnite"
	"github.com/ectrc/snow/handlers"
	"github.com/ectrc/snow/person"
	"github.com/ectrc/snow/shop"
	"github.com/ectrc/snow/storage"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

//go:embed config.ini
var configFile []byte

func init() {
	aid.LoadConfig(configFile) 
	var device storage.Storage
	switch aid.Config.Database.Type {
	case "postgres":
		postgresStorage := storage.NewPostgresStorage()
		if aid.Config.Database.DropAllTables {
			postgresStorage.DropTables()
			aid.Print("(snow) all tables dropped and reset")
		}
		postgresStorage.MigrateAll()
		device = postgresStorage
	default:
		panic("Invalid database type: " + aid.Config.Database.Type)
	}

	storage.Repo = storage.NewStorage(device)

	if aid.Config.Amazon.Enabled {
		storage.Repo.Amazon = storage.NewAmazonClient(aid.Config.Amazon.BucketURI, aid.Config.Amazon.AccessKeyID, aid.Config.Amazon.SecretAccessKey, aid.Config.Amazon.ClientSettingsBucket)
	}
}

func init() {
	discord.IntialiseClient()
	fortnite.PreloadCosmetics()
	fortnite.PreloadEvents()
	shop.GetShop()

	for _, username := range aid.Config.Accounts.Gods {
		found := person.FindByDisplay(username)
		if found == nil {
			found = fortnite.NewFortnitePersonWithId(username, username, aid.Config.Fortnite.Everything)
		}
		found.Discord = &storage.DB_DiscordPerson{
			ID: found.ID,
			PersonID: found.ID,
			Username: username,
		}
		found.Save()

		found.AddPermission(person.PermissionAllWithRoles)
		aid.Print("(snow) max account " + username + " loaded")
	}

	for _, username := range aid.Config.Accounts.Owners {
		found := person.FindByDisplay(username)
		if found == nil {
			continue
		}

		found.AddPermission(person.PermissionAllWithRoles)
		aid.Print("(snow) owner account " + username + " loaded")
	}
}

func main() {
	r := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	r.Use(aid.FiberLogger())
	r.Use(aid.FiberLimiter(1000))
	r.Use(aid.FiberCors())

	r.Get("/region", handlers.GetRegion)
	r.Get("/content/api/pages/fortnite-game", handlers.GetContentPages)
	r.Get("/waitingroom/api/waitingroom", handlers.GetWaitingRoomStatus)
	r.Get("/affiliate/api/public/affiliates/slug/:slug", handlers.GetAffiliate)
	r.Get("/profile/privacy_settings", handlers.MiddlewareFortnite, handlers.GetPrivacySettings)
	r.Put("/profile/play_region", handlers.AnyNoContent)
	r.Get("/api/v1/search/:accountId", handlers.GetPersonSearch)
	r.Get("/", handlers.RedirectSocket)
	r.Get("/socket", handlers.MiddlewareWebsocket, websocket.New(handlers.WebsocketConnection))

	account := r.Group("/account/api")
	account.Get("/public/account", handlers.GetPublicAccounts)
	account.Get("/public/account/:accountId", handlers.GetPublicAccount)
	account.Get("/public/account/:accountId/externalAuths", handlers.GetPublicAccountExternalAuths)
	account.Get("/public/account/displayName/:displayName", handlers.GetPublicAccountByDisplayName)
	account.Get("/oauth/verify", handlers.GetTokenVerify)
	account.Post("/oauth/token", handlers.PostFortniteToken)
	account.Delete("/oauth/sessions/kill", handlers.DeleteToken)

	fortnite := r.Group("/fortnite/api")
	fortnite.Get("/receipts/v1/account/:accountId/receipts", handlers.MiddlewareFortnite, handlers.GetFortniteReceipts)
	fortnite.Get("/v2/versioncheck/:version", handlers.GetFortniteVersion)
	fortnite.Get("/calendar/v1/timeline", handlers.GetFortniteTimeline)

	storefront := fortnite.Group("/storefront/v2")
	storefront.Use(handlers.MiddlewareFortnite)
	storefront.Use(aid.FiberLimiter(4))
	storefront.Get("/catalog", handlers.GetStorefrontCatalog)
	storefront.Get("/keychain", handlers.GetStorefrontKeychain)
	r.Get("/catalog/api/shared/bulk/offers", handlers.GetStorefrontCatalogBulkOffers)

	matchmaking := fortnite.Group("/matchmaking")
	matchmaking.Get("/session/findPlayer/:accountId", handlers.GetMatchmakingSession)

	storage := fortnite.Group("/cloudstorage")
	storage.Get("/system", handlers.GetCloudStorageFiles)
	storage.Get("/system/config", handlers.GetCloudStorageConfig)
	storage.Get("/system/:fileName", handlers.GetCloudStorageFile)
	storage.Get("/user/:accountId", handlers.MiddlewareFortnite, handlers.GetUserStorageFiles)
	storage.Get("/user/:accountId/:fileName", handlers.MiddlewareFortnite, handlers.GetUserStorageFile)
	storage.Put("/user/:accountId/:fileName", handlers.MiddlewareFortnite, handlers.PutUserStorageFile)

	friends := r.Group("/friends/api")
	friends.Use(handlers.MiddlewareFortnite)
	friends.Get("/public/friends/:accountId", handlers.GetFriendList)
	friends.Post("/public/friends/:accountId/:wanted", handlers.PostCreateFriend)
	friends.Delete("/public/friends/:accountId/:wanted", handlers.DeleteFriend)
	friends.Get("/:version/:accountId/summary", handlers.GetFriendListSummary)
	friends.Post("/:version/:accountId/friends/:wanted", handlers.PostCreateFriend)
	friends.Delete("/:version/:accountId/friends/:wanted", handlers.DeleteFriend)

	events := r.Group("/api/v1/events/Fortnite")
	events.Use(handlers.MiddlewareFortnite)
	events.Get("/download/:accountId", handlers.GetEvents)
	events.Get("/:eventId/history/:accountId", handlers.GetEventsBulkHistory)

	game := fortnite.Group("/game/v2")
	game.Get("/enabled_features", handlers.GetGameEnabledFeatures)
	game.Post("/tryPlayOnPlatform/account/:accountId", handlers.PostGamePlatform)
	game.Post("/grant_access/:accountId", handlers.PostGameAccess)
	game.Post("/profileToken/verify/:accountId", handlers.AnyNoContent)

	profile := game.Group("/profile/:accountId")
	profile.Use(handlers.MiddlewareFortnite)
	profile.Post("/client/:action", handlers.PostClientProfileAction)
	profile.Post("/dedicated_server/:action", handlers.PostServerProfileAction)

	lightswitch := r.Group("/lightswitch/api")
	lightswitch.Use(handlers.MiddlewareFortnite)
	lightswitch.Get("/service/bulk/status", handlers.GetLightswitchBulkStatus)

	purchasing := r.Group("/purchase")
	purchasing.Get("/", handlers.GetHtmlPurchasePage)
	purchasing.Get("/offer", handlers.MiddlewareFortnite, handlers.GetPurchaseOffer)
	purchasing.Post("/offer", handlers.MiddlewareFortnite, handlers.PostPurchaseOffer)
	purchasing.Get("/assets", handlers.GetPurchaseAsset)

	party := r.Group("/party/api/v1/Fortnite")
	party.Use(handlers.MiddlewareFortnite)
	party.Get("/user/:accountId", handlers.GetPartiesForUser)
	party.Get("/user/:accountId/settings/privacy", handlers.GetPartyUserPrivacy)
	party.Get("/user/:accountId/notifications/undelivered/count", handlers.GetPartyNotifications)
	party.Get("/user/:accountId/pings/:friendId/parties", handlers.GetPartyPingsFromFriend)
	party.Post("/user/:accountId/pings/:friendId/join", handlers.PostPartyJoinFromPing)
	party.Delete("/user/:accountId/pings/:friendId", handlers.PostPartyDeletePings)
	party.Get("/parties/:partyId", handlers.GetPartyForMember)
	party.Post("/parties", handlers.PostPartyCreate)
	party.Post("/parties/:partyId/invites/:accountId", handlers.PostPartyInvite)
	party.Post("/parties/:partyId/members/:accountId/join", handlers.PostPartyJoin)
	party.Post("/parties/:partyId/members/:accountId/promote", handlers.PostPartyPromoteMember)
	party.Patch("/parties/:partyId", handlers.PatchPartyUpdateState)
	party.Patch("/parties/:partyId/members/:accountId/meta", handlers.PatchPartyUpdateMemberState)
	party.Delete("/parties/:partyId/members/:accountId", handlers.DeletePartyMember)
	party.Post("/members/:friendId/intentions/:accountId", handlers.PostPartyCreateIntention)

	snow := r.Group("/snow")
	snow.Post("/log", handlers.PostSnowLog)

	discord := snow.Group("/discord")
	discord.Get("/", handlers.GetDiscordOAuthURL)

	launcher := snow.Group("/launcher")
	launcher.Get("/", handlers.GetLauncherStatus)

	player := snow.Group("/player")
	player.Use(handlers.MiddlewareWeb)
	player.Get("/", handlers.GetPlayer)
	player.Get("/okay", handlers.GetPlayerOkay)
	player.Post("/code", handlers.PostPlayerCreateCode)

	debug := snow.Group("/")
	debug.Use(handlers.MiddlewareOnlyDebug)
	debug.Get("/cache", handlers.GetSnowCachedPlayers)
	debug.Get("/sockets", handlers.GetSnowConnectedSockets)
	debug.Get("/cosmetics", handlers.GetSnowPreloadedCosmetics)
	debug.Get("/parties", handlers.GetSnowParties)
	debug.Get("/shop", handlers.GetSnowShop)

	r.Hooks().OnListen(func(ld fiber.ListenData) error {
		aid.Print("(fiber) listening on " + aid.Config.API.Host + ":" + ld.Port)
		return nil
	})
	r.All("*", func(c *fiber.Ctx) error { return c.Status(fiber.StatusNotFound).JSON(aid.ErrorNotFound) })

	// if aid.Config.Fortnite.Season <= 2 {
	// 	t := handlers.NewServer()
	// 	go t.Listen()
	// }

	err := r.Listen("0.0.0.0" + aid.Config.API.Port)
	if err != nil {
		panic(fmt.Sprintf("(fiber) failed to listen: %v", err))
	}
}
