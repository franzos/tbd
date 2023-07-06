# TBD

After having played with Pocketbase recently, I felt motivated to use go for my next backend project. The name is yet to be determined, but I have a rough idea where this is heading.

Tldr: Modern classifieds with trust and built-in actions like reservations, payments, shipping, etc. that optionally consume other classifieds.

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

## Build and run

```bash
# guix environment --pure --ad-hoc go gcc-toolchain
CGO_ENABLED=1 go build .
```

Then simply run the binary.

```bash
./tbd
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
- [ ] Invalidate uploaded but never used files
- [ ] Frontend
- [ ] Docker image
- [ ] Support SQLite and Postgres
- [ ] Social login (Google, Facebook, Twitter, etc.)
- [ ] API docs
- [ ] Support for multiple communities

## Tests

Note: The tests are generated using GPT with minor adjustments.

```
go test -v ./...
go test -v ./... -count=1
```