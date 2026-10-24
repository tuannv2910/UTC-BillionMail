package mail_service

import (
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/frame/g"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Title   string
	Content string
	Headers map[string]string
}

// MailTitle get title of email
func (m Message) MailTitle() string {
	return mime.QEncoding.Encode("UTF-8", m.Title)
}

// MailText get content of email
func (m Message) MailText() string {
	buf := new(bytes.Buffer)
	wt := quotedprintable.NewWriter(buf)
	wt.Write([]byte(m.Content))
	wt.Close()
	return buf.String()
}

// MailHeader get headers of email
func (m Message) MailHeader() string {
	headers := ""
	for key, value := range m.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return headers
}

// MessageID get Message-ID of email
func (m Message) MessageID() string {
	if id, ok := m.Headers["Message-ID"]; ok {
		return id
	}
	return ""
}

// SetMessageID set Message-ID of email
func (m *Message) SetMessageID(id string) {
	m.SetHeader("Message-ID", id)
}

// SetHeader set header of email
func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

// SetHeaders set headers of email
func (m *Message) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		m.SetHeader(key, value)
	}
}

// SetRealName set the sender display name. It is encoded when the From header
// is built, so names with commas, quotes or non-ASCII characters stay intact (#376).
func (m *Message) SetRealName(realName string) {
	m.SetHeader("RealName", strings.TrimSpace(realName))
}

// FormatFromHeader builds an RFC 5322 From value such as
// "Vua Nem <vuanem@example.com>", quoting/encoding the name when needed.
func FormatFromHeader(displayName, email string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	return (&mail.Address{Name: displayName, Address: email}).String()
}

var (
	mailerHeaderOnce  sync.Once
	mailerHeaderValue = "BillionMail"
)

// MailerHeader returns the X-Mailer value including the running version (#215).
func MailerHeader() string {
	mailerHeaderOnce.Do(func() {
		defer func() { _ = recover() }()
		if v := strings.TrimSpace(g.Cfg().MustGet(context.Background(), "server.version", "").String()); v != "" {
			mailerHeaderValue = "BillionMail " + strings.TrimPrefix(v, "v")
		}
	})
	return mailerHeaderValue
}

func NewMessage(title string, content string) Message {
	return Message{
		Title:   title,
		Content: content,
		Headers: make(map[string]string),
	}
}

type EmailSender struct {
	Email     string       `json:"email" v:"required|email"` // email of sender
	UserName  string       `json:"username" v:"required"`
	Host      string       `json:"host" v:"required"`     // SMTP server address
	Port      string       `json:"port" v:"required"`     // SMTP server port
	Password  string       `json:"password" v:"required"` // SMTP password
	SNI       string       `json:"sni"`                   // SNI for TLS
	client    *smtp.Client // persistent SMTP client connection
	conn      net.Conn     // underlying connection, used for I/O deadlines
	mutex     sync.Mutex   // mutex for thread safety
	connected bool         // connection status
}

type customAuth struct {
	Username, Password string
}

func (a *customAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.Username + "\x00" + a.Password), nil
}

func (a *customAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server response")
	}
	return nil, nil
}

func NewEmailSender() *EmailSender {
	e := &EmailSender{
		Host:      "localhost",
		Port:      "25",
		connected: false,
	}

	if public.IsRunningInContainer() {
		e.Host = "postfix"
		// e.Port = "587"
		// e.SNI, _ = public.DockerEnv("BILLIONMAIL_HOSTNAME")
	}

	return e
}

func NewEmailSenderWithLocal(email string) (es *EmailSender, err error) {
	es = NewEmailSender()

	es.Email = email
	es.UserName = email
	es.Password, err = PasswordPlainByEmail(context.Background(), email)

	if err != nil {
		return
	}

	if !es.IsConfigured() {
		err = errors.New("Email Sender not configured")
		return
	}

	return
}

// PasswordPlainByEmail
func PasswordPlainByEmail(ctx context.Context, email string) (string, error) {
	val, err := g.DB().Model("mailbox").Where("username", email).Value("password_encode")
	if err != nil {
		return "", fmt.Errorf("query password failed: %w", err)
	}
	if val.IsEmpty() {
		return "", fmt.Errorf("password not found for %s", email)
	}
	rawHex, err := hex.DecodeString(val.String())
	if err != nil {
		return "", fmt.Errorf("hex decode failed: %w", err)
	}
	plainBytes, err := base64.StdEncoding.DecodeString(string(rawHex))
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}
	return string(plainBytes), nil
}

// Close closes the SMTP connection
func (e *EmailSender) Close() {
	_ = e.Disconnect()
}

// Connect establishes an SMTP connection
func (e *EmailSender) Connect() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.connected && e.client != nil {
		// Connection already exists
		return nil
	}

	var err error

	if e.isSecure() {
		err = e.connectWithSSL()
	} else {
		err = e.connectPlain()
	}

	if err != nil {
		e.connected = false
		e.client = nil
		return err
	}

	e.connected = true
	return nil
}

// smtpDialTimeout bounds TCP connect; smtpHandshakeTimeout bounds the TLS
// handshake, greeting, STARTTLS and AUTH so a misconfigured port surfaces as an
// error instead of hanging for minutes (#379).
var (
	smtpDialTimeout      = 15 * time.Second
	smtpHandshakeTimeout = 30 * time.Second
	smtpSendTimeout      = 5 * time.Minute
)

// implicitTLSPorts are ports that speak TLS from the first byte (SMTPS).
var implicitTLSPorts = map[string]bool{
	"465":  true,
	"2465": true,
	"8465": true,
}

// IsImplicitTLSPort reports whether a port is conventionally used for SMTPS.
func IsImplicitTLSPort(port string) bool {
	return implicitTLSPorts[strings.TrimSpace(port)]
}

func (e *EmailSender) tlsConfig() *tls.Config {
	serverName := e.SNI
	if serverName == "" {
		serverName = e.Host
	}
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		ServerName:         serverName,
	}
}

// connectWithSSL establishes a secure SMTP connection
func (e *EmailSender) connectWithSSL() error {
	dialer := &net.Dialer{Timeout: smtpDialTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(e.Host, e.Port), e.tlsConfig())
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}

	_ = conn.SetDeadline(time.Now().Add(smtpHandshakeTimeout))

	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("new SMTP client: %w", err)
	}

	if err = e.authenticate(client, smtp.PlainAuth("", e.UserName, e.Password, e.Host)); err != nil {
		client.Close()
		return err
	}

	// Clear the handshake deadline; sending sets its own per message.
	_ = conn.SetDeadline(time.Time{})
	e.conn = conn
	e.client = client

	return nil
}

// connectPlain establishes a plain SMTP connection, upgrading with STARTTLS
// whenever the server offers it (not only on port 587).
func (e *EmailSender) connectPlain() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(e.Host, e.Port), smtpDialTimeout)
	if err != nil {
		return fmt.Errorf("SMTP dial: %w", err)
	}

	_ = conn.SetDeadline(time.Now().Add(smtpHandshakeTimeout))

	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		conn.Close()
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return fmt.Errorf("SMTP dial: no greeting from %s within %s (the port may require implicit TLS/SSL): %w",
				net.JoinHostPort(e.Host, e.Port), smtpHandshakeTimeout, err)
		}
		return fmt.Errorf("SMTP dial: %w", err)
	}

	// Port 25 to the local Postfix keeps the historical plaintext behaviour.
	if e.Port != "25" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(e.tlsConfig()); err != nil {
				client.Close()
				return fmt.Errorf("SMTP STARTTLS: %w", err)
			}
		}
	}

	var auth smtp.Auth

	if e.Port == "25" {
		auth = &customAuth{e.UserName, e.Password}
	} else {
		auth = smtp.PlainAuth("", e.UserName, e.Password, e.Host)
	}

	if err = e.authenticate(client, auth); err != nil {
		client.Close()
		return err
	}

	_ = conn.SetDeadline(time.Time{})
	e.conn = conn
	e.client = client
	return nil
}

// authenticate runs SMTP AUTH unless no credentials were configured
// (IP-authenticated relays, #55).
func (e *EmailSender) authenticate(client *smtp.Client, auth smtp.Auth) error {
	if e.UserName == "" && e.Password == "" {
		return nil
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}
	return nil
}

// Disconnect closes the SMTP connection
func (e *EmailSender) Disconnect() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.connected || e.client == nil {
		return nil
	}

	err := e.client.Quit()
	if err != nil {
		// Force close connection if Quit fails
		_ = e.client.Close()
		g.Log().Warning(context.Background(), "Error closing SMTP connection: ", err)
	}

	e.connected = false
	e.client = nil
	e.conn = nil
	return err
}

// GenerateMessageID generates a unique Message-ID for email
func (e *EmailSender) GenerateMessageID() string {
	randomBytes := grand.B(16)
	randomID := hex.EncodeToString(randomBytes)
	timestampMillis := time.Now().UnixMilli()

	domain := strings.SplitN(e.Email, "@", 2)
	domainPart := "billionmail"
	if len(domain) > 1 {
		domainPart = domain[1]
	}

	return fmt.Sprintf("<%d.%s@%s>", timestampMillis, randomID, domainPart)
}

// Send sends an email to specified recipients
func (e *EmailSender) Send(message Message, recipients []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Make sure we have a connection
	if !e.connected || e.client == nil {
		e.mutex.Unlock()
		if err := e.Connect(); err != nil {
			e.mutex.Lock()
			return fmt.Errorf("failed to connect: %w", err)
		}
		e.mutex.Lock()
	}

	// Try to send the message, with reconnect on failure
	err := e.doSend(message, recipients)
	if err != nil {
		// Connection might be stale, try to reconnect once
		g.Log().Debug(context.Background(), "SMTP send failed, attempting reconnection")

		// Clean up
		e.client.Close()
		e.connected = false
		e.client = nil
		e.conn = nil

		// Try to reconnect
		e.mutex.Unlock()
		if err := e.Connect(); err != nil {
			e.mutex.Lock()
			return fmt.Errorf("reconnect failed: %w", err)
		}
		e.mutex.Lock()

		// Try sending again after reconnect
		return e.doSend(message, recipients)
	}

	return nil
}

// doSend performs the actual message sending
func (e *EmailSender) doSend(message Message, recipients []string) error {
	// Work on a copy of the headers: they are modified below (RealName/From are
	// removed, Content-Type may become multipart). Mutating the caller's map
	// made the retry after a reconnect lose the display name (#376).
	headers := make(map[string]string, len(message.Headers)+4)
	for k, v := range message.Headers {
		headers[k] = v
	}
	message.Headers = headers

	// Bound each message transaction so a stalled server cannot block forever
	if e.conn != nil {
		_ = e.conn.SetDeadline(time.Now().Add(smtpSendTimeout))
		defer func() {
			if e.conn != nil {
				_ = e.conn.SetDeadline(time.Time{})
			}
		}()
	}

	// Add Message-ID if not already set
	if _, exists := message.Headers["Message-ID"]; !exists {
		message.Headers["Message-ID"] = e.GenerateMessageID()
	}

	// Add Date if not already set
	if _, exists := message.Headers["Date"]; !exists {
		message.Headers["Date"] = time.Now().Format(time.RFC1123Z)
	}

	// Default Content-Type if not already set
	if _, exists := message.Headers["Content-Type"]; !exists {
		message.Headers["Content-Type"] = "text/html; charset=utf-8"
	}

	// Default From header: display name (or local part) + address
	from := FormatFromHeader(message.Headers["RealName"], e.Email)
	delete(message.Headers, "RealName")

	if v, exists := message.Headers["From"]; exists {
		from = v
		delete(message.Headers, "From")
	}

	// HTML bodies go out as multipart/alternative with a text/plain part (#331);
	// other content types keep the single-part quoted-printable body.
	transferEncoding := "Content-Transfer-Encoding: quoted-printable\r\n"
	body := message.MailText()
	if isHTMLContentType(message.Headers["Content-Type"]) && strings.TrimSpace(message.Content) != "" {
		message.Headers["Content-Type"], body = buildAlternativeBody(message.Content)
		transferEncoding = ""
	}

	// Build email message with headers
	headerString := fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ",")) +
		fmt.Sprintf("Subject: %s\r\n", message.MailTitle()) +
		"MIME-Version: 1.0\r\n" +
		transferEncoding +
		fmt.Sprintf("X-Mailer: %s\r\n", MailerHeader()) +
		message.MailHeader() +
		"\r\n" +
		body +
		"\r\n"

	msg := []byte(headerString)

	var err error

	defer func() {
		// Reset the connection state if sending fails
		if err != nil {
			if err = e.client.Reset(); err != nil {
				g.Log().Warning(context.Background(), "SMTP reset: %w", err)
			}
		}
	}()

	// Set the sender
	// MAIL FROM
	if err = e.client.Mail(e.Email); err != nil {
		return fmt.Errorf("SMTP mail: %w", err)
	}

	// Set the recipients
	// RCPT TO
	for _, to := range recipients {
		if err = e.client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP rcpt: %w", err)
		}
	}

	// Get a writer for the message body
	var w io.WriteCloser
	w, err = e.client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data: %w", err)
	}

	// Write the message
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write: %w", err)
	}

	// Close the writer
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP close writer: %w", err)
	}

	return nil
}

// isSecure check if TLS connection is needed
func (e *EmailSender) isSecure() bool {
	// Implicit TLS (SMTPS) ports, e.g. 465 and the common alternative 2465.
	// Other ports (587, 2525, ...) connect in plaintext and upgrade with
	// STARTTLS when the server advertises it.
	return IsImplicitTLSPort(e.Port)
}

// IsConfigured check if Email notification is configured
func (e *EmailSender) IsConfigured() bool {
	e.Email = strings.TrimSpace(e.Email)
	e.Host = strings.TrimSpace(e.Host)
	e.Port = strings.TrimSpace(e.Port)
	e.UserName = strings.TrimSpace(e.UserName)
	e.Password = strings.TrimSpace(e.Password)

	if e.UserName == "" || e.Host == "" || e.Port == "" || e.Password == "" || e.Email == "" {
		return false
	}

	if e.Host != "postfix" && !public.IsHost(e.Host) {
		g.Log().Warning(context.Background(), "Host is not a valid host: ", e.Host)
		return false
	}

	if err := g.Validator().Data(e).Run(context.Background()); err != nil {
		g.Log().Warning(context.Background(), "Email sender validation failed: ", err)
		return false
	}

	return true
}
