package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
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
	FirstName string
	LastName  string
}

var confirmationTmpl = template.Must(template.New("rsvp-confirmation").Parse(confirmationHTML))

// SendRsvpConfirmation sends a styled HTML confirmation to toAddr.
// Silently skips if SMTP credentials are not configured.
func (s *Service) SendRsvpConfirmation(toAddr string, firstName string, lastName string) error {
	log.Printf("Sending rsvp-confirmation to %s %s@%s", firstName, lastName, toAddr)
	if s.cfg.Username == "" || s.cfg.Password == "" {
		return nil
	}

	var body bytes.Buffer
	if err := confirmationTmpl.Execute(&body, templateData{FirstName: firstName, LastName: lastName}); err != nil {
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
  <link
      href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,400;0,500;1,300;1,400&family=Jost:wght@200;300;400;500&display=swap"
      rel="stylesheet"
    />
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>RSVP Confirmed — Kelsie &amp; Gabriel</title>
</head>
<body style="margin:0;padding:0;background-color:#f9f5ef;font-family:'Jost',sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9f5ef;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;background-color:#fdfaf6;border:1px solid #c5d3c6;">

          <!-- Header -->
          <tr>
            <td style="background-color:#3b4834;padding:40px 48px;text-align:center;">
              <h1 style="margin:0;color:#f9f5ef;font-family:'Cormorant Garamond', serif;font-size:44px;font-weight:300;letter-spacing:0.04em;">
                K <span style="color:#b89b6a;font-style: italic;">&amp;</span> G
              </h1>
            </td>
          </tr>

          <!-- Gold top accent line -->
          <tr>
            <td style="padding:0;height:2px;background-color:#b89b6a;line-height:2px;font-size:2px;">&nbsp;</td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:48px 40px 40px;">

              <p style="margin:0 0 6px;color:#b89b6a;font-size:10px;letter-spacing:0.28em;text-transform:uppercase;font-family:'Cormorant Garamond',serif;">RSVP Confirmation</p>
              <h2 style="margin:0 0 20px;color:#2c2c2c;font-family:'Cormorant Garamond',serif;font-size:28px;font-weight:300;line-height:1.3;">
                Hey {{.firstName}}, we received your response
              </h2>
              <!-- Accent divider -->
              <table role="presentation" width="60" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
                <tr>
                  <td style="height:1px;background-color:#b89b6a;line-height:1px;font-size:1px;">&nbsp;</td>
                </tr>
              </table>

              <p style="color:#2c2c2c;font-size:15px;line-height:1.75;margin:0 0 20px;">
                Thank you for responding to our wedding invitation. We have received your RSVP and we look forward to seeing you there!
              </p>

              <p style="color:#2c2c2c;font-size:15px;line-height:1.75;margin:0 0 36px;">
                We will reach out to you if there are any updates. In the meantime, don't hesitate to contact us with any questions at 
                <a href="mailto:kelsie.renfrow347@gmail.com" style="color:#b89b6a;text-decoration:underline;">Kelsie's email</a>
                or 
                <a href="mailto:chequeros1@gmail.com" style="color:#b89b6a;text-decoration:underline;">Gabriel's email</a>.
              </p>

              <!-- Event details card -->
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f9f5ef;border:1px solid #c5d3c6;margin-bottom:40px;">
                <tr>
                  <td style="padding:28px 32px;">
                    <p style="margin:0 0 4px;color:#b89b6a;font-size:10px;letter-spacing:0.28em;text-transform:uppercase;font-family:'Cormorant Garamond',serif;">The Celebration</p>
                    <h3 style="margin:0 0 20px;color:#2c2c2c;font-family:'Cormorant Garamond',serif;font-size:20px;font-weight:300;">Kelsie &amp; Gabriel</h3>
                    <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                      <tr>
                        <td style="padding:10px 0;border-bottom:1px solid #c5d3c6;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:'Cormorant Garamond',serif;">Date</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:'Jost',sans-serif;">September 25, 2027</span>
                        </td>
                      </tr>
                      <tr>
                        <td style="padding:10px 0;border-bottom:1px solid #c5d3c6;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:'Cormorant Garamond',serif;">Venue</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:'Jost',sans-serif;">The Magnolia</span>
                        </td>
                      </tr>
                      <tr>
                        <td style="padding:10px 0;">
                          <span style="color:#8a9e8c;font-size:10px;letter-spacing:0.15em;text-transform:uppercase;display:inline-block;width:80px;font-family:'Cormorant Garamond',serif;">Location</span>
                          <span style="color:#2c2c2c;font-size:14px;font-family:'Jost',sans-serif;">Clarksville, Indiana</span>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>

              <p style="color:#8b7355;font-family:'Cormorant Garamond',serif;font-size:14px;line-height:1.75;margin:0;">
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
