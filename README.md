```
                   __                    _
   _________ ___  / /_____        ____  (_)___ ____  ____  ____
  / ___/ __ `__ \/ __/ __ \______/ __ \/ / __ `/ _ \/ __ \/ __ \
 (__  ) / / / / / /_/ /_/ /_____/ /_/ / / /_/ /  __/ /_/ / / / /
/____/_/ /_/ /_/\__/ .___/     / .___/_/\__, /\___/\____/_/ /_/
                  /_/         /_/      /____/
```

## `smtp-pigeon`

`smtp-pigeon` is a tiny SMTP server that accepts mail and delivers it as a
HTTP POST request. By default it uses a JSON formatted payload, but you can
provide any template you want.

If you have tools generating mail reports such as Debians unattended upgrades
but don't want to configure and run Postfix, SPF & DKIM & DMARC, etc or simply
want to funnel those reports into a *"please, anything but email"* service;
`smtp-pigeon` can act as the carrier.

- Configurable payload (via Go's template engine)
- Configurable headers
- Portable Go binary
- Tiny memory footprint

## Usage

Run `smtp-pigeon` with the `--url url` option. By default `smtp-pigeon` will
only accept connections from `127.0.0.1` at port `1025`.

See `--help` for other options.

```sh
smtp-pigeon --url https://my.endpoint.com/mail
# => smtp-pigeon listening at 127.0.0.1:1025
```

By default `smtp-pigeon` POSTs the following JSON:

```json
{
  "id": "per-message-uuid",
  "received_at": "ISO 8601 datetime when mail was received",
  "data": "multiline string\nof SMTP data",
  "from": "from@address",
  "to": ["to@address", "another@address"] 
}
```

You can configure what `smtp-pigeon` POSTs with the `--template` flag, using
any standard Go templating functions. You are not limited to sending JSON but
the `Content-Type` header is set to `application/json` by default. You must
override it with your own `--header` flag.

`smtp-pigeon` is intended to be run by systemd or any container runtime and
does not have a daemon form. These systems should be used to handle any
"service" requirements such as stop, start, restart and log aggregation.

Important caveats and gotchas:

- **Authentication**

  `smtp-pigeon` provides no authentication methods and will allow any mail
  client to connect and send mail.

  Do not run `smtp-pigeon` on a world accessible port without recognizing the
  consequences.

- **Message Content**

  `smtp-pigeon` performs no message screening. Whatever is sent in the mail will
  be passed along. This could include large amounts of encoded data if a service
  attaches a file for example. This may have repercussions downstream at the
  endpoint.

- **Delivery Guarantee**

  `smtp-pigeon` does not currently cache messages for re-delivery, if the POST
  fails for any reason (due to a crash, network error, endpoint failure, etc),
  the message will not be re-attempted. Do not rely on `smtp-pigeon` for
  business critical messaging.

- **MTAs**

  You will still need an MTA to deliver local mail *to* `smtp-pigeon`.

  An easy to use MTA is [sSMPT](https://wiki.debian.org/sSMTP) which provides a
  sendmail interface and runs daemon-less. Simply set `mailhub=localhost:1025`.

## Templating

You can specify a custom template using Go's
[text/template](https://pkg.go.dev/text/template) library. The following are
available in the template:

- `.ID`

  `string`

  UUID generated for each SMTP session, which effectively means for each
  message.

- `.Recipients`

  `list of strings`

  Each string will be provided by the SMTP `RCPT` command, this *may* be
  different to the `To` header. Will always be present, will always have at
  least one address. `To` header may or may not be given by the mail client.

- `.From`

  `string`

  Provided by the SMTP `FROM` command, this *may* be different to the `From`
  header. Will always be present, will always have one address. `From` header
  may or may not be given by the mail client.

- `.ReceivedAt`

  `time.Time`

  Marks the initial SMTP connection time, this *may* be different from the
  `Date` header. Will always be present. Will be in server timezone. You can
  convert to UTC via `{{.ReceivedAt.UTC().Format(time.RFC3399)}}`. `Date`
  header may or may not be given by the mail client.

- `.Data`

  `string`

  Provided by the SMTP `DATA` command. This contains the raw message data sent
  to the server, including headers, etc.

- `.Body`

  `string`

   `mail.Message.Body`, as parsed by
   [net/mail](https://pkg.go.dev/net/mail#ReadMessage). Converted to `string`
   from `io.Reader` for convenience.

- `.Header`

  `mail.Header`

   See [net/mail.Header](https://pkg.go.dev/net/mail#Header). You can safely
   access header values in your template with `{{.Header.Get "Subject"}}`,
   which will return `""` if the header does not exist.

## Testing the Server

You can manually inspect `smtp-pigeon`s behaviour by doing the following:

*Run the server. Consider generating a [ptsv2.com endpoint](https://ptsv2.com/)
to use or running a
[netcat](https://nmap.org/ncat/guide/index.html#ncat-overview) listen server.*

```sh
smtp-pigeon \
  --url https://my.endpoint.com/mail \
  --header "Authorization: Bearer my_auth_token" \
  --header "MyService-Node: $(hostname)" \
  --host 0.0.0.0 \
  --port 9925
# => smtp-pigeon listening at 0.0.0.0:9925
```

*Send mail via netcat.*

```sh
echo """EHLO localhost
MAIL FROM:<g.freeman@intern.blackmesa.com>
RCPT TO:<i.kleiner@materials.blackmesa.com>
RCPT TO:<e.vance@materials.blackmesa.com>
DATA
Subject: ON MY WAY
hey guys running l8 2day
on the tram now
cu soon
.""" | ncat localhost 9925

# => 220 localhost ESMTP Service Ready
# => 250-Hello localhost
# => 250-PIPELINING
# => 250-8BITMIME
# => 250-ENHANCEDSTATUSCODES
# => 250-CHUNKING
# => 250-AUTH PLAIN
# => 250 SIZE 1048576
# => 250 2.0.0 Roger, accepting mail from <gordon@intern.blackmesa.com>
# => 250 2.0.0 I'll make sure <kleiner@materials.blackmesa.com> gets this
# => 250 2.0.0 I'll make sure <vance@materials.blackmesa.com> gets this
# => 354 2.0.0 Go ahead. End your data with <CR><LF>.<CR><LF>
# => 250 2.0.0 OK: queued
```

*Server logs, note how a new session (and with it, a new message id) is
automatically created after delivery. Session reset and logout messages will
indicate whether a POST request was made or not.*

```text
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: New session
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Mail from: g.freeman@intern.blackmesa.com
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Rcpt to: i.kleiner@materials.blackmesa.com
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Rcpt to: e.vance@materials.blackmesa.com
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Data: [redacted (68 bytes)]
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Performing HTTP POST to http://my.endpoint.com/mail
121229b2-7676-4d0c-9a91-4431a245b067: HTTP POST responded with 200
99ed4e82-a5a6-4818-b5e2-69428d6be5e3: Session reset after POST
0a9f97ad-ca05-4df9-906e-6f0e885b906d: New session
0a9f97ad-ca05-4df9-906e-6f0e885b906d: Session logout without POST
```

## See also

- The [bokysan/docker-postfix
container](https://github.com/bokysan/docker-postfix) is a relatively painless
postfix solution.
- [emersion/go-smtp](https://github.com/emersion/go-smtp) provides the SMTP server implementation.
