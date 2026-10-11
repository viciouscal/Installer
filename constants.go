/*
 * SPDX-License-Identifier: GPL-3.0
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"image/color"
	"vencordinstaller/buildinfo"
)

const ReleaseUrl = "https://api.github.com/repos/viciouscal/Vencord/releases/latest"
const ReleaseUrlFallback = "https://github.com/viciouscal/Vencord/releases/latest"
const InstallerReleaseUrl = "https://api.github.com/repos/viciouscal/Installer/releases/latest"
const InstallerReleaseUrlFallback = "https://github.com/viciouscal/Installer/releases/latest"

var UserAgent = "VencordARInstaller/" + buildinfo.InstallerGitHash + " (https://github.com/viciouscal/Installer)"

var (
	DiscordGreen        = color.RGBA{0, 133, 69, 0xff}
	DiscordGreenHovered = color.RGBA{0, 108, 55, 0xff}
	DiscordRed          = color.RGBA{210, 45, 57, 0xff}
	DiscordRedHovered   = color.RGBA{169, 35, 46, 0xff}
	DiscordBlue         = color.RGBA{88, 101, 242, 0xff}
	DiscordBlueHovered  = color.RGBA{68, 82, 187, 0xff}
	DiscordYellow       = color.RGBA{0xfe, 0xe7, 0x5c, 0xff}
)

var LinuxDiscordNames = []string{
	"Discord",
	"DiscordPTB",
	"DiscordCanary",
	"DiscordDevelopment",
	"discord",
	"discordptb",
	"discordcanary",
	"discorddevelopment",
	"discord-ptb",
	"discord-canary",
	"discord-development",

	"com.discordapp.Discord",
	"com.discordapp.DiscordPTB",
	"com.discordapp.DiscordCanary",
	"com.discordapp.DiscordDevelopment",
}
