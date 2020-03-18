# RelayBot

RelayBot relays messages between channels/rooms ('targets') on different chat platforms, currently supporting `Matrix` and `IRC`.

The bot can be configured to relay messages between targets in either a one-to-one or one-to-many relationship and is bi-directional.

## Configuration

A sample configuration file (`config.sample.yml`) is included within this repository

* `debug`: Specifies debug logging should be toggled on
* `servers`:
  * `irc`: Dictionary of configured IRC servers
    * `host`: Hostname/address for IRC server
    * `use_tls`: Indicates TLS should be used when connecting
    * `skip_tls_verify`: Indicates TLS verification should be disabled
    * `username`: Username for authenticating with IRC server
    * `password`: Password for authenticating with IRC server
    * `nick`: Nickname to set for bot
  * `matrix`: Dictionary of configured Matrix servers
    * `homeserver`: URI for Matrix server
    * `username`: Username for authenticating with Matrix server
    * `password`: Password for authenticating with Matrix server
    * `display_name`: Display name to set for bot
* `mappings`: Array of mapping objects
  * `from`:
    * `server`: Name of source server that mapping refers to
    * `name`: Name of source channel/room that mapping refers to
  * `to`:
    * `server`: Name of destination server that mapping refers to
    * `name`: Name of destination channel/room that mapping refers to