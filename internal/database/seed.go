package database

import (
	"database/sql"
	"fmt"
)

type categorySeed struct {
	name        string
	slug        string
	description string
	tagline     string
}

var defaultCategories = []categorySeed{
	{
		"Horror",
		"horror",
		"Dark worlds, survival horror and games that keep you awake.",
		"Turn off the lights. Keep the volume up.",
	},
	{
		"Adventure",
		"adventure",
		"Explore new worlds, stories and unforgettable journeys.",
		"The map ends here. Your story doesn't.",
	},
	{
		"Action",
		"action",
		"Fast combat, explosive moments and pure adrenaline.",
		"Less talking. More explosions.",
	},
	{
		"RPG",
		"rpg",
		"Characters, choices, builds and worlds worth getting lost in.",
		"One more quest. Then we sleep. Probably.",
	},
	{
		"Shooter",
		"shooter",
		"FPS, third-person shooters, tactics and loadouts.",
		"Aim first. Argue about the meta later.",
	},
	{
		"Strategy",
		"strategy",
		"Plan ahead, outthink opponents and control the battlefield.",
		"Winning starts before the first move.",
	},
	{
		"Sports",
		"sports",
		"Football, basketball, competitive sports and more.",
		"Same game. A thousand opinions.",
	},
	{
		"Racing",
		"racing",
		"Cars, circuits, speed and racing communities.",
		"Brake later. Debate faster.",
	},
	{
		"Simulation",
		"simulation",
		"Build, manage, fly, drive and simulate.",
		"Why live one life when you can simulate twenty?",
	},
	{
		"Survival",
		"survival",
		"Gather, craft, survive and make every resource count.",
		"Everything wants you dead. Good luck.",
	},
	{
		"Puzzle",
		"puzzle",
		"Challenges that reward patience, logic and creativity.",
		"There is always a solution. Usually.",
	},
	{
		"Multiplayer",
		"multiplayer",
		"Play together, compete together and connect with other players.",
		"Friendships tested. Lobbies blamed.",
	},
	{
		"Sandbox",
		"sandbox",
		"Create your own goals in worlds built for experimentation.",
		"No rules. No path. Your world.",
	},
	{
		"Indie",
		"indie",
		"Creative games, smaller studios and unexpected ideas.",
		"Small studios. Big ideas.",
	},
	{
		"Card & Board",
		"card-board",
		"Decks, boards, tactics and tabletop-inspired games.",
		"Every move tells a story. Make yours count.",
	},
	{
		"VR",
		"vr",
		"Step inside the game with virtual reality.",
		"Why watch the game when you can step inside?",
	},
	{
		"Fantasy",
		"fantasy",
		"Magic, monsters, kingdoms and legendary adventures.",
		"Dragons, magic, and questionable side quests.",
	},
	{
		"Sci-Fi",
		"sci-fi",
		"Space, technology, distant worlds and impossible futures.",
		"The future is weird. We should talk about it.",
	},
	{
		"Fighting",
		"fighting",
		"Combos, matchups, tournaments and competitive fighting games.",
		"Settle it in the next round.",
	},
	{
		"MMO",
		"mmo",
		"Massive worlds, guilds, raids and long-term adventures.",
		"Thousands of players. One more grind.",
	},
	{
		"Retro",
		"retro",
		"Classic consoles, arcade legends and pixel nostalgia.",
		"Old pixels. Timeless arguments.",
	},
}

func SeedCategories(db *sql.DB) error {
	for _, category := range defaultCategories {
		_, err := db.Exec(
			`
			INSERT OR IGNORE INTO categories (
				name,
				slug,
				description,
				tagline
			)
			VALUES (?, ?, ?, ?)
			`,
			category.name,
			category.slug,
			category.description,
			category.tagline,
		)
		if err != nil {
			return fmt.Errorf(
				"seed category %q: %w",
				category.name,
				err,
			)
		}
	}

	return nil
}
