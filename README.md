# C.R.A.I.G - Your Personal Agent
- Run on your own computer
- Learns and remembers details about you and your preferences
- Simple interaction through discord

## Installation
### Create Discord Bot
- [Create a discord bot at](https://discord.com/developers/applications)
- Give your bot "Privileged Gateway Intents" in the "Bot" menu of your discord app
- In the "OAth2" menu, select "bot" then select "Administrator" (If you want you can figure out exactly what perms it needs but admin is simplest)
- Copy the link below the checkboxes, paste it in a browser, and add the bot to your server (you may want  to create a server with just you and the bot in it for now)
- Reset the token in the "Bot" menu, copy it, and store in an evironment var `CRAIG_DISCORD_TOKEN` on your host machine
### Create OpenAI / Gemini Developer Key
- Copy your openai token from openai developer and put it an environment variable `OPENAI_KEY` on the host (optional)
- Copy your gemini token from google cloud developer and put it an environment variable `GEMINI_KEY` on the host (optional)
### Run
- Clone this repo, and in the repo run:
    - `CRAIG_INIT=yes docker compose up` for your first time run - this will also setup the CRAIG data directory
    - `docker compose up` for all susequent runs
- You can also add the `-d` flag onto the end of either of those to run in the background
- All data will be mounted at `/craig-data`
    - Add claude-code style skills at skills/
    - Change the models that are used at models/
    - See the agent's current scratchpad at scratchpad.txt