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
  "timestamp": "RFC 3399 datetime",
  "sender": "from@address",
  "recipients": ["to@address", "another@address"],
  "body": "multiline string",
  "subject": "string"
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
[text/template](https://pkg.go.dev/text/template) library.

*Hint: You can use more complex templates by passing `--template """$(cat
template.txt)"""` or similar, depending on your shell.*

The following are available in the template:

- `.ID`

  `string`

  UUID generated for each SMTP session, which effectively means for each
  message.

- `.Timestamp`

  `time.Time`

  Marks the initial SMTP connection time, this *may* be different from the
  `Date` header. Will always be present. Will be in server timezone. You can
  convert to UTC via `{{.Timestamp.UTC.Format "2006-01-02T15:04:05Z07:00" }}`.
  See [time Constants](https://pkg.go.dev/time#pkg-constants), note the string
  given indicates the format to use, not the time value. `Date` header may or
  may not be given by the mail client.

- `.Sender`

  `string`

  Provided by the SMTP `FROM` command, this *may* be different to the `From`
  header. Will always be present, will always have one address. `From` header
  may or may not be given by the mail client.

- `.Recipients`

  `list of strings`

  Each string will be provided by the SMTP `RCPT` command, this *may* be
  different to the `To` header. Will always be present, will always have at
  least one address. `To` header may or may not be given by the mail client.

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
MAIL FROM:<g.freeman@mailhub.bm.net>
RCPT TO:<i.kleiner@mailhub.bm.net>
RCPT TO:<e.vance@mailhub.bm.net>
DATA
Subject: ON MY WAY
From: Gordon Freeman <freeman@materials.blackmesa.com>
To: Eli Vance <vance@materials.blackmesa.com>
Cc: Issac Kleiner <kleiner@materials.blackmesa.com>

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
# => 250 2.0.0 Roger, accepting mail from <g.freeman@mailhub.bm.net>
# => 250 2.0.0 I'll make sure <i.kleiner@mailhub.bm.net> gets this
# => 250 2.0.0 I'll make sure <e.vance@mailhub.bm.net> gets this
# => 354 2.0.0 Go ahead. End your data with <CR><LF>.<CR><LF>
# => 250 2.0.0 OK: queued
```

*Server logs, note how a new session (and with it, a new message id) is
automatically created after delivery. Session reset and logout messages will
indicate whether a POST request was made or not.*

```text
c7f132cb-6043-4964-9853-02133a1e3181: New session
c7f132cb-6043-4964-9853-02133a1e3181: MAIL: g.freeman@mailhub.bm.net
c7f132cb-6043-4964-9853-02133a1e3181: RCPT: i.kleiner@mailhub.bm.net
c7f132cb-6043-4964-9853-02133a1e3181: RCPT: e.vance@mailhub.bm.net
c7f132cb-6043-4964-9853-02133a1e3181: DATA: [redacted (222 bytes)]
c7f132cb-6043-4964-9853-02133a1e3181: POST returned status: 200
c7f132cb-6043-4964-9853-02133a1e3181: Session reset after POST
02d6aebf-bfac-4c1f-9b7d-b9e781c7f1d1: New session
02d6aebf-bfac-4c1f-9b7d-b9e781c7f1d1: Session logout without POST
```

*Received payload:*

```json
{
  "id": "c7f132cb-6043-4964-9853-02133a1e3181",
  "timestamp": "2022-01-01T10:10:20Z",
  "sender": "g.freeman@mailhub.bm.net",
  "recipients": ["i.kleiner@mailhub.bm.net", "e.vance@mailhub.bm.net"],
  "body": "hey guys running l8 2day\non the tram now\ncu soon\n",
  "subject":"ON MY WAY"
}
```

## See also

- The [bokysan/docker-postfix
container](https://github.com/bokysan/docker-postfix) is a relatively painless
postfix solution.
- [emersion/go-smtp](https://github.com/emersion/go-smtp) provides the SMTP server implementation.
