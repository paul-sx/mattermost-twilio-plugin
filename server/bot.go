package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type twilioBot struct {
	*model.Bot
}

func (p *TwilioPlugin) getBot() (*twilioBot, error) {
	if p.bot == nil || p.bot.IsValid() != nil {
		bot, err := p.initializeBot()
		if err != nil {
			return nil, errors.Wrap(err, "Could not initialize bot")
		}
		p.bot = bot
		return bot, nil
	}
	return p.bot, nil
}

func (p *TwilioPlugin) initializeBot() (*twilioBot, error) {
	config := p.getConfiguration()
	UserId := config.InstallUserId

	user, userErr := p.API.GetUser(UserId)
	if userErr != nil || user == nil {
		// Find any user with system administrator role
		userList, listErr := p.API.GetUsers(&model.UserGetOptions{
			Role:    "system_admin",
			Page:    0,
			PerPage: 1,
		})
		if listErr != nil || len(userList) == 0 {
			return nil, errors.Wrap(listErr, "Could not find system admin user to own the bot")
		}

		UserId = userList[0].Id
	}

	botGetOptions := &model.BotGetOptions{
		OwnerId:        "",
		IncludeDeleted: false,
		OnlyOrphaned:   false,
		Page:           0,
		PerPage:        20,
	}
	for bots, err := p.API.GetBots(botGetOptions); err == nil && len(bots) > 0; bots, err = p.API.GetBots(botGetOptions) {
		botGetOptions.Page++
		for _, bot := range bots {
			if bot.Username == "twilio" && bot.IsValid() == nil {
				return &twilioBot{
					bot,
				}, nil
			}
		}
	}
	bot := &model.Bot{
		Username:    "twilio",
		DisplayName: "Twilio",
		Description: "Twilio Bot",
		OwnerId:     UserId,
	}

	if err := bot.IsValidCreate(); err != nil {
		return nil, err
	}
	bot, err := p.API.CreateBot(bot)
	if err != nil {
		return nil, err
	}
	return &twilioBot{
		Bot: bot,
	}, nil

}
