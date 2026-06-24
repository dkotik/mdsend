# Mdsend

Send markdown files as mail.

**Project status: active development towads beta.**

## Features

- **Durable:** mail queues are fault tolerant and atomic, brokered by <Watermill.io> over SQLite3.
- **Flexible:** select the mailer backend based on highest deliverability. Swap it out later without changing anything in your letters or templates.
  - Supports recipient lists formats: CSV, JSON, YAML, TOML, and Cue.
- **Portable:** runs on many kinds of systems. Will include an embeddable HTTP service and Posgres support in the future.

## Installation

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```

## Letter

```markdown
---
subject: "Test Email"
from: "Test Account <test@gmail.com>"
to: "Test Account <test@gmail.com>"
---

# Title of the Example Letter

Write text in **Markdown** notation.
```

## Load Recipient Lists

Target address fields, `to`, `cc`, and `bcc` support a list of entries. Any entry can also point to a configuration file in a variety of common formats that will be merged into the parent list.

```yaml
to: mailinglist.yaml
cc:
  - another_list.toml
  - jsonIsFine.json
  - cuelist.cue
bcc:
  - name: Named Entry
    email: test@test.com
    title: Mr.
    first_name: First
```

Each entry is a map, regardless of how it was loaded. Any value of the map is accessible to the template engine through the context.

```markdown
Greeting {{ .Recipient.title }} {{ .Recipient.first_name }},

I am writing, because ...
```

## Extend One Letter with Another

Merge the content and frontmatter fields into a letter from another file by using the `extends` field.

```yaml
extends:
  # this letter will inherit the content
  # and configuration values of `template.md`
  - template.md
  # load a configuration file
  - config.yaml
```

Any Markdown content below the last horizontal rule is added as a footer to the current letter. You may nest the extensions deep as long as there are no circular references.

## Development Roadmap

- [x] Anticipate circular imports for recipient lists and extensions.
- [ ] Add queue.Marshaler interface and a JSON implementation?
- [ ] Write a better markdown renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Deprecate test package at root.
- [ ] https://github.com/charmbracelet/glamour
- [ ] Beautify the default template. Add `dark.html` one.
- [ ] Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
- [x] Mailgun
- [ ] Resend
- [ ] Loops
- [ ] https://purelymail.com/
- [ ] SendGrid
- [ ] Amazon SES
- [ ] Postmark
- [ ] Sparkpost
- [ ] Brevo
- [ ] SendGrid
- [ ] Mailchimp
- [ ] HubSpot
- [ ] Twilio
- [ ] <https://emaillabs.io/en/product/>
- [ ] <https://mailtrap.io/>
- [ ] add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps

## Tools

* OBEY THE LAW. The CAN-SPAM act became law on Jan. 1, 2004. It says there many things you must do as a commercial email-er. Highlights are basically don't be deceptive, and that you MUST include a physical mailing address as well as a working unsubscribe link.
* unsubscribe button: https://blog.leavemealone.app/how-does-the-gmail-unsubscribe-button-work/

- <https://www.caniemail.com/> - check what is template-appropriate
- <https://www.htmlemailcheck.com/check/>
- Use <https://www.mail-tester.com/> to check the deliverability of your mail.
- Address verification:
  - <https://github.com/AfterShip/email-verifier>
  - <https://github.com/reacherhq/check-if-email-exists>
  - <https://hunter.io>
  - <https://verify-email.org>
  - <https://email-checker.net>
  - <https://github.com/mailcheck/mailcheck>
  - <https://github.com/ivolo/disposable-email-domains>
  - <https://github.com/willwhite/freemail>
- Hosting:
  - <https://github.com/modoboa/modoboa>

## Similar Projects

- <https://github.com/domodwyer/mailyak>
- <https://sendune.com/>
- <https://github.com/charmbracelet/pop>
- <https://github.com/wneessen/go-mail>
