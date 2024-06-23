# Discord-outdate-delete-bot
Discord bot for deleting messages after a specified period of time
> With this bot you can set the time that messages will exist in the channel.  
> If you want to prevent some messages from being deleted, pin them.

### Bot commands:  
* **/set-timeout** - Set the time after which messages will be deleted
* **/info-timeout** - View the time after which messages will be deleted
* **/remove-timeout** - Stop deleting messages

# Using a deployed bot
You can try or fully use the bot by inviting it to your discord server **(the bot may not be available)**:  
https://discord.com/oauth2/authorize?client_id=1248834167882518579

# Deployment (docker)
1. Create folder to store files between bot runs _(or you can use docker volumes)_
2. Place the **config.ini** file in this folder and fill it ([example](https://github.com/mdpakhmurin/discord-outdate-delete-bot/blob/main/app/data/config_example.ini))
3. Download the build from docker:
```
docker pull mdpakhmurin/discord-outdate-delete-bot:latest
```
4. Run it:
```
docker run --mount type=bind,src="ABSOLUTE/PATH/TO/CREATED/FOLDER",target=/app/data mdpakhmurin/discord-outdate-delete-bot
```
# Deploymet (source code)
1. Clone or download repository:
```
git clone https://github.com/mdpakhmurin/discord-outdate-delete-bot.git
```
2. Go to the app directory
3. Run:
```
go run .
```
