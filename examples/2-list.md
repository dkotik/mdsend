---
from: Test Author <joedoe@test.com>
to: to@test.com
cc:
  - ../address/testdata/recipients.yaml
  - ../address/testdata/recipients.csv
  - ../address/testdata/recipients.cue
  - ../address/testdata/recipients.json
  - ../address/testdata/recipients.toml
language: en
bcc: bcc1@test.com
subject: typical mailing list message
headers:
  List-ID: Some List <greatlist@test.com>
  List-Unsubscribe: <mailto:unsub@yourdomain.com>, <{{ reify "unsubscribe_url" }}>
  List-Unsubscribe-Post: token={{ reify "unsubscribe_token" }}
---

{{- define "unsubscribe_token" -}}
  {{ base58 (print .Recipient.email "?list=testList") }}
{{- end -}}

{{- define "unsubscribe_url" -}}
  https://yourdomain.com/unsub?id={{ reify "unsubscribe_token" }}
{{- end -}}

# Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce commodo sapien sed magna eleifend, a aliquam nunc condimentum. Sed feugiat nibh at lorem malesuada, lobortis tincidunt urna bibendum. Maecenas pulvinar, quam id tincidunt auctor, sem est tincidunt nulla, ac fringilla velit sapien eget nisl.

## Footer

You may unsubscribe <a title="unsubscribe from the mailing list" href="{{ reify "unsubscribe_url" }}">here</a>.
