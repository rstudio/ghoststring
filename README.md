# ghoststring

[![Go Reference](https://pkg.go.dev/badge/github.com/rstudio/ghoststring.svg)](https://pkg.go.dev/github.com/rstudio/ghoststring)

A Go string wrapper type that is encrypted when JSONified. Encryption is symmetric using
AES-256-GMC with per-string nonce.

## usage

Typical usage assumes at least two systems that share a key. For example, two or more web
services may not directly communicate, but are passed data via state transfer through a
browser session to which they both need access without leaking to the browser.

Each instance of `GhostString` requires a non-empty namespace which must match a
registered `Ghostifyer`, e.g.:

Declare a type that uses a `GhostString` as a field:

```go
type Message struct {
	Recipient string                  `json:"recipient"`
	Content   ghoststring.GhostString `json:"content"`
	Mood      ghoststring.GhostString `json:"mood"`
}
```

Load a secret key from somewhere that meets your security requirements, such as from a
mounted kubernetes secret:

```go
secretKeyBytes, err := os.ReadFile("/path/to/secret-key")
if err != nil {
	return err
}
```

Register and set a `Ghostifyer` with the secret key and namespace in use:

```go
_, err := ghoststring.SetGhostifyer("heck.example.org", string(secretKeyBytes))
if err != nil {
	return err
}
```

Use the declared type as with any other struct type:

```go
msg := &Message{
	Recipient: "morningstar@heck.example.org",
	Content: ghoststring.GhostString{
		Namespace: "heck.example.org",
		String:    "We meet me at the fjord at dawn. Bring donuts, please.",
	},
	Mood: ghoststring.GhostString{
		Namespace: "heck.example.org",
		String:    "giddy",
	},
}
```

When the type instance is encoded to JSON, the `GhostString` fields will automatically
encode to a JSON string type:

```javascript
{
  "recipient": "morningstar@heck.example.org",
  "content": "👻:NTNjZjdmMWZjNzU2NTY1ODFkN2E4ZDI4aGVjay5leGFtcGxlLm9yZzo60XHqdSEbswpPbKLrDUuBZCrtUQLjmJ1NxHsqHMgBJjrB10O7JC4rwiMZw+wf/BOJGmu9ZCpMjgMpu18/VgDBpLn4n1nBNw==",
  "mood": "👻:YmYzZDVjNDVhNTUyMTA1Mzg0ZWU0NjI0aGVjay5leGFtcGxlLm9yZzo6SML6iqmwSoDhX3b2SQXsMLJPMHHS"
}
```

In the event of an unrecognized namespace, the `GhostString` fields will be encoded as
empty strings, e.g.:

```go
msg := &Message{
	Recipient: "frith@heck.example.org",
	Content: ghoststring.GhostString{
		Namespace: "heck.example.org",
		String:    "Next time I'm bringing the coffee.",
	},
	Mood: ghoststring.GhostString{
		Namespace: "wat.example.org",
		String:    "zzz",
	},
}
```

Because the `"wat.example.org"` namespace isn't registered, the output will look like
this:

```javascript
{
  "recipient": "frith@heck.example.org",
  "content": "👻:NTNjZjdmMWZjNzU2NTY1ODFkN2E4ZDI4aGVjay5leGFtcGxlLm9yZzo60XHqdSEbswpPbKLrDUuBZCrtUQLjmJ1NxHsqHMgBJjrB10O7JC4rwiMZw+wf/BOJGmu9ZCpMjgMpu18/VgDBpLn4n1nBNw==",
  "mood": ""
}
```

Any system that needs to read the encrypted contents must decode the JSON into a type that
uses `GhostString` for the matching fields in a process where a matching `Ghostifyer` has
been registered. Other systems may treat the values as opaque strings.

<!--
vim:tw=90
-->
