package database

import (
	"database/sql"
	"fmt"
)

type categorySeed struct {
	name        string
	slug        string
	description string
}

var defaultCategories = []categorySeed{
	{"Horror", "horror", "Dark worlds, survival horror and games that keep you awake."},
	{"Adventure", "adventure", "Explore new worlds, stories and unforgettable journeys."},
	{"Action", "action", "Fast combat, explosive moments and pure adrenaline."},
	{"RPG", "rpg", "Characters, choices, builds and worlds worth getting lost in."},
	{"Shooter", "shooter", "FPS, third-person shooters, tactics and loadouts."},
	{"Strategy", "strategy", "Plan ahead, outthink opponents and control the battlefield."},
	{"Sports", "sports", "Football, basketball, competitive sports and more."},
	{"Racing", "racing", "Cars, circuits, speed and racing communities."},
	{"Simulation", "simulation", "Build, manage, fly, drive and simulate."},
	{"Survival", "survival", "Gather, craft, survive and make every resource count."},
	{"Puzzle", "puzzle", "Challenges that reward patience, logic and creativity."},
	{"Multiplayer", "multiplayer", "Play together, compete together and connect with other players."},
	{"Sandbox", "sandbox", "Create your own goals in worlds built for experimentation."},
	{"Indie", "indie", "Creative games, smaller studios and unexpected ideas."},
	{"Card & Board", "card-board", "Decks, boards, tactics and tabletop-inspired games."},
	{"VR", "vr", "Step inside the game with virtual reality."},
	{"Fantasy", "fantasy", "Magic, monsters, kingdoms and legendary adventures."},
	{"Sci-Fi", "sci-fi", "Space, technology, distant worlds and impossible futures."},
	{"Fighting", "fighting", "Combos, matchups, tournaments and competitive fighting games."},
	{"MMO", "mmo", "Massive worlds, guilds, raids and long-term adventures."},
	{"Retro", "retro", "Classic consoles, arcade legends and pixel nostalgia."},
}

func SeedCategories(db *sql.DB) error {
	for _, category := range defaultCategories {
		_, err := db.Exec(
			`
			INSERT OR IGNORE INTO categories (
				name,
				slug,
				description
			)
			VALUES (?, ?, ?)
			`,
			category.name,
			category.slug,
			category.description,
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
