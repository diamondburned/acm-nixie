package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/bot/extras/arguments"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/pkg/errors"
	"gitlab.com/diamondburned/nixie/db"

	"gitlab.com/diamondburned/nixie/commands/aocboard"
	"gitlab.com/diamondburned/nixie/commands/playground"

	// Playground imports
	_ "gitlab.com/diamondburned/nixie/commands/playground/golang"
	_ "gitlab.com/diamondburned/nixie/commands/playground/rextester"
	_ "gitlab.com/diamondburned/nixie/commands/playground/rust"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("$BOT_TOKEN empty")
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		log.Fatalln("$DATABASE_PATH empty")
	}

	if err := os.MkdirAll(databasePath, 0750); err != nil {
		log.Fatalln("Failed to create database directory:", err)
	}

	kv, err := db.NewKVFile(filepath.Join(databasePath + "nixie.db"))
	if err != nil {
		log.Fatalln("Failed to create a database:", err)
	}
	defer kv.Close()

	wait, err := bot.Start(token, &Commands{}, func(ctx *bot.Context) error {
		return initBot(ctx, kv)
	})
	if err != nil {
		log.Fatalln("Bot cannot start:", err)
	}

	log.Println("Bot started")

	if err := wait(); err != nil {
		log.Fatalln("Bot failed:", err)
	}
}

func initBot(ctx *bot.Context, kv *db.KV) error {
	ctx.AddIntents(gateway.IntentGuilds)
	ctx.AddIntents(gateway.IntentGuildMessages)
	ctx.AddIntents(gateway.IntentGuildMessageReactions)

	u, err := ctx.Me()
	if err != nil {
		return errors.Wrap(err, "failed to get current user")
	}

	ctx.HasPrefix = bot.NewPrefix(
		"n!",
		fmt.Sprintf("<@%d> ", u.ID),
		fmt.Sprintf("<@!%d> ", u.ID),
	)

	ctx.FormatError = func(err error) string {
		return "Error: " + err.Error()
	}

	// If there is not a valid command, then don't say anything.  We'd want to
	// tell the user that a subcommand is wrong when the command is right.
	ctx.SilentUnknown.Command = true

	// Treat message edits the same as message creates for convenience.
	ctx.EditableCommands = true

	// Share the same bot database.
	ctx.MustRegisterSubcommand(aocboard.New(kv.Node("aocboard")), "aoc")
	ctx.MustRegisterSubcommand(playground.New())

	return nil
}

type Commands struct {
	Ctx *bot.Context
}

func (c *Commands) Setup(sub *bot.Subcommand) {
	sub.FindCommand("Help").Description = "Print this help."
	sub.FindCommand(c.Help).Arguments[0].String = "[subcommand]"
}

func (c *Commands) Help(m *gateway.MessageCreateEvent, f arguments.Flag) error {
	var help string
	var showHidden bool

	fs := arguments.NewFlagSet()
	fs.BoolVar(&showHidden, "all", false, "Show hidden commands")

	if err := f.With(fs.FlagSet); err != nil {
		return errors.Wrap(err, "invalid usage")
	}

	p, _ := c.Ctx.Permissions(m.ChannelID, m.Author.ID)
	showHidden = showHidden && p.Has(discord.PermissionAdministrator)

	switch len(fs.Args()) {
	case 0:
		help = contextHelp(c.Ctx, showHidden)
	case 1:
		help = subcommandHelp(c.Ctx.Subcommands(), fs.Arg(0), showHidden)
	case 2:
		return errors.New("unexpected arguments given")
	}

	if help == "" {
		return errors.New("unknown subcommand")
	}

	if _, err := c.Ctx.SendText(m.ChannelID, help); err != nil {
		return errors.Wrap(err, "failed to send help message")
	}

	return nil
}
