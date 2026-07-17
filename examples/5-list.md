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

Carbon copy recipients (`cc`) are treated exactly the same as blind
carbon copy recipients. The field exists for historic reasons. Those
recipients will not be displayed in sent messages by design for
the following reasons:

- To prevent accidental exposure of private contact information.
- To prevent possible message duplication in mailer APIs due to
  implementation ambiguity in handling carbon copy recipients.

## Load Recipient List From an Executable

If the recipient file does not have a recognizable extension,
like `.yaml`, and is executable, it will be executed. The standard
output will be captured and parsed according to the `Content-Type`
set in the first line of the output.

```yaml
to: script.sh
bcc: program.exe
```

The first line of output must match the desired file format.
This protects `mdsend` from malformed content of failing scripts
and binary files, and precludes ambiguity in choosing between JSON and
Cuelang, CSV and other formats.

- *JSON:* Content-Type: application/json
- *Cue:* Content-Type: application/cue
- *YAML:* Content-Type: application/yaml
- *TOML:* Content-Type: application/toml
- *CSV:* Content-Type: text/csv

```sh
#!/bin/sh

# example of an executable script returning JSON data.
cat << EOF
Content-type: application/json

[
  {
    "name": "first",
    "email": "first@testmail.json"
  },
  {
    "name": "second",
    "email": "second@testmail.json"
  },
  {
    "name": "third",
    "email": "third@testmail.json"
  },
  {
    "name": "fourth",
    "email": "fourth@testmail.json"
  }
]
EOF
```

Sourcing recipients from an executable gives you the limitless
flexibility to fetch private contact information from a remote
database or encrypted location, to groom or combine and pre-process the
contact list.

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
