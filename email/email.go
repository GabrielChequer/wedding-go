package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

// Config holds SMTP credentials. Populate from environment variables.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Service sends transactional emails via SMTP.
type Service struct {
	cfg Config
}

func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

type templateData struct {
	Attending bool
}

var confirmationTmpl = template.Must(template.New("rsvp-confirmation").Parse(confirmationHTML))

// SendRsvpConfirmation sends a styled HTML confirmation to toAddr.
// Silently skips if SMTP credentials are not configured.
func (s *Service) SendRsvpConfirmation(toAddr string, attending bool) error {
	if s.cfg.Username == "" || s.cfg.Password == "" {
		return nil
	}

	var body bytes.Buffer
	if err := confirmationTmpl.Execute(&body, templateData{Attending: attending}); err != nil {
		return fmt.Errorf("rendering confirmation template: %w", err)
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: RSVP Confirmed — Kelsie & Gabriel\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n",
		s.cfg.From,
		toAddr,
	)

	msg := append([]byte(headers), body.Bytes()...)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	return smtp.SendMail(s.cfg.Host+":"+s.cfg.Port, auth, s.cfg.From, []string{toAddr}, msg)
}

const confirmationHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>RSVP Confirmed — Kelsie &amp; Gabriel</title>
</head>
<body style="margin:0;padding:0;background-color:#f9f5ef;font-family:Arial,Helvetica,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9f5ef;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;background-color:#fdfaf6;border:1px solid #c5d3c6;">

          <!-- Header -->
          <tr>
            <td style="background-color:#3b4834;padding:40px 48px;text-align:center;">
              <p style="margin:0 0 14px;color:#b89b6a;font-size:10px;letter-spacing:0.3em;text-transform:uppercase;font-family:Arial,sans-serif;">You&#39;re Invited</p>
              <h1 style="margin:0;color:#f9f5ef;font-family:Georgia,'Times New Roman',serif;font-size:44px;font-weight:300;letter-spacing:0.04em;">
                K <span style="color:#b89b6a;font-style:italic;">&amp;</span> G
              </h1>
              <p style="margin:14px 0 0;color:#c5d3c6;font-size:11px;letter-spacing:0.22em;text-transform:uppercase;font-family:Arial,sans-serif;">Kelsie &amp; Gabriel</p>
            </td>
          </tr>

          <!-- Gold top accent line -->
          <tr>
            <td style="padding:0;height:2px;background-color:#b89b6a;line-height:2px;font-size:2px;">&nbsp;</td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:48px 48px 40px;">

              <p style="margin:0 0 6px;color:#b89b6a;font-size:10px;letter-spacing:0.28em;text-transform:uppercase;font-family:Arial,sans-serif;">RSVP Confirmation</p>
              <h2 style="margin:0 0 20px;color:#2c2c2c;font-family:Georgia,'Times New Roman',serif;font-size:28px;font-weight:300;line-height:1.3;">
                We received your response
              </h2>
              <!-- Accent divider -->
              <table role="presentation" width="60" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
                <tr>
                  <td style="height:1px;background-color:#b89b6a;line-height:1px;font-size:1px;">&nbsp;</td>
                </tr>
              </table>

              <p style="color:#2c2c2c;font-size:15px;line-height:1.75;margin:0 0 20px;">
                Thank you for responding to our wedding invitation. We have received your RSVP and are so grateful you took the time to let us know.
              </p>

              {{if .Attending}}
              <p style="color:#2c2c2c;font-size:15px;line-height:1.75;margin:0 0 36px;">
                We are <em>so</em> excited to celebrate this day with you. See you there!
              </p>
              {{else}}
              <p style="color:#2c2c2c;font-size:15px;line-height:1.75;margin:0 0 36px;">
                We&#39;re sorry you won&#39;t be able to join us, but we truly appreciate you letting us know. You will be missed!
              </p>
              {{end}}

              <!-- Event details card -->
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9f5ef;border:1px solid #c5d3c6;margin-bottom:40px;">
                <tr>
                  <td style="padding:28px 32px;">
                    <p style="margin:0 0 4px;color:#b89b6a;font-size:10px;letter-spacing:0.28em;text-transform:uppercase;font-family:Arial,sans-serif;">The Celebration</p>
                    <h3 style="margin:0 0 20px;color:#2c2c2c;font-family:Georgia,'Times New Roman',serif;font-size:20px;font-weight:300;">Kelsie &amp; Gabriel</h3>
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                      <tr>
                        <td style="padding:10px 0;border-bottom:1px solid #c5d3c6;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:Arial,sans-serif;">Date</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:Arial,sans-serif;">September 25, 2027</span>
                        </td>
                      </tr>
                      <tr>
                        <td style="padding:10px 0;border-bottom:1px solid #c5d3c6;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:Arial,sans-serif;">Venue</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:Arial,sans-serif;">The Magnolia</span>
                        </td>
                      </tr>
                      <tr>
                        <td style="padding:10px 0;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:Arial,sans-serif;">Location</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:Arial,sans-serif;">Clarksville, Indiana</span>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>

              <p style="color:#8b7355;font-size:14px;line-height:1.75;margin:0;">
                With love,<br>
                <span style="font-family:Georgia,serif;font-size:17px;color:#2c2c2c;">Kelsie &amp; Gabriel</span>
              </p>
            </td>
          </tr>

          <!-- Bottom accent line -->
          <tr>
            <td style="padding:0 48px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="height:1px;background-color:#c5d3c6;line-height:1px;font-size:1px;">&nbsp;</td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:24px 48px;text-align:center;">
              <p style="margin:0;color:#8b7355;font-size:11px;letter-spacing:0.1em;font-family:Arial,sans-serif;">
                September 25, 2027&nbsp;&nbsp;&middot;&nbsp;&nbsp;The Magnolia&nbsp;&nbsp;&middot;&nbsp;&nbsp;Clarksville, Indiana
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
