package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	telebot "gopkg.in/telebot.v3"
)

func (b *Bot) callbackModerateApprove(c telebot.Context) error {
	productID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID", ShowAlert: true})
	}

	_, err = b.productSvc.ModerateProduct(context.Background(), productID, dtos.ModerateProductRequest{
		Status: "active",
	})
	if err != nil {
		b.log.Error("bot: approve product failed", slog.String("error", err.Error()))
		return c.Respond(&telebot.CallbackResponse{Text: "❌ Ошибка при одобрении", ShowAlert: true})
	}

	c.Edit(fmt.Sprintf("✅ *Товар #%d одобрен*\n\nОпубликовано в каталог", productID), backKeyboard(), telebot.ModeMarkdown)
	return c.Respond(&telebot.CallbackResponse{Text: "✅ Одобрено"})
}

func (b *Bot) callbackModerateReject(c telebot.Context) error {
	productID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID", ShowAlert: true})
	}

	reason := "Отклонено администратором"
	_, err = b.productSvc.ModerateProduct(context.Background(), productID, dtos.ModerateProductRequest{
		Status: "rejected",
		Reason: &reason,
	})
	if err != nil {
		b.log.Error("bot: reject product failed", slog.String("error", err.Error()))
		return c.Respond(&telebot.CallbackResponse{Text: "❌ Ошибка при отклонении", ShowAlert: true})
	}

	c.Edit(fmt.Sprintf("❌ *Товар #%d отклонён*", productID), telebot.ModeMarkdown)
	return c.Respond(&telebot.CallbackResponse{Text: "❌ Отклонено"})
}

func (b *Bot) callbackAppApprove(c telebot.Context) error {
	appID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID", ShowAlert: true})
	}

	app, err := b.appSvc.Approve(context.Background(), appID)
	if err != nil {
		b.log.Error("bot: approve application failed", slog.String("error", err.Error()))
		errMsg := "❌ Ошибка при одобрении"
		if strings.Contains(err.Error(), "already") {
			errMsg = "⚠️ Заявка уже обработана"
		}
		return c.Respond(&telebot.CallbackResponse{Text: errMsg, ShowAlert: true})
	}

	c.Edit(fmt.Sprintf("✅ *Заявка #%d одобрена*\n\n🏪 Магазин «%s» создан", appID, escapeMarkdown(app.StoreName)), telebot.ModeMarkdown)
	return c.Respond(&telebot.CallbackResponse{Text: "✅ Одобрено"})
}

func (b *Bot) callbackAppReject(c telebot.Context) error {
	appID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID", ShowAlert: true})
	}

	_, err = b.appSvc.Reject(context.Background(), appID, "Отклонено администратором")
	if err != nil {
		b.log.Error("bot: reject application failed", slog.String("error", err.Error()))
		errMsg := "❌ Ошибка при отклонении"
		if strings.Contains(err.Error(), "already") {
			errMsg = "⚠️ Заявка уже обработана"
		}
		return c.Respond(&telebot.CallbackResponse{Text: errMsg, ShowAlert: true})
	}

	c.Edit(fmt.Sprintf("❌ *Заявка #%d отклонена*", appID), telebot.ModeMarkdown)
	return c.Respond(&telebot.CallbackResponse{Text: "❌ Отклонено"})
}