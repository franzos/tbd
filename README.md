# TBD

SHUTDOWN: I started working on this project to build a better community platform - until I discovered Nostr, which is what TBD could / should have been. I'm now working on Nostr instead. Checkout my Nostr library and web client here: https://github.com/franzos/nostr-ts.

---

Tldr: Modern classifieds with trust and built-in actions like reservations, payments, shipping, etc. that optionally consume other classifieds.

Vue.js and React front-end as well as Typescript libraries: [tbd-client](https://github.com/franzos/tbd-client)/

After having played with Pocketbase recently, I felt motivated to use go for my next backend project. The name is yet to be determined, but I have a rough idea where this is heading.

*The state has made us dependent on itself; Time to depend on one another.*

## Quickstart

```bash
# guix environment --pure --ad-hoc go gcc-toolchain
CGO_ENABLED=1 go build .
```

Then simply run the binary.

```bash
./tbd
```

### Database

I currently default on SQLite for ease. 

- The database is created in the current working directory, and called `tbd.db`. 
- You can change the path using the `DB_PATH` environment variable.

The project relies on a ORM, which should make it easy to support Postgres, particularly for large communities. I plan to always support SQLite and Postgres. SQLite is just too easy, for small communities.

### Configuration

The project relies on a `.env` file in the current working directory. You can use the `example.env` file in the repo root as a starting point.

- At minimum you need to set the `JWT_SECRET`
- To upload files, set the AWS credentials and bucket name (local storage is planned)

```
JWT_SECRET=
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_BUCKET_NAME=
AWS_REGION=
DB_PATH=tbd.db
DOMAIN=
```

## Development

#### Hot reload

Hot reload with `air`:

```bash
# guix environment --pure --ad-hoc go gcc-toolchain
go install github.com/cosmtrek/air@latest
export GOBIN=/home/$(whoami)/go/bin
export PATH=$PATH:$GOBIN
```

Create a `.air.toml` from (`example.air.toml` in the repo root).

Run `air`:

```bash
air
```

## TODO

- [ ] CRUD for common operations
- [ ] Proper data validation - WIP
- [ ] Invalidate uploaded but never used files - WIP
- [ ] Frontend
- [ ] Docker image
- [ ] Support SQLite and Postgres
- [ ] Social login (Google, Facebook, Twitter, etc.)
- [ ] API docs
- [ ] Support for multiple communities
- [ ] Support files on local storage

## Tests

Note: The tests are generated using GPT with minor adjustments.

```
go test -v ./...
go test -v ./... -count=1
```

Run individual tests:

```
go test -v ./... -count=1 -run=TestPostEntryWithFilesAndUpdate
```

# Background

Rough outline:

- This is a community (Server)
  - Users can join this community, simply by signing up to post an entry
  - A user is free to follow and unfollow communities, to "shape" the experience
  - A user signs everything with their private key, and thus can be identified across communities (Communities will not have access to the private key)
  - There may be other communities, in other cities, states, countries, or network states
  - Communities are related to one another
    - for ex. you may picture a country's community, that recognizes a number of state communities
    - or a network community that represents a number of smaller communities
  - Communities provide trust, and thus one community may be more "trusted" than another (TBD)
  - Communities should be able to exist in different environments, under different laws, with different mechanisms to transfer value (for ex. USD (Fiat via Stripe for ex.), BTC, ETH, etc.)
- Every community has a market place that's made up of entries (Server)
  - This can be a product, a service, a job, a request, etc.
  - Entries can be booked, reserved, bought (product, miles delivery, ...), sold, etc., depending on type
  - Entries fulfillment may consume other entries, for ex. a product entry, may consume a service entry (delivery); for ex. the Delivery Service for your Pizza (so a Restaurant that sells Pizza, uses the service of another entry, which is a local delivery service which charges by the mile, to deliver the Pizza to your home)
  - Market places of communities with sub-communities, list all items
  - Members may freely post entries to their community's market place
  - Members may additionally post entries to other market places, as long as the community permits it
    - think of traveling to a larger market place to sell or source goods and services
  - This is what makes up the community economy, and incentives both community members as well as the community hosts

To express it in different terms, a larger community may be something like Alibaba, where vendors from a whole country come together, compared to going to a local market, where vendors gather with regional goods, in a more authentic atmosphere. On Alibaba you can pay with credit card and in any currency, and on the local market they may only take cash. Of course communities are more than buying and selling, but this is just the start.
  
For now I'll focus on getting one community up and running, with a market place, and a few entry types.

## Considerations

- Files: IPFS would be great but will add a lot of overhead on smaller nodes
- Signing: There will be two options: Having the key on the server (low security) or having the key on a mobile device (high security)
  - Low security is mostly for development but could still be useful in production to lower the bar of entry (user could migrate later)
  - Medium security: use user password? but if an attacker has access to the server, they can intercept the password
  - High security might involve an app that allows signing and posting directly (so the server doesn't need / store the key)
  - Maybe offer the option to store as file on user device instead (desktop, mobile, etc.) and sign in browser but that's too inconvenient
  - JSON keys?