package main

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/bot"
)

func subcommandHelp(subcommands []*bot.Subcommand, name string, showHidden bool) string {
	var found *bot.Subcommand
	for _, subcommand := range subcommands {
		if subcommand.Command == name {
			found = subcommand
			break
		}
	}

	if found == nil {
		return ""
	}

	help := found.HelpShowHidden(showHidden)
	if help == "" {
		return ""
	}

	builder := strings.Builder{}
	builder.WriteString("**")
	builder.WriteString(found.Command)
	builder.WriteString("**")

	for _, alias := range found.Aliases {
		builder.WriteString("|")
		builder.WriteString("**")
		builder.WriteString(alias)
		builder.WriteString("**")
	}

	if found.Description != "" {
		builder.WriteString(": ")
		builder.WriteString(found.Description)
	}

	builder.WriteByte('\n')
	builder.WriteString(bot.IndentLines(help))

	return builder.String()
}

func contextHelp(ctx *bot.Context, showHidden bool) string {
	// Generate the header.
	buf := strings.Builder{}
	buf.WriteString("__Help__")

	// Name an
	if ctx.Name != "" {
		buf.WriteString(": " + ctx.Name)
	}
	if ctx.Description != "" {
		buf.WriteString("\n" + bot.IndentLines(ctx.Description))
	}

	// Separators
	buf.WriteString("\n---\n")

	// Generate all commands
	if help := ctx.Subcommand.Help(); help != "" {
		buf.WriteString("__Commands__\n")
		buf.WriteString(bot.IndentLines(help))
		buf.WriteByte('\n')
	}

	var subcommands = ctx.Subcommands()
	var subhelps = make([]string, 0, len(subcommands))

	for _, sub := range subcommands {
		if sub.Hidden && !showHidden {
			continue
		}

		// Not efficient; whatever.
		if sub.HelpShowHidden(showHidden) == "" {
			continue
		}

		builder := strings.Builder{}
		builder.WriteString("**")
		builder.WriteString(sub.Command)
		builder.WriteString("**")

		for _, alias := range sub.Aliases {
			builder.WriteString("|")
			builder.WriteString("**")
			builder.WriteString(alias)
			builder.WriteString("**")
		}

		if sub.Description != "" {
			builder.WriteString(": ")
			builder.WriteString(sub.Description)
		}

		subhelps = append(subhelps, builder.String())
	}

	if len(subhelps) > 0 {
		buf.WriteString("---\n")
		buf.WriteString("__Subcommands__\n")
		buf.WriteString(bot.IndentLines(strings.Join(subhelps, "\n")))
	}

	return buf.String()
}
