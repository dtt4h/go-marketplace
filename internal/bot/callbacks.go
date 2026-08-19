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

func moderationKeyboard(productID int64) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("✅ Одобрить", "mod_approve", strconv.FormatInt(productID, 10)),
			rm.Data("❌ Отклонить", "mod_reject", strconv.FormatInt(productID, 10)),
		),
	)
	return rm
}

func applicationKeyboard(appID int64) *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data("✅ Одобрить", "app_approve", strconv.FormatInt(appID, 10)),
			rm.Data("❌ Отклонить", "app_reject", strconv.FormatInt(appID, 10)),
		),
	)
	return rm
}

func (b *Bot) callbackModerateApprove(c telebot.Context) error {
	productID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID"})
	}

	_, err = b.productSvc.ModerateProduct(context.Background(), productID, dtos.ModerateProductRequest{
		Status: "active",
	})
	if err != nil {
		b.log.Error("bot: approve product failed", slog.String("error", err.Error()))
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка при одобрении"})
	}

	c.Edit(fmt.Sprintf("✅ Товар #%d одобрен и опубликован", productID))
	return c.Respond()
}

func (b *Bot) callbackModerateReject(c telebot.Context) error {
	productID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID"})
	}

	reason := "Отклонено администратором"
	_, err = b.productSvc.ModerateProduct(context.Background(), productID, dtos.ModerateProductRequest{
		Status: "rejected",
		Reason: &reason,
	})
	if err != nil {
		b.log.Error("bot: reject product failed", slog.String("error", err.Error()))
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка при отклонении"})
	}

	c.Edit(fmt.Sprintf("❌ Товар #%d отклонён", productID))
	return c.Respond()
}

func (b *Bot) callbackAppApprove(c telebot.Context) error {
	appID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID"})
	}

	app, err := b.appSvc.Approve(context.Background(), appID)
	if err != nil {
		b.log.Error("bot: approve application failed", slog.String("error", err.Error()))
		errMsg := "Ошибка при одобрении"
		if strings.Contains(err.Error(), "already") {
			errMsg = "Заявка уже обработана"
		}
		return c.Respond(&telebot.CallbackResponse{Text: errMsg})
	}

	c.Edit(fmt.Sprintf("✅ Заявка #%d одобрена\n🏪 Магазин «%s» создан", appID, app.StoreName))
	return c.Respond()
}

func (b *Bot) callbackAppReject(c telebot.Context) error {
	appID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Некорректный ID"})
	}

	_, err = b.appSvc.Reject(context.Background(), appID, "Отклонено администратором")
	if err != nil {
		b.log.Error("bot: reject application failed", slog.String("error", err.Error()))
		errMsg := "Ошибка при отклонении"
		if strings.Contains(err.Error(), "already") {
			errMsg = "Заявка уже обработана"
		}
		return c.Respond(&telebot.CallbackResponse{Text: errMsg})
	}

	c.Edit(fmt.Sprintf("❌ Заявка #%d отклонена", appID))
	return c.Respond()
}