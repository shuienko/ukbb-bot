# UKBB Bot
Weather Info from Kyiv Boryspil Airport radar

## about
Radar info is here: https://meteoinfo.by/radar/?q=UKBB

## prerequisites
- Create Telegram bot and get `UKBB_BOT_TOKEN` first. 
See instructions here: https://core.telegram.org/bots#6-botfather

- As an external user storage bot needs Amazon DynamoDB table.
Please create it beforehand and update `tableName` and `AWSRegion` variables accordignly.

- Also you need to meke sure bot process is able to put/get info to/from DynamoDB.
Use either AWS IAM role or set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment varibles

## depployment and run
### go
```bash
export UKBB_BOT_TOKEN="<your token here>"
export AWS_ACCESS_KEY_ID="<access key>"
export AWS_SECRET_ACCESS_KEY="<secret key>"
go build && ./ukbb-bot
```

### docker
```bash
docker build . -t ukbb-bot
docker run -it --env UKBB_BOT_TOKEN="<Telegram Bot API Token>"  --env AWS_ACCESS_KEY_ID="<...>" --env AWS_SECRET_ACCESS_KEY="..." ukbb-bot
```