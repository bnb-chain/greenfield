# Greenfield
Greenfield is a decentralized storage platform proposed to meet the requirements of storing non-financial information for 
decentralized applications integrated with BNB Smart Chain.

**Greenfield** is a blockchain built using Cosmos SDK and Tendermint and created with [Ignite CLI](https://ignite.com/cli).

## Get started

```
# To install the ignite binary in /usr/local/bin run the following command:
curl https://get.ignite.com/cli | bash
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
ignite chain serve
```

`serve` command installs dependencies, builds, initializes, and starts Greenfield in development.

### Configure

Greenfield in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

## Release
To release a new version of Greenfield, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.gg/ignite)
