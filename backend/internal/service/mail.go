// Package service: business layer (mail, search index, heavy external deps).
package service

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// lightweight key for password obfuscation. Note: this is obfuscation, not encryption;
// production must use KMS / symmetric encryption (e.g. AES-GCM).
var obfsKey = []byte("pico-lite-obfs-v1")

// ObfuscatePassword XOR+base64; stores obfuscated string
func ObfuscatePassword(raw string) string {
	if raw == "" {
		return ""
	}
	xored := make([]byte, len(raw))
	for i := range []byte(raw) {
		xored[i] = raw[i] ^ obfsKey[i%len(obfsKey)]
	}
	return "x:" + base64.StdEncoding.EncodeToString(xored)
}

// DeobfuscatePassword reverses obfuscation
func DeobfuscatePassword(stored string) string {
	if !strings.HasPrefix(stored, "x:") {
		return stored // legacy/unobfuscated data returned as-is
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, "x:"))
	if err != nil {
		return stored
	}
	out := make([]byte, len(raw))
	for i := range raw {
		out[i] = raw[i] ^ obfsKey[i%len(obfsKey)]
	}
	return string(out)
}

func addrOf(a *imap.Address) string {
	if a == nil {
		return ""
	}
	mbox, host := a.Mailbox, a.Host
	if mbox == "" || host == "" {
		return ""
	}
	return mbox + "@" + host
}

// SyncInbox fetches latest 50 INBOX via IMAP, upsert by (account_id,uid).
// returns rows written.
func SyncInbox(db *gorm.DB, acc *model.MailAccount) (int, error) {
	pass := DeobfuscatePassword(acc.Password)
	addr := net.JoinHostPort(acc.ImapHost, strconv.Itoa(acc.ImapPort))

	var cl *imapclient.Client
	var err error
	if acc.UseTLS {
		cl, err = imapclient.DialTLS(addr, nil)
	} else {
		cl, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return 0, fmt.Errorf("imap dial %s: %w", addr, err)
	}
	defer cl.Logout().Wait()

	if err := cl.Login(acc.Username, pass).Wait(); err != nil {
		return 0, fmt.Errorf("imap login: %w", err)
	}
	sel, err := cl.Select("INBOX", nil).Wait()
	if err != nil {
		return 0, fmt.Errorf("imap select inbox: %w", err)
	}
	if sel.NumMessages == 0 {
		return 0, nil
	}
	// latest 50: sequence range [n-49, n]
	from := uint32(1)
	if sel.NumMessages > 50 {
		from = sel.NumMessages - 49
	}
	seqSet := imap.SeqSet{}
	seqSet.AddRange(from, sel.NumMessages)

	opts := &imap.FetchOptions{
		UID:          true,
		Envelope:     true,
		InternalDate: true,
		BodySection:  []*imap.FetchItemBodySection{{}},
	}
	cmd := cl.Fetch(&seqSet, opts)
	defer cmd.Close()

	added := 0
	for {
		msg := cmd.Next()
		if msg == nil {
			break
		}
		buf, err := msg.Collect()
		if err != nil {
			break
		}
		uidStr := strconv.FormatUint(uint64(buf.UID), 10)
		from, to, subject := "", "", ""
		var recvAt time.Time
		if buf.Envelope != nil {
			if len(buf.Envelope.From) > 0 {
				from = addrOf(&buf.Envelope.From[0])
			}
			if len(buf.Envelope.To) > 0 {
				to = addrOf(&buf.Envelope.To[0])
			}
			subject = buf.Envelope.Subject
			recvAt = buf.Envelope.Date
		}
		if recvAt.IsZero() {
			recvAt = buf.InternalDate
		}
		body := ""
		if len(buf.BodySection) > 0 {
			body = string(buf.BodySection[0].Bytes)
		}

		// dedup upsert by (account,uid)
		var cnt int64
		db.Model(&model.MailMessage{}).Where("account_id = ? AND uid = ?", acc.ID, uidStr).Count(&cnt)
		if cnt > 0 {
			continue
		}
		m := model.MailMessage{
			AccountID:  acc.ID,
			UID:        uidStr,
			From:       from,
			To:         to,
			Subject:    subject,
			Body:       body,
			Folder:     "INBOX",
			ReceivedAt: recvAt,
		}
		if err := db.Create(&m).Error; err == nil {
			added++
		}
	}
	return added, nil
}

// SendMail sends a plain-text email via SMTP
func SendMail(acc *model.MailAccount, to, subject, body string) error {
	pass := DeobfuscatePassword(acc.Password)
	if to == "" {
		return fmt.Errorf("to required")
	}
	addr := net.JoinHostPort(acc.SmtpHost, strconv.Itoa(acc.SmtpPort))

	headers := []string{
		"From: " + acc.Email,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	auth := smtp.PlainAuth("", acc.Username, pass, acc.SmtpHost)

	// UseTLS = 465 implicit TLS; otherwise standard SendMail (auto STARTTLS)
	if acc.UseTLS {
		tlsCfg := &tls.Config{ServerName: acc.SmtpHost}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("smtp tsl dial %s: %w", addr, err)
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, acc.SmtpHost)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(acc.Email); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt to: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("smtp write: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("smtp close data: %w", err)
		}
		return client.Quit()
	}

	if err := smtp.SendMail(addr, auth, acc.Email, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
