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
  List-Unsubscribe: <mailto:unsub@yourdomain.com>, <https://yourdomain.com/unsub?id={{ base58 .Recipient.email }}>
  List-Unsubscribe-Post: token={{ base58 .Recipient.email }}
---

# Content

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce commodo sapien sed magna eleifend, a aliquam nunc condimentum. Sed feugiat nibh at lorem malesuada, lobortis tincidunt urna bibendum. Maecenas pulvinar, quam id tincidunt auctor, sem est tincidunt nulla, ac fringilla velit sapien eget nisl.

Integer fringilla ipsum in elit tempus varius. Phasellus viverra velit in pulvinar bibendum. Pellentesque non varius est, in faucibus felis. Praesent sem tellus, consequat id sem nec, tempor efficitur mi. Proin et sem condimentum, dictum dolor vitae, blandit mi. Vestibulum eget diam ante. Etiam facilisis mi non augue malesuada placerat. Maecenas porttitor nulla a sem mattis eleifend. Fusce at neque augue.

Integer at odio laoreet orci iaculis hendrerit. Curabitur sollicitudin volutpat dui at sodales. Vestibulum elit magna, semper nec dolor eget, pharetra tempus augue.

Suspendisse iaculis tortor at massa congue condimentum. Ut interdum leo vel dignissim feugiat. Pellentesque volutpat molestie interdum. Vivamus vel purus sem. In eu hendrerit purus, vel gravida libero.
