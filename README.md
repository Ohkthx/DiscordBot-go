### Discord Bot
+ Discord API: [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo)
+ MySQL Drivers: [my-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
+ Vita-Nex: Core API [d0x1p2/VNCgo](https://github.com/d0x1p2/VNCgo)

This is my first attempt at the Go/Golang and also writing a Discord Bot. It requires both of the above projects.
This wouldn't have been possible for me to do without contributors of the projects above, so I greatly appreciate their work.
*Vita-Nex: Core API is maintained by myself.

### Features
+ Console Interface
+ Remote capabilities
+ SQL DB to store commands between runs
+ On-the-fly runtime abilities to add additional commands
+ Checks permissions to allow access for commands
+ Error checking, handling, and logging

### Some of the TODO list
+ Upload the SQL DB format for all tables.
+ Additional built-in commands.
+ ,yay and ,nay options
+ list all DMs (UserChannels())
+ Greet on "Join"
+ Database containing users and last-seen
+ Allow vendor owners to update commands
+ google/wiki search in-line
+ Linking commmands to each other
+ User and Channel mentions. @1234567890
+ HELP increase
+ Organize SQL processing better. Possible split
+ Add an official "client" (using plain-text / netcat atm)
+ Launch additional process to monitor DM/PMs
+ Command DUMP/HELP thru PM
+ Increment "uses" in commands table
+ SQL setup PER triggers. No more global. -> Large overhaul, maybe v1.0

#### Versions
Current: v1.1.3
See [changelog](https://github.com/d0x1p2/DiscordBot-go/blob/master/changelog)