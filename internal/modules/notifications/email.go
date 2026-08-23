package notifications

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/config"
)

// EmailService defines email sending capabilities.
type EmailService interface {
	SendOrderConfirmation(to, customerName, orderID, total, items string) error
	SendOrderStatusUpdate(to, customerName, orderID, status string) error
	SendSellerNewOrder(to, sellerName, orderID, total, customerEmail, customerName string) error
	SendReceipt(to, customerName, orderID, total, items, receiptHTML string) error
	SendSellerApplicationApproved(to, username, storeName string) error
	SendSellerApplicationRejected(to, username, reason string) error
	SendNewApplicationNotification(to, username, storeName, description string) error
}

type emailService struct {
	cfg      config.SMTPConfig
	fromName string
	log      *slog.Logger
}

// NewEmailService creates a new email service.
func NewEmailService(cfg config.SMTPConfig, log *slog.Logger) EmailService {
	return &emailService{
		cfg:      cfg,
		fromName: cfg.FromName,
		log:      log,
	}
}

func (s *emailService) send(to string, subject string, body string) error {
	if s.cfg.Host == "" {
		s.log.Warn("SMTP not configured, skipping email", "to", to, "subject", subject)
		return nil
	}

	from := s.cfg.FromName + " <" + s.cfg.FromEmail + ">"
	toAddrs := strings.Split(to, ";")

	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	var headers = map[string]string{
		"From":                      from,
		"To":                        to,
		"Subject":                   subject,
		"MIME-Version":              "1.0",
		"Content-Type":              "text/html; charset=UTF-8",
		"Content-Transfer-Encoding": "7bit",
	}

	headerParts := make([]string, 0, len(headers))
	for k, v := range headers {
		headerParts = append(headerParts, k+": "+v)
	}

	message := strings.Join(headerParts, "\r\n") + "\r\n\r\n" + body

	addr := s.cfg.Host + ":" + fmt.Sprintf("%d", s.cfg.Port)

	var err error
	switch s.cfg.Encryption {
	case "tls":
		tlsConfig := &tls.Config{
			ServerName: s.cfg.Host,
		}
		conn, tlsErr := tls.Dial("tcp", addr, tlsConfig)
		if tlsErr != nil {
			return fmt.Errorf("tls dial: %w", tlsErr)
		}
		client, clientErr := smtp.NewClient(conn, s.cfg.Host)
		if clientErr != nil {
			return fmt.Errorf("smtp new client: %w", clientErr)
		}
		err = client.Auth(auth)
		if err == nil {
			err = client.Mail(s.cfg.FromEmail)
		}
		if err == nil {
			for _, addr := range toAddrs {
				if sendErr := client.Rcpt(strings.TrimSpace(addr)); sendErr != nil {
					err = sendErr
					break
				}
			}
		}
		if err == nil {
			w, sendErr := client.Data()
			if sendErr != nil {
				err = sendErr
			} else {
				_, writeErr := w.Write([]byte(message))
				if writeErr != nil {
					err = writeErr
				} else {
					err = w.Close()
				}
			}
		}
		_ = conn.Close()

	case "starttls":
		client, clientErr := smtp.Dial(addr)
		if clientErr != nil {
			return fmt.Errorf("smtp dial: %w", clientErr)
		}
		if clientErr = client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); clientErr != nil {
			return fmt.Errorf("starttls: %w", clientErr)
		}
		err = client.Auth(auth)
		if err == nil {
			err = client.Mail(s.cfg.FromEmail)
		}
		if err == nil {
			for _, addr := range toAddrs {
				if sendErr := client.Rcpt(strings.TrimSpace(addr)); sendErr != nil {
					err = sendErr
					break
				}
			}
		}
		if err == nil {
			w, sendErr := client.Data()
			if sendErr != nil {
				err = sendErr
			} else {
				_, writeErr := w.Write([]byte(message))
				if writeErr != nil {
					err = writeErr
				} else {
					err = w.Close()
				}
			}
		}
		_ = client.Close()

	default:
		err = smtp.SendMail(addr, auth, s.cfg.FromEmail, toAddrs, []byte(message))
	}

	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	s.log.Info("email sent", "to", to, "subject", subject)
	return nil
}

func (s *emailService) SendOrderConfirmation(to, customerName, orderID, total, items string) error {
	subject := "Подтверждение заказа #" + orderID
	body := `
		<h2>Заказ оформлен</h2>
		<p>Здравствуйте, ` + customerName + `!</p>
		<p>Ваш заказ на сумму <strong>` + total + ` ₽</strong> успешно оформлен.</p>
		<h3>Состав заказа:</h3>
		` + items + `
		<p>Мы уведомим вас, когда заказ будет отправлен.</p>
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendOrderStatusUpdate(to, customerName, orderID, status string) error {
	statusText := map[string]string{
		"paid":      "Оплачен",
		"shipped":   "Отправлен",
		"delivered": "Доставлен",
		"cancelled": "Отменён",
	}
	subject := "Статус заказа #" + orderID
	body := `
		<h2>Обновление статуса заказа</h2>
		<p>Здравствуйте, ` + customerName + `!</p>
		<p>Статус вашего заказа #` + orderID + ` изменён на: <strong>` + statusText[status] + `</strong></p>
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendSellerNewOrder(to, sellerName, orderID, total, customerEmail, customerName string) error {
	subject := "Новый заказ #" + orderID
	body := `
		<h2>Новый заказ</h2>
		<p>Здравствуйте, ` + sellerName + `!</p>
		<p>Поступил новый заказ на сумму <strong>` + total + ` ₽</strong>.</p>
		<p>Покупатель: ` + customerName + ` (` + customerEmail + `)</p>
		<p>Перейдите в панель продавца для обработки заказа.</p>
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendReceipt(to, customerName, orderID, total, items, receiptHTML string) error {
	subject := "Чек по заказу #" + orderID
	body := `
		<h2>Чек по заказу #` + orderID + `</h2>
		<p>Здравствуйте, ` + customerName + `!</p>
		` + receiptHTML + `
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendSellerApplicationApproved(to, username, storeName string) error {
	subject := "Заявка на продавца одобрена"
	body := `
		<h2>Поздравляем!</h2>
		<p>Здравствуйте, ` + username + `!</p>
		<p>Ваша заявка на регистрацию продавца для магазина "<strong>` + storeName + `</strong>" одобрена.</p>
		<p>Теперь вы можете добавлять товары и управлять своим магазином.</p>
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendSellerApplicationRejected(to, username, reason string) error {
	subject := "Заявка на продавца отклонена"
	body := `
		<h2>Заявка отклонена</h2>
		<p>Здравствуйте, ` + username + `!</p>
		<p>К сожалению, ваша заявка на регистрацию продавца была отклонена.</p>
		<p>` + reason + `</p>
	`
	return s.send(to, subject, body)
}

func (s *emailService) SendNewApplicationNotification(to, username, storeName, description string) error {
	subject := "Новая заявка на продавца"
	body := `
		<h2>Новая заявка на регистрацию продавца</h2>
		<p>Поступила новая заявка:</p>
		<p><strong>Пользователь:</strong> ` + username + `</p>
		<p><strong>Название магазина:</strong> ` + storeName + `</p>
		` + func() string {
		if description != "" {
			return `<p><strong>Описание:</strong> ` + description + `</p>`
		}
		return ""
	}() + `
		<p>Перейдите в панель администратора для рассмотрения.</p>
	`
	return s.send(to, subject, body)
}
