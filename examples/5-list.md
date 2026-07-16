---
subject: typical mailing list message
from: Test Author <joedoe@test.com>
to: to@test.com
bcc:
  - ../address/testdata/recipients.yaml
  - ../address/testdata/recipients.csv
  - ../address/testdata/recipients.cue
  - ../address/testdata/recipients.json
  - ../address/testdata/recipients.toml
language: en
headers:
  List-ID: Some List <greatlist@test.com>
  List-Unsubscribe: <mailto:unsub@yourdomain.com>, <{{ reify "unsubscribe_url" }}>
  List-Unsubscribe-Post: token={{ reify "unsubscribe_token" }}
---

# Load Recipient Lists

Target address fields, `to`, `cc`, and `bcc` support a list of
entries. Any entry can also point to a configuration file in a
variety of common formats that will be merged into the parent list.

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

Each entry is a map, regardless of how it was loaded. Any value of
the map is accessible to the template engine through the context.

```markdown
Greeting {{ .Recipient.title }} {{ .Recipient.first_name }},

I am writing, because ...
```

## Footer

You may unsubscribe <a title="unsubscribe from the mailing list"
href="{{ reify "unsubscribe_url" }}">here</a>.

The CAN-SPAM act became law on Jan. 1, 2004. It says there many
things you must do as a commercial email-er. Highlights are
basically don't be deceptive, and that you **MUST** include a
physical mailing address as well as a working unsubscribe link in
the body.

{{- define "unsubscribe_token" -}}
  {{ base58 (print .Recipient.email "?list=testList") }}
{{- end -}}

{{- define "unsubscribe_url" -}}
  https://yourdomain.com/unsub?id={{ reify "unsubscribe_token" }}
{{- end -}}
