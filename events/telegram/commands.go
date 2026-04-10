package telegram

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"telegrambot/lib/e"
	"telegrambot/storage"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
	ListCmd  = "/list"
	Taslim   = "Ва Алейкум Ас Салям👋"
)

func (p *Processor) doCmd(text string, chatId int, username string, chatType string) error {
	text = strings.TrimSpace(text)

	if username == "" {
		return p.tg.SendMessage(chatId, msgNoUsername)
	}

	log.Printf("get new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(chatId, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatId, username)
	case HelpCmd:
		return p.sendHelp(chatId)
	case StartCmd:
		return p.sendHello(chatId)
	case ListCmd:
		return p.sendList(chatId, username)
	default:
		if isGreeting(text) {
			return p.tg.SendMessage(chatId, Taslim)
		}
		if chatType == "private" {
			return p.tg.SendMessage(chatId, msgUnknownCommand)
		}
		return nil
	}
}

func (p *Processor) savePage(chatId int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(context.Background(), page)
	if err != nil {
		return err
	}
	if isExists {
		return p.tg.SendMessage(chatId, msgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatId, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatId int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatId, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatId, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) sendList(chatId int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send all saved pages", err) }()

	pages, err := p.storage.List(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatId, msgNoSavedPages)
	}
	var urls strings.Builder

	for _, page := range pages {
		urls.WriteString(page.URL + "\n")
	}

	if err := p.tg.SendMessage(chatId, urls.String()); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendHelp(chatId int) error {
	return p.tg.SendMessage(chatId, msgHelp)
}

func (p *Processor) sendHello(chatId int) error {
	return p.tg.SendMessage(chatId, msgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)
	return err == nil && u.Host != "" // обрабатывает только ссылки с протоколом
}

func isGreeting(text string) bool {
	return strings.Contains(strings.ToLower(text), "алейкум")
}
