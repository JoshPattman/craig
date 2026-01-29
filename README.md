# C.R.A.I.G - Your Personal Agent
- Run on your own computer
- Learns and remembers details about you and your preferences
- Simple interaction through discord

## Installation
- [Create a discord bot at](https://discord.com/developers/applications)
- Give your bot "Privileged Gateway Intents" in the "Bot" menu of your discord app
- In the "OAth2" menu, select "bot" then select "Administrator" (If you want you can figure out exactly what perms it needs but admin is simplest)
- Copy the link below the checkboxes, paste it in a browser, and add the bot to your server (you may want  to create a server with just you and the bot in it for now)
- Reset the token in the "Bot" menu, copy it, and store in an evironment var `CRAIG_DISCORD_TOKEN` on your host machine
- Copy your openai token from openai developer and put it an environment variable `OPENAI_KEY` on the host too
- Clone this repo, and in the repo run `docker compose up -d`